package agent

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/dagu-dev/dagu/internal/persistence"

	"github.com/dagu-dev/dagu/internal/dag"
	"github.com/dagu-dev/dagu/internal/engine"
	"github.com/dagu-dev/dagu/internal/mailer"
	"github.com/dagu-dev/dagu/internal/persistence/model"
	"github.com/dagu-dev/dagu/internal/reporter"
	"github.com/dagu-dev/dagu/internal/scheduler"
	"github.com/dagu-dev/dagu/internal/sock"
	"github.com/dagu-dev/dagu/internal/util"
	"github.com/google/uuid"
)

// Agent is responsible for running the DAG and handling communication
// via the unix socket. The agent performs the following tasks:
// 1. Start the DAG execution.
// 2. Propagate a signal to the running processes.
// 3. Handle the HTTP request via the unix socket.
// 4. Write the log and status to the data store.
type Agent struct {
	*Config

	dataStore    persistence.DataStoreFactory
	engine       engine.Engine
	scheduler    *scheduler.Scheduler
	graph        *scheduler.ExecutionGraph
	reporter     *reporter.Reporter
	historyStore persistence.HistoryStore
	socketServer *sock.Server
	logFile      *os.File

	// reqID is request ID to identify the DAG run.
	// The request ID can be used for history lookup, retry, etc.
	reqID    string
	finished atomic.Bool

	lock sync.RWMutex
}

// New creates a new Agent.
func New(config *Config, engine engine.Engine, dataStore persistence.DataStoreFactory) *Agent {
	return &Agent{Config: config, engine: engine, dataStore: dataStore}
}

// Config is the configuration for the Agent.
type Config struct {
	// DAG is the DAG to run.
	DAG *dag.DAG
	// Dry is a dry-run mode. It does not execute the actual command.
	Dry bool
	// RetryTarget is the status to retry.
	RetryTarget *model.Status
}

var (
	waitForRunning = time.Millisecond * 100
)

var (
	errFailedSetupUnixSocket = errors.New("failed to start the unix socket")
	errDAGIsAlreadyRunning   = errors.New("the DAG is already running")
)

// Run setups the scheduler and runs the DAG.
func (a *Agent) Run(ctx context.Context) error {
	if err := a.setup(); err != nil {
		return err
	}

	if err := a.checkPreconditions(); err != nil {
		return err
	}

	if a.Dry {
		return a.dryRun()
	}

	if err := a.checkIsAlreadyRunning(); err != nil {
		return err
	}

	if err := a.setupDatabase(); err != nil {
		return err
	}

	if err := a.setupSocketServer(); err != nil {
		return err
	}

	tw := newTeeWriter(a.logFile)
	if err := tw.Open(); err != nil {
		return err
	}

	defer func() {
		util.LogErr("close log file", a.closeLog())
		tw.Close()
	}()

	defer func() {
		if err := a.historyStore.Close(); err != nil {
			log.Printf("failed to close history store: %v", err)
		}
	}()

	util.LogErr("write status", a.historyStore.Write(a.Status()))

	listen := make(chan error)
	go func() {
		err := a.socketServer.Serve(listen)
		if err != nil && !errors.Is(err, sock.ErrServerRequestedShutdown) {
			log.Printf("failed to start socket frontend %v", err)
		}
	}()

	defer func() {
		util.LogErr("shutdown socket frontend", a.socketServer.Shutdown())
	}()

	if err := <-listen; err != nil {
		return errFailedSetupUnixSocket
	}

	done := make(chan *scheduler.Node)
	defer close(done)

	go func() {
		for node := range done {
			status := a.Status()
			util.LogErr("write status", a.historyStore.Write(status))
			util.LogErr("report step", a.reporter.ReportStep(a.DAG, status, node))
		}
	}()

	// Write the first status just after the start to store the running status.
	// If the DAG is already finished, skip it.
	go func() {
		time.Sleep(waitForRunning)
		if a.finished.Load() {
			return
		}
		util.LogErr("write status", a.historyStore.Write(a.Status()))
	}()

	// Start the DAG execution.
	lastErr := a.scheduler.Schedule(
		dag.NewContext(ctx, a.DAG, a.dataStore.NewDAGStore()),
		a.graph,
		done,
	)

	finishedStatus := a.Status()
	log.Println("schedule finished.")
	util.LogErr("write status", a.historyStore.Write(a.Status()))

	a.reporter.ReportSummary(finishedStatus, lastErr)
	util.LogErr("send email", a.reporter.SendMail(a.DAG, finishedStatus, lastErr))

	util.LogErr("close data file", a.historyStore.Close())
	a.finished.Store(true)

	return lastErr
}

// Status collects the current running status of the DAG and returns it.
func (a *Agent) Status() *model.Status {
	// Lock to avoid race condition.
	a.lock.RLock()
	defer a.lock.RUnlock()

	schedulerStatus := a.scheduler.Status(a.graph)
	if schedulerStatus == scheduler.StatusNone && a.graph.IsStarted() {
		// Match the status to the execution graph.
		schedulerStatus = scheduler.StatusRunning
	}

	// Create the status object to record the current status.
	status := &model.Status{
		RequestId:  a.reqID,
		Name:       a.DAG.Name,
		Status:     schedulerStatus,
		StatusText: schedulerStatus.String(),
		Pid:        model.Pid(os.Getpid()),
		Nodes:      model.FromNodesOrSteps(a.graph.NodeData(), a.DAG.Steps),
		StartedAt:  model.FormatTime(a.graph.StartAt()),
		FinishedAt: model.FormatTime(a.graph.FinishAt()),
		Log:        a.logFile.Name(),
		Params:     model.Params(a.DAG.Params),
	}

	// Collect the handler nodes.
	if node := a.scheduler.HandlerNode(dag.HandlerOnExit); node != nil {
		status.OnExit = model.FromNode(node.Data())
	}
	if node := a.scheduler.HandlerNode(dag.HandlerOnSuccess); node != nil {
		status.OnSuccess = model.FromNode(node.Data())
	}
	if node := a.scheduler.HandlerNode(dag.HandlerOnFailure); node != nil {
		status.OnFailure = model.FromNode(node.Data())
	}
	if node := a.scheduler.HandlerNode(dag.HandlerOnCancel); node != nil {
		status.OnCancel = model.FromNode(node.Data())
	}

	return status
}

// Signal sends the signal to the processes running
// if processes do not terminate after MaxCleanUp time, it will send KILL signal.
func (a *Agent) Signal(sig os.Signal) {
	a.signal(sig, false)
}

// Simple regular expressions for request routing
var (
	statusRe = regexp.MustCompile(`^/status[/]?$`)
	stopRe   = regexp.MustCompile(`^/stop[/]?$`)
)

// HandleHTTP handles HTTP requests via unix socket.
func (a *Agent) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	switch {
	case r.Method == http.MethodGet && statusRe.MatchString(r.URL.Path):
		// Return the current status of the execution.
		status := a.Status()
		status.Status = scheduler.StatusRunning
		b, err := status.ToJson()
		if err != nil {
			encodeError(w, err)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(b)
	case r.Method == http.MethodPost && stopRe.MatchString(r.URL.Path):
		// Handle Stop request for the DAG execution.
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
		go func() {
			log.Printf("stop request received. shutting down...")
			a.signal(syscall.SIGTERM, true)
		}()
	default:
		// Unknown request
		encodeError(w, &httpError{Code: http.StatusNotFound, Message: "Not found"})
	}
}

func (a *Agent) setup() error {
	// Lock to prevent race condition.
	a.lock.Lock()
	defer a.lock.Unlock()

	if err := a.setupReqID(); err != nil {
		return err
	}

	if err := a.setupLog(); err != nil {
		return err
	}

	// Setup the scheduler for the DAG.
	cfg := &scheduler.Config{
		LogDir:        a.DAG.GetLogDir(),
		MaxActiveRuns: a.DAG.MaxActiveRuns,
		Delay:         a.DAG.Delay,
		Dry:           a.Dry,
		ReqID:         a.reqID,
	}

	if a.DAG.HandlerOn.Exit != nil {
		cfg.OnExit = a.DAG.HandlerOn.Exit
	}

	if a.DAG.HandlerOn.Success != nil {
		cfg.OnSuccess = a.DAG.HandlerOn.Success
	}

	if a.DAG.HandlerOn.Failure != nil {
		cfg.OnFailure = a.DAG.HandlerOn.Failure
	}

	if a.DAG.HandlerOn.Cancel != nil {
		cfg.OnCancel = a.DAG.HandlerOn.Cancel
	}

	a.scheduler = &scheduler.Scheduler{Config: cfg}
	a.reporter = &reporter.Reporter{
		Config: &reporter.Config{
			Mailer: &mailer.Mailer{
				Config: &mailer.Config{
					Host:     a.DAG.Smtp.Host,
					Port:     a.DAG.Smtp.Port,
					Username: a.DAG.Smtp.Username,
					Password: a.DAG.Smtp.Password,
				},
			},
		}}

	if err := a.setupGraph(); err != nil {
		return err
	}

	return nil
}

// dryRun performs a dry-run of the DAG.
// It only simulates the execution of the DAG without running the actual command.
func (a *Agent) dryRun() error {
	// done channel receives the node when the node is done.
	// It's a way to update the status in real-time in efficient manner.
	done := make(chan *scheduler.Node)
	defer func() {
		close(done)
	}()

	go func() {
		for node := range done {
			status := a.Status()
			_ = a.reporter.ReportStep(a.DAG, status, node)
		}
	}()

	log.Printf("***** Starting DRY-RUN *****")

	lastErr := a.scheduler.Schedule(
		dag.NewContext(context.Background(), a.DAG, a.dataStore.NewDAGStore()),
		a.graph,
		done,
	)

	a.reporter.ReportSummary(a.Status(), lastErr)

	log.Printf("***** Finished DRY-RUN *****")

	return lastErr
}

func (a *Agent) signal(sig os.Signal, allowOverride bool) {
	log.Printf("Sending %s signal to running child processes.", sig)
	done := make(chan bool)
	go func() {
		a.scheduler.Signal(a.graph, sig, done, allowOverride)
	}()
	timeout := time.NewTimer(a.DAG.MaxCleanUpTime)
	tick := time.NewTimer(time.Second * 5)
	defer timeout.Stop()
	defer tick.Stop()

	for {
		select {
		case <-done:
			log.Printf("All child processes have been terminated.")
			return
		case <-timeout.C:
			log.Printf("Time reached to max cleanup time")
			log.Printf("Sending KILL signal to running child processes.")
			a.scheduler.Signal(a.graph, syscall.SIGKILL, nil, false)
			return
		case <-tick.C:
			log.Printf("Sending signal again")
			a.scheduler.Signal(a.graph, sig, nil, false)
			tick.Reset(time.Second * 5)
		default:
			log.Printf("Waiting for child processes to exit...")
			time.Sleep(time.Second * 3)
		}
	}
}

func (a *Agent) setupGraph() (err error) {
	if a.RetryTarget != nil {
		log.Printf("setup for retry")
		return a.setupRetry()
	}
	a.graph, err = scheduler.NewExecutionGraph(a.DAG.Steps...)
	return
}

func (a *Agent) setupRetry() (err error) {
	nodes := make([]*scheduler.Node, 0, len(a.RetryTarget.Nodes))
	for _, n := range a.RetryTarget.Nodes {
		nodes = append(nodes, n.ToNode())
	}
	a.graph, err = scheduler.NewExecutionGraphForRetry(nodes...)
	return
}

// setupReqID generates a new request ID.
func (a *Agent) setupReqID() error {
	id, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	a.reqID = id.String()

	return nil
}

func (a *Agent) setupDatabase() error {
	a.historyStore = a.dataStore.NewHistoryStore()
	if err := a.historyStore.RemoveOld(a.DAG.Location, a.DAG.HistRetentionDays); err != nil {
		util.LogErr("clean old history data", err)
	}

	return a.historyStore.Open(a.DAG.Location, time.Now(), a.reqID)
}

func (a *Agent) setupSocketServer() (err error) {
	a.socketServer, err = sock.NewServer(&sock.Config{
		Addr:        a.DAG.SockAddr(),
		HandlerFunc: a.HandleHTTP,
	})
	return
}

func (a *Agent) checkPreconditions() error {
	if len(a.DAG.Preconditions) > 0 {
		log.Printf("checking preconditions for \"%s\"", a.DAG.Name)
		if err := dag.EvalConditions(a.DAG.Preconditions); err != nil {
			a.scheduler.Cancel(a.graph)
			return err
		}
	}
	return nil
}

func (a *Agent) checkIsAlreadyRunning() error {
	status, err := a.engine.GetCurrentStatus(a.DAG)
	if err != nil {
		return err
	}
	if status.Status != scheduler.StatusNone {
		return fmt.Errorf("%w. socket=%s", errDAGIsAlreadyRunning, a.DAG.SockAddr())
	}
	return nil
}

const (
	logFileTimeStampFmt = "20060102.15:04:05.000"
	reqIDLenSafe        = 8
)

func (a *Agent) setupLog() error {
	fileName := fmt.Sprintf("agent_%s.%s.%s.log",
		util.ValidFilename(a.DAG.Name),
		time.Now().Format(logFileTimeStampFmt),
		util.TruncString(a.reqID, reqIDLenSafe),
	)
	absFilepath := filepath.Join(a.DAG.GetLogDir(), fileName)

	// Create the log directory
	if err := os.MkdirAll(path.Dir(absFilepath), 0755); err != nil {
		return err
	}

	file, err := util.OpenOrCreateFile(absFilepath)
	if err != nil {
		return err
	}
	a.logFile = file
	return nil
}

func (a *Agent) closeLog() error {
	if a.logFile != nil {
		return a.logFile.Close()
	}
	return nil
}

type httpError struct {
	Code    int
	Message string
}

// Error implements error interface.
func (e *httpError) Error() string { return e.Message }

// encodeError returns error to the HTTP client.
func encodeError(w http.ResponseWriter, err error) {
	var httpErr *httpError
	if errors.As(err, &httpErr) {
		http.Error(w, httpErr.Error(), httpErr.Code)
	} else {
		http.Error(w, httpErr.Error(), http.StatusInternalServerError)
	}
}

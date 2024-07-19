package scheduler

import (
	"errors"
	"time"

	"github.com/dagu-dev/dagu/internal/dag"
	dagscheduler "github.com/dagu-dev/dagu/internal/dag/scheduler"
	"github.com/dagu-dev/dagu/internal/engine"
	"github.com/dagu-dev/dagu/internal/util"
)

var (
	errJobRunning      = errors.New("job already running")
	errJobIsNotRunning = errors.New("job is not running")
	errJobFinished     = errors.New("job already finished")
)

var _ jobCreator = (*jobCreatorImpl)(nil)

type jobCreatorImpl struct {
	Executable string
	WorkDir    string
	Engine     engine.Engine
}

func (jf jobCreatorImpl) CreateJob(dg *dag.DAG, next time.Time) job {
	return &jobImpl{
		DAG:        dg,
		Executable: jf.Executable,
		WorkDir:    jf.WorkDir,
		Next:       next,
		Engine:     jf.Engine,
	}
}

var _ job = (*jobImpl)(nil)

type jobImpl struct {
	DAG        *dag.DAG
	Executable string
	WorkDir    string
	Next       time.Time
	Engine     engine.Engine
}

func (j *jobImpl) GetDAG() *dag.DAG {
	return j.DAG
}

func (j *jobImpl) Start() error {
	latestStatus, err := j.Engine.GetLatestStatus(j.DAG)
	if err != nil {
		return err
	}

	if latestStatus.Status == dagscheduler.StatusRunning {
		// already running
		return errJobRunning
	}

	// check the last execution time
	lastExecTime, err := util.ParseTime(latestStatus.StartedAt)
	if err == nil {
		lastExecTime = lastExecTime.Truncate(time.Second * 60)
		if lastExecTime.After(j.Next) || j.Next.Equal(lastExecTime) {
			return errJobFinished
		}
	}
	return j.Engine.Start(j.DAG, engine.StartOptions{
		Quiet: true,
	})
}

func (j *jobImpl) Stop() error {
	latestStatus, err := j.Engine.GetLatestStatus(j.DAG)
	if err != nil {
		return err
	}
	if latestStatus.Status != dagscheduler.StatusRunning {
		return errJobIsNotRunning
	}
	return j.Engine.Stop(j.DAG)
}

func (j *jobImpl) Restart() error {
	return j.Engine.Restart(j.DAG, engine.RestartOptions{
		Quiet: true,
	})
}

func (j *jobImpl) String() string {
	return j.DAG.Name
}

package cmd

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/dagu-dev/dagu/internal/agent"
	"github.com/dagu-dev/dagu/internal/config"
	"github.com/dagu-dev/dagu/internal/dag"
	"github.com/dagu-dev/dagu/internal/dag/scheduler"
	"github.com/dagu-dev/dagu/internal/engine"
	"github.com/dagu-dev/dagu/internal/logger"
	"github.com/spf13/cobra"
)

func restartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restart <DAG file>",
		Short: "Stop the running DAG and restart it",
		Long:  `dagu restart <DAG file>`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.Load()
			if err != nil {
				log.Fatalf("Failed to load config: %v", err)
			}

			quiet, err := cmd.Flags().GetBool("quiet")
			if err != nil {
				log.Fatalf("Failed to get quiet flag: %v", err)
			}

			initLogger := logger.NewLogger(logger.NewLoggerArgs{
				Config: cfg,
				Quiet:  quiet,
			})

			// Load the DAG file and stop the DAG if it is running.
			dagFilePath := args[0]
			dg, err := dag.Load(cfg.BaseConfig, dagFilePath, "")
			if err != nil {
				initLogger.Error("Failed to load DAG", "error", err)
				os.Exit(1)
			}

			eng := newEngine(cfg)

			if err := stopDAGIfRunning(eng, dg, initLogger); err != nil {
				initLogger.Error("Failed to stop the DAG", "error", err)
				os.Exit(1)
			}

			// Wait for the specified amount of time before restarting.
			waitForRestart(dg.RestartWait, initLogger)

			// Retrieve the parameter of the previous execution.
			params, err := getPreviousExecutionParams(eng, dg)
			if err != nil {
				initLogger.Error("Failed to get previous execution params", "error", err)
				os.Exit(1)
			}

			// Start the DAG with the same parameter.
			// Need to reload the DAG file with the parameter.
			dg, err = dag.Load(cfg.BaseConfig, dagFilePath, params)
			if err != nil {
				initLogger.Error("Failed to load DAG", "error", err)
				os.Exit(1)
			}

			requestID, err := generateRequestID()
			if err != nil {
				initLogger.Error("Failed to generate request ID", "error", err)
				os.Exit(1)
			}

			logFile, err := openLogFileForDAG("restart_", cfg.LogDir, dg, requestID)
			if err != nil {
				initLogger.Error("Failed to open log file for DAG", "error", err)
				os.Exit(1)
			}
			defer logFile.Close()

			agentLogger := logger.NewLogger(logger.NewLoggerArgs{
				Config:  cfg,
				LogFile: logFile,
				Quiet:   quiet,
			})

			agentLogger.Info("Restarting DAG", "dag", dg.Name)

			dagAgent := agent.New(
				requestID,
				dg,
				agentLogger,
				filepath.Dir(logFile.Name()),
				logFile.Name(),
				eng,
				newDataStores(cfg),
				&agent.AgentOpts{Dry: false})

			listenSignals(cmd.Context(), dagAgent)
			if err := dagAgent.Run(cmd.Context()); err != nil {
				agentLogger.Error("Failed to start DAG", "error", err)
				os.Exit(1)
			}
		},
	}
	cmd.Flags().BoolP("quiet", "q", false, "suppress output")
	return cmd
}

// stopDAGIfRunning stops the DAG if it is running.
// Otherwise, it does nothing.
func stopDAGIfRunning(e engine.Engine, dg *dag.DAG, lg logger.Logger) error {
	curStatus, err := e.GetCurrentStatus(dg)
	if err != nil {
		return err
	}

	if curStatus.Status == scheduler.StatusRunning {
		lg.Info("Stopping DAG for restart", "dag", dg.Name)
		cobra.CheckErr(stopRunningDAG(e, dg))
	}
	return nil
}

// stopRunningDAG attempts to stop the running DAG
// by sending a stop signal to the agent.
func stopRunningDAG(e engine.Engine, dg *dag.DAG) error {
	for {
		curStatus, err := e.GetCurrentStatus(dg)
		if err != nil {
			return err
		}

		// If the DAG is not running, do nothing.
		if curStatus.Status != scheduler.StatusRunning {
			return nil
		}

		if err := e.Stop(dg); err != nil {
			return err
		}

		time.Sleep(time.Millisecond * 100)
	}
}

// waitForRestart waits for the specified amount of time before restarting
// the DAG.
func waitForRestart(restartWait time.Duration, lg logger.Logger) {
	if restartWait > 0 {
		lg.Info("Waiting for restart", "duration", restartWait)
		time.Sleep(restartWait)
	}
}

func getPreviousExecutionParams(e engine.Engine, dg *dag.DAG) (string, error) {
	latestStatus, err := e.GetLatestStatus(dg)
	if err != nil {
		return "", err
	}

	return latestStatus.Params, nil
}

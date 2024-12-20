// Copyright (C) 2024 Yota Hamada
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"fmt"

	"github.com/dagu-org/dagu/internal/config"
	"github.com/dagu-org/dagu/internal/frontend"
	"github.com/dagu-org/dagu/internal/scheduler"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func startAllCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "start-all",
		Short:   "Launches both the Dagu web UI server and the scheduler process.",
		Long:    `dagu start-all [--dags=<DAGs dir>] [--host=<host>] [--port=<port>]`,
		PreRunE: bindStartAllFlags,
		RunE:    runStartAll,
	}

	initStartAllFlags(cmd)
	return cmd
}

func bindStartAllFlags(cmd *cobra.Command, _ []string) error {
	flags := []string{"port", "host", "dags"}
	for _, flag := range flags {
		if err := viper.BindPFlag(flag, cmd.Flags().Lookup(flag)); err != nil {
			return fmt.Errorf("failed to bind flag %s: %w", flag, err)
		}
	}
	return nil
}

func runStartAll(cmd *cobra.Command, _ []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	logger := buildLogger(cfg, false)

	// Update DAGs directory if specified
	if dagsDir, _ := cmd.Flags().GetString("dags"); dagsDir != "" {
		cfg.DAGs = dagsDir
	}

	ctx := cmd.Context()
	dataStore := newDataStores(cfg)
	cli := newClient(cfg, dataStore, logger)

	// Start scheduler in a goroutine
	errChan := make(chan error, 1)
	go func() {
		logger.Info("Scheduler initialization", "dags", cfg.DAGs)

		sc := scheduler.New(ctx, cfg, logger, cli)
		if err := sc.Start(ctx); err != nil {
			errChan <- fmt.Errorf("scheduler initialization failed: %w", err)
			return
		}
		errChan <- nil
	}()

	// Start server in main thread
	logger.Info("Server initialization", "host", cfg.Host, "port", cfg.Port)

	server := frontend.New(cfg, logger, cli)
	serverErr := make(chan error, 1)
	go func() {
		if err := server.Serve(ctx); err != nil {
			serverErr <- fmt.Errorf("server initialization failed: %w", err)
			return
		}
		serverErr <- nil
	}()

	// Wait for either error to occur
	select {
	case err := <-errChan:
		if err != nil {
			return err
		}
	case err := <-serverErr:
		if err != nil {
			return err
		}
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

func initStartAllFlags(cmd *cobra.Command) {
	flags := []struct {
		name, shorthand, defaultValue, usage string
	}{
		{
			name:      "dags",
			shorthand: "d",
			usage:     "location of DAG files (default is $HOME/.config/dagu/dags)",
		},
		{
			name:         "host",
			shorthand:    "s",
			defaultValue: defaultHost,
			usage:        "server host",
		},
		{
			name:         "port",
			shorthand:    "p",
			defaultValue: defaultPort,
			usage:        "server port",
		},
	}

	for _, flag := range flags {
		cmd.Flags().StringP(
			flag.name,
			flag.shorthand,
			flag.defaultValue,
			flag.usage,
		)
	}
}

// Copyright (C) 2024 Yota Hamada
// SPDX-License-Identifier: GPL-3.0-or-later

package scheduler

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/dagu-org/dagu/internal/client"
	"github.com/dagu-org/dagu/internal/fileutil"
	"github.com/dagu-org/dagu/internal/logger"
	"github.com/dagu-org/dagu/internal/scheduler/filenotify"
	"github.com/robfig/cron/v3"

	"github.com/dagu-org/dagu/internal/digraph"
	"github.com/fsnotify/fsnotify"
)

var _ entryReader = (*entryReaderImpl)(nil)

type entryReaderImpl struct {
	dagsDir    string
	dagsLock   sync.Mutex
	dags       map[string]*digraph.DAG
	jobCreator jobCreator
	logger     logger.Logger
	client     client.Client
}

type jobCreator interface {
	CreateJob(dag *digraph.DAG, next time.Time, schedule cron.Schedule) job
}

func newEntryReader(ctx context.Context, dagsDir string, jobCreator jobCreator, logger logger.Logger, client client.Client) *entryReaderImpl {
	er := &entryReaderImpl{
		dagsDir:    dagsDir,
		dagsLock:   sync.Mutex{},
		dags:       map[string]*digraph.DAG{},
		jobCreator: jobCreator,
		logger:     logger,
		client:     client,
	}
	if err := er.initDAGs(ctx); err != nil {
		er.logger.Error("DAG initialization failed", "error", err)
	}
	return er
}

func (er *entryReaderImpl) Start(ctx context.Context, done chan any) {
	go er.watchDags(ctx, done)
}

func (er *entryReaderImpl) Read(ctx context.Context, now time.Time) ([]*entry, error) {
	er.dagsLock.Lock()
	defer er.dagsLock.Unlock()

	var entries []*entry
	addEntriesFn := func(dag *digraph.DAG, schedules []digraph.Schedule, entryType entryType) {
		for _, schedule := range schedules {
			next := schedule.Parsed.Next(now)
			entries = append(entries, &entry{
				Next:      schedule.Parsed.Next(now),
				Job:       er.jobCreator.CreateJob(dag, next, schedule.Parsed),
				EntryType: entryType,
				Logger:    er.logger,
			})
		}
	}

	for _, dag := range er.dags {
		id := strings.TrimSuffix(
			filepath.Base(dag.Location),
			filepath.Ext(dag.Location),
		)

		if er.client.IsSuspended(ctx, id) {
			continue
		}
		addEntriesFn(dag, dag.Schedule, entryTypeStart)
		addEntriesFn(dag, dag.StopSchedule, entryTypeStop)
		addEntriesFn(dag, dag.RestartSchedule, entryTypeRestart)
	}

	return entries, nil
}

func (er *entryReaderImpl) initDAGs(ctx context.Context) error {
	er.dagsLock.Lock()
	defer er.dagsLock.Unlock()

	fis, err := os.ReadDir(er.dagsDir)
	if err != nil {
		return err
	}

	var fileNames []string
	for _, fi := range fis {
		if fileutil.IsYAMLFile(fi.Name()) {
			dag, err := digraph.LoadMetadata(ctx, filepath.Join(er.dagsDir, fi.Name()))
			if err != nil {
				er.logger.Error(
					"DAG load failed",
					"error", err,
					"DAG", fi.Name(),
				)
				continue
			}
			er.dags[fi.Name()] = dag
			fileNames = append(fileNames, fi.Name())
		}
	}

	er.logger.Info("Scheduler initialized", "specs", strings.Join(fileNames, ","))
	return nil
}

func (er *entryReaderImpl) watchDags(ctx context.Context, done chan any) {
	watcher, err := filenotify.New(time.Minute)
	if err != nil {
		er.logger.Error("Watcher creation failed", "error", err)
		return
	}

	defer func() {
		_ = watcher.Close()
	}()
	_ = watcher.Add(er.dagsDir)

	for {
		select {
		case <-done:
			return
		case event, ok := <-watcher.Events():
			if !ok {
				return
			}
			if !fileutil.IsYAMLFile(event.Name) {
				continue
			}
			er.dagsLock.Lock()
			if event.Op == fsnotify.Create || event.Op == fsnotify.Write {
				dag, err := digraph.LoadMetadata(ctx, filepath.Join(er.dagsDir, filepath.Base(event.Name)))
				if err != nil {
					er.logger.Error(
						"DAG load failed",
						"error",
						err,
						"file",
						event.Name,
					)
				} else {
					er.dags[filepath.Base(event.Name)] = dag
					er.logger.Info("DAG added/updated", "DAG", filepath.Base(event.Name))
				}
			}
			if event.Op == fsnotify.Rename || event.Op == fsnotify.Remove {
				delete(er.dags, filepath.Base(event.Name))
				er.logger.Info("DAG removed", "DAG", filepath.Base(event.Name))
			}
			er.dagsLock.Unlock()
		case err, ok := <-watcher.Errors():
			if !ok {
				return
			}
			er.logger.Error("Watcher error", "error", err)
		}
	}

}

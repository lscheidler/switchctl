/*
  Copyright 2020 Lars Eric Scheidler

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.
*/
package progress

import (
	"os"
	"sync"
	"time"

	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
	"go.uber.org/zap"

	"github.com/lscheidler/switchctl/cli"
	"github.com/lscheidler/switchctl/common"
	"github.com/lscheidler/switchctl/conf"
)

type Progress struct {
	slog                      *zap.SugaredLogger
	FailedApplications        []*common.Application
	SuccessfulApplications    []*common.Application
	colorizeInstanceCompleted func(string, bool) string
	workers                   int
}

func New(slog *zap.SugaredLogger, workers int, colorizeInstanceCompleted func(string, bool) string) *Progress {
	return &Progress{
		slog:                      slog,
		colorizeInstanceCompleted: colorizeInstanceCompleted,
		workers:                   workers,
	}
}

func (progress *Progress) Load(args *cli.Arguments, config *conf.Config) {
	var wg sync.WaitGroup
	var successMutex sync.Mutex
	var failMutex sync.Mutex

	p := mpb.New(
		mpb.WithWaitGroup(&wg),
		mpb.WithWidth(1),
	)

	wg.Add(len(args.Applications))

	var bar *mpb.Bar
	bar = p.AddSpinner(int64(len(args.Applications)), mpb.SpinnerOnMiddle,
		mpb.PrependDecorators(
			decor.Name("Loading version information", decor.WCSyncSpaceR),
		),
		mpb.BarRemoveOnComplete(),
	)

	wp := NewWorkerPool(progress.workers)
	for _, application := range []*common.Application(args.Applications) {
		progress.slog.Debug("Loading application ", application.Name)

		//go progress.loadApplication(&wg, bar, application, config, args, &successMutex, &failMutex)
		go progress.loadApplication(wp, &wg, bar, application, config, args, &successMutex, &failMutex)
	}
	p.Wait()
	wp.Close()
}

func (progress *Progress) loadApplication(wp *WorkerPool, wg *sync.WaitGroup, bar *mpb.Bar, application *common.Application, config *conf.Config, args *cli.Arguments, successMutex *sync.Mutex, failMutex *sync.Mutex) {
	defer wg.Done()
	defer wp.Done()
	wp.Add()

	start := time.Now()
	if err := application.Load(progress.slog, config, args.Environment, args.Dryrun); err != nil {
		failMutex.Lock()
		progress.FailedApplications = append(progress.FailedApplications, application)
		failMutex.Unlock()
	} else {
		successMutex.Lock()
		progress.SuccessfulApplications = append(progress.SuccessfulApplications, application)
		successMutex.Unlock()
	}
	bar.IncrBy(1, time.Since(start))
	progress.slog.Debug("Loaded application ", application.Name)
}

func (progress *Progress) SwitchApplications() {
	var doneWg sync.WaitGroup
	p := mpb.New(mpb.WithWidth(1), mpb.WithWaitGroup(&doneWg))
	wp := NewWorkerPool(progress.workers)

	failed := []*bool{}

	var bars []*mpb.Bar
	var switchWgg []*sync.WaitGroup
	for _, application := range progress.SuccessfulApplications {
		progress.slog.Info("Switching application ", application.Name, "=", application.Version)

		var wg sync.WaitGroup
		applicationFailed := false
		failed = append(failed, &applicationFailed)

		instanceChan := make(chan int)
		wg.Add(1)
		switchWgg = append(switchWgg, &wg)

		var b *mpb.Bar
		b = p.AddSpinner(
			int64(len(application.SuccessfulInstances)),
			mpb.SpinnerOnMiddle,
			mpb.BarClearOnComplete(),
			mpb.PrependDecorators(
				decor.Name(application.Name, decor.WCSyncSpaceR),
				OnStepFunction(application.InstanceCompleted(progress.colorizeInstanceCompleted), decor.WCSyncSpaceR),
			),
			mpb.AppendDecorators(
				decor.OnComplete(OnCompleteFailed(&applicationFailed, "finished with errors"), "done!"),
			),
		)
		bars = append(bars, b)

		go progress.switchApplication(&applicationFailed, wp, &wg, &instanceChan, b, application)
	}
	exitCode := 0
	for i, _ := range progress.SuccessfulApplications {
		switchWgg[i].Wait()

		if *failed[i] {
			exitCode = 1
		}
	}
	p.Wait()
	wp.Close()

	os.Exit(exitCode)
}

func (progress *Progress) switchApplication(failed *bool, wp *WorkerPool, wg *sync.WaitGroup, instanceChan *chan int, bar *mpb.Bar, application *common.Application) {
	defer wg.Done()
	defer wp.Done()
	wp.Add()

	start := time.Now()
	for _, instance := range application.SuccessfulInstances {

		if instance.Connected() && len(instance.Errors) == 0 {
			if command := instance.Switch(application.Name, application.Version); command.Error != nil {
				progress.slog.Warnf("%s[%s] failed: %s, %s", application.Name, instance.Hostname(), command.Description, command.Error)
				progress.slog.Warnf("%s[%s] output: %s", application.Name, instance.Hostname(), command.Combined)
				bar.IncrBy(1, time.Since(start))
				*failed = true
				return
			} else {
				progress.slog.Debugf("%s[%s] output: %s", application.Name, instance.Hostname(), command.Combined)
			}
			bar.IncrBy(1, time.Since(start))
		}
	}
	progress.slog.Debug("Switched application ", application.Name)
}

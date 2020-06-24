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
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/agtorre/gocolorize"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/lscheidler/switchctl/cli"
	"github.com/lscheidler/switchctl/conf"
	"github.com/lscheidler/switchctl/progress"
)

var (
	slog *zap.SugaredLogger
)

func main() {
	args := cli.ParseArguments()
	config := conf.LoadConfig()

	openLog(args)
	defer slog.Sync()

	cred := gocolorize.Colorize{Fg: gocolorize.Red}

	defer args.Applications.Close()
	p := progress.New(slog, colorizeInstanceCompleted)
	p.Load(args, config)

	printApplicationInformation(p)

	if len(p.SuccessfulApplications) > 0 {
		fmt.Println(cred.Paint("please enter 'ok' to proceed (<control>+c or <enter> for exit):"))
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')

		if text == "ok\n" {
			p.SwitchApplications()
			for _, application := range p.SuccessfulApplications {
				for _, instance := range application.SuccessfulInstances {
					slog.Debugf("%#v", instance.Commands)
				}
			}
		}
	} else {
		fmt.Println("All applications failed.")
		os.Exit(1)
	}
}

func openLog(args *cli.Arguments) {
	os.Mkdir("logs/", 0755)

	config := zap.NewProductionConfig()
	config.EncoderConfig = zap.NewProductionEncoderConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.Encoding = "console"
	if args.Debug {
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}
	config.OutputPaths = []string{"logs/switchctl.log"}

	logOption := zap.AddCaller()
	logger, err := config.Build(logOption)
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	slog = logger.Sugar()
}

func printApplicationInformation(p *progress.Progress) {
	cyellow := gocolorize.Colorize{Fg: gocolorize.Yellow}
	cred := gocolorize.Colorize{Fg: gocolorize.Red}

	if len(p.SuccessfulApplications) > 0 {
		fmt.Println("Going to switch following applications:")
		fmt.Println()

		for _, application := range p.SuccessfulApplications {
			fmt.Printf("  - name:       %s\n    version:    %s\n", cyellow.Paint(application.Name), cyellow.Paint(application.Version))

			for _, instance := range application.SuccessfulInstances {
				fmt.Printf("    - hostname: %s\n", cyellow.Paint(instance.Hostname()))
				fmt.Printf("      current:  %s\n", instance.CurrentVersion().String())
			}
			for _, instance := range application.FailedInstances {
				fmt.Printf("    - hostname: %s (skipping...)\n      current:  %s\n      errors:   %v\n", cred.Paint(instance.Hostname()), instance.CurrentVersion().String(), instance.Errors)
			}
			fmt.Println()
		}
	}

	if len(p.FailedApplications) > 0 {
		fmt.Println("Following applications are going to be skipped:")
		fmt.Println()

		for _, application := range p.FailedApplications {
			slog.Warnf("Skipping application %s because of errors (%v)", application.Name, application.Errors)

			fmt.Printf("  - name:       %s\n    version:    %s\n", cred.Paint(application.Name), cred.Paint(application.Version))

			for _, instance := range application.FailedInstances {
				fmt.Printf("    - hostname: %s\n      current:  %s\n      errors:   %v\n", cred.Paint(instance.Hostname()), instance.CurrentVersion().String(), instance.Errors)
			}
			for _, instance := range application.SuccessfulInstances {
				fmt.Printf("    - hostname: %s %s\n", cyellow.Paint(instance.Hostname()), instance.CurrentVersion().String())
			}
			fmt.Println()
		}
	}
}

func colorizeInstanceCompleted(name string, failed bool) string {
	if failed {
		cred := gocolorize.Colorize{Fg: gocolorize.Red}

		return cred.Paint(name)
	} else {
		cgreen := gocolorize.Colorize{Fg: gocolorize.Green}

		return cgreen.Paint(name)
	}
}

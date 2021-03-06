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
package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/lscheidler/switchctl/common"
)

const (
	version = "0.4"

	applicationUsage   = "set application to switch"
	debugUsage         = "debug mode"
	debugDefault       = false
	dryrunDefault      = false
	dryrunUsage        = "do not execute switch"
	environmentDefault = "production"
	environmentUsage   = "set environment to use"
	logfileDefault     = "logs/switchctl.log"
	logfileUsage       = "logfile path"
	workersDefault     = 5
	workersUsage       = "number of workers run simultaneously"
)

type Arguments struct {
	Applications common.Applications
	Debug        bool
	Dryrun       bool
	Environment  string
	Logfile      string
	Workers      int
}

func ParseArguments() *Arguments {
	args := Arguments{}

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s (%s):\n", os.Args[0], version)
		flag.PrintDefaults()
	}

	flag.Var(&args.Applications, "application", applicationUsage)
	flag.Var(&args.Applications, "a", applicationUsage)
	flag.StringVar(&args.Environment, "environment", environmentDefault, environmentUsage)
	flag.StringVar(&args.Environment, "e", environmentDefault, environmentUsage)
	flag.BoolVar(&args.Debug, "debug", debugDefault, debugUsage)
	flag.BoolVar(&args.Debug, "d", debugDefault, debugUsage)
	flag.BoolVar(&args.Dryrun, "dryrun", dryrunDefault, dryrunUsage)
	flag.BoolVar(&args.Dryrun, "n", dryrunDefault, dryrunUsage)
	flag.StringVar(&args.Logfile, "logfile", logfileDefault, logfileUsage)
	flag.StringVar(&args.Logfile, "l", logfileDefault, logfileUsage)
	flag.IntVar(&args.Workers, "workers", workersDefault, workersUsage)
	flag.IntVar(&args.Workers, "w", workersDefault, workersUsage)

	flag.Parse()

	err := 0
	if len(args.Applications) == 0 {
		err++
		fmt.Println("Option -a, --application must be set")
	}

	if err > 0 {
		os.Exit(1)
	}

	return &args
}

func checkArgument(arg string, message string) int {
	if arg == "" {
		fmt.Println(message)
		return 1
	} else {
		return 0
	}
}

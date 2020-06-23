package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/lscheidler/switchctl/common"
)

const (
	version = "0.1"

	applicationUsage   = "set application to switch"
	debugUsage         = "debug mode"
	debugDefault       = false
	environmentDefault = "production"
	environmentUsage   = "set environment to use"
	dryrunDefault      = false
	dryrunUsage        = "do not execute switch"
)

type Arguments struct {
	Applications common.Applications
	Environment  string
	Debug        bool
	Dryrun       bool
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

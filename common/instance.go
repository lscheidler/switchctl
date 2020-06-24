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
package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"go.uber.org/zap"

	"github.com/lscheidler/switchctl/dns"
	"github.com/lscheidler/switchctl/ssh"
)

type Instance struct {
	hostname string
	port     string
	username string

	connected      bool
	currentVersion *Version
	dns            bool
	ssh            *ssh.Ssh
	dryrun         bool

	Commands []*Command
	Errors   []*Error

	slog *zap.SugaredLogger
}

type Command struct {
	Command     string
	Description string

	Stdout       *bytes.Buffer
	StdoutWriter io.Writer
	Stderr       *bytes.Buffer
	StderrWriter io.Writer
	Combined     *bytes.Buffer

	Error error
}

type Version struct {
	CurrentVersion      string `json:"currentVersion"`
	CurrentVersionMtime string `json:"currentVersionMtime"`
}

func NewInstance(slog *zap.SugaredLogger, hostname string, port string, username string, dryrun bool) *Instance {
	return &Instance{
		slog:      slog,
		hostname:  hostname,
		port:      port,
		username:  username,
		connected: false,
		dns:       false,
		dryrun:    dryrun,
		ssh:       nil,
	}
}

func (instance *Instance) Connect() error {
	if dns.Check(instance.hostname) {
		instance.dns = true

		instance.ssh = ssh.New(
			instance.hostname,
			instance.username,
			instance.port,
		)
		if err := instance.ssh.Connect(); err != nil {
			instance.connected = false
			instance.Errors = append(instance.Errors, &Error{Message: err.Error()})
			return err
		} else {
			instance.connected = true
		}
	} else {
		instance.dns = false
	}
	return nil
}

func (instance *Instance) Close() {
	if instance.ssh != nil {
		instance.ssh.Close()
	}
}

func (instance *Instance) Connected() bool {
	return instance.connected
}

func (instance *Instance) CurrentVersion() *Version {
	return instance.currentVersion
}

func (instance *Instance) Dns() bool {
	return instance.dns
}

func (instance *Instance) NewCommand(command string, description string) *Command {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	combined := &bytes.Buffer{}

	commandStruct := Command{
		Command:      command,
		Description:  description,
		Stdout:       stdout,
		StdoutWriter: io.MultiWriter(stdout, combined),
		Stderr:       stderr,
		StderrWriter: io.MultiWriter(stderr, combined),
		Combined:     combined,
	}
	instance.Commands = append(instance.Commands, &commandStruct)
	return &commandStruct
}

func (instance *Instance) GetVersion(application string) *Command {
	command := instance.NewCommand("switch -i -a "+application, "get version information")

	err := instance.ssh.Execute(command.Command, command.StdoutWriter, command.StderrWriter)
	if err != nil {
		command.Error = err
		instance.Errors = append(instance.Errors, &Error{Message: "Failed to retrieve version information"})
		instance.slog.Warn(instance.Hostname() + ": Failed to retrieve version information")
		instance.slog.Warn(command.Combined)
	} else if command.Stdout != nil {
		var version Version
		dec := json.NewDecoder(command.Stdout)
		if err = dec.Decode(&version); err != nil {
			instance.slog.Warn("Cannot unmarshal data: %v", err)
		} else {
			instance.currentVersion = &version
		}
	}

	return command
}

func (instance *Instance) Hostname() string {
	return instance.hostname
}

func (instance *Instance) Prefetch(application string, version string) *Command {
	command := instance.NewCommand("switch -a "+application+" -v "+version+" --prefetch", "prefetch artifact")
	err := instance.ssh.Execute(command.Command, command.StdoutWriter, command.StderrWriter)
	if err != nil {
		command.Error = err
		instance.Errors = append(instance.Errors, &Error{Message: "Failed to prefetch artifact"})
		instance.slog.Warn(instance.Hostname() + ": Failed to prefetch artifact")
		instance.slog.Warn(command.Combined)
		return command
	}
	return command
}

func (instance *Instance) Switch(application string, version string) *Command {
	cmd := "switch -a " + application + " -v " + version + " -y"
	if instance.dryrun {
		cmd = cmd + " -n"
	}

	instance.slog.Info(instance.hostname + ": " + cmd)
	command := instance.NewCommand(cmd, "switch application")
	//command := instance.NewCommand("/home/lscheidler/fail", "switch application")
	err := instance.ssh.Execute(command.Command, command.StdoutWriter, command.StderrWriter)
	if err != nil {
		command.Error = err
		instance.Errors = append(instance.Errors, &Error{Message: "Failed to switch"})
		return command
	}
	return command
}

func (instance *Instance) Completed(function func(string, bool) string) func() string {
	return func() string {
		failed := instance.Commands[len(instance.Commands)-1].Error != nil
		return function(instance.hostname, failed)
	}
}

func (v *Version) String() string {
	if v == nil {
		return "<error>"
	} else if v.CurrentVersion != "" && v.CurrentVersionMtime != "" {
		return fmt.Sprintf("%v [%v]", v.CurrentVersion, v.CurrentVersionMtime)
	} else if v.CurrentVersion != "" {
		return fmt.Sprintf("%v", v.CurrentVersion)
	} else {
		return "<not_found>"
	}
}

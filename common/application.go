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
	"errors"
	"fmt"
	"log"
	"os/user"
	"regexp"
	"strings"
	"text/template"

	"go.uber.org/zap"

	"github.com/lscheidler/switchctl/conf"
	"github.com/lscheidler/switchctl/dns"
)

type Application struct {
	Name    string
	Version string

	SuccessfulInstances []*Instance
	FailedInstances     []*Instance

	Errors []*Error
}

type Error struct {
	Message string
	Command *Command
}

func (e *Error) String() string {
	return e.Message
}

func NewApplication(name string, version string) *Application {
	return &Application{
		Name:    name,
		Version: version,
	}
}

func (application *Application) Load(slog *zap.SugaredLogger, config *conf.Config, environment string, dryrun bool) error {
	if err := application.GetInstances(slog, config, environment, dryrun); err != nil {
		return err
	} else {
		return application.Prefetch(slog, environment)
	}
	return nil
}

func (application *Application) GetInstances(slog *zap.SugaredLogger, conf *conf.Config, environment string, dryrun bool) error {
	for _, entry := range conf.Entries {
		var applicationAlias *string
		applicationFound := false
		environmentFound := false

		for _, applicationS := range entry.Applications {
			if application.Name == applicationS.Name {
				applicationFound = true
				break
			} else if applicationS.Alias != nil && application.Name == *applicationS.Alias {
				applicationFound = true
				applicationAlias = applicationS.Alias
				break
			} else if matched, err := regexp.MatchString(applicationS.Regexp, application.Name); strings.Compare(applicationS.Regexp, "") != 0 && matched && err == nil {
				applicationFound = true
				break
			}
		}
		for _, cEnvironment := range entry.Environments {
			if environment == cEnvironment {
				environmentFound = true
				break
			}
		}

		if applicationFound && environmentFound {
			for i := 1; i <= entry.Instances; i++ {
				instanceNumber := i
				if entry.ReverseOrder {
					instanceNumber = entry.Instances - (i - 1)
				}

				t := template.Must(template.New("instance").Parse(entry.Template))
				var instance bytes.Buffer
				if applicationAlias == nil {
					t.Execute(&instance, &templateData{Application: application.Name, Environment: environment, InstanceNumber: instanceNumber})
				} else {
					t.Execute(&instance, &templateData{Application: *applicationAlias, Environment: environment, InstanceNumber: instanceNumber})
				}

				if dns.Check(instance.String()) {
					application.SuccessfulInstances = append(application.SuccessfulInstances, NewInstance(slog, instance.String(), "22", getUsername(), dryrun))
				}
			}
		}
	}

	instances := application.SuccessfulInstances
	application.SuccessfulInstances = application.SuccessfulInstances[:0]

	for _, instance := range instances {
		if err := instance.Connect(); err == nil {
			application.SuccessfulInstances = append(application.SuccessfulInstances, instance)
			instance.GetVersion(application.Name)
		} else {
			application.FailedInstances = append(application.FailedInstances, instance)
			application.Errors = append(application.Errors, &Error{Message: instance.Hostname() + ": " + err.Error()})
			if strings.Compare(environment, "staging") != 0 {
				return err
			}
		}
	}

	if len(application.SuccessfulInstances) == 0 {
		err := errors.New(application.Name + ": no successful instance found")
		application.Errors = append(application.Errors, &Error{Message: err.Error()})
		return err
	}
	return nil
}

func (application *Application) Close() {
	for _, instance := range application.SuccessfulInstances {
		if instance.Connected() {
			instance.Close()
		}
	}
}

func (application *Application) InstanceHostnames() []func() string {
	var result []func() string
	for _, instance := range application.SuccessfulInstances {
		result = append(result, instance.Hostname)
	}
	return result
}

func (application *Application) InstanceCompleted(function func(string, bool) string) []func() string {
	var result []func() string
	for _, instance := range application.SuccessfulInstances {
		result = append(result, instance.Completed(function))
	}
	return result
}

func (application *Application) Prefetch(slog *zap.SugaredLogger, environment string) error {
	instances := application.SuccessfulInstances
	application.SuccessfulInstances = application.SuccessfulInstances[:0]

	for _, instance := range instances {
		if instance.Connected() {
			if command := instance.Prefetch(application.Name, application.Version); command.Error != nil {
				message := instance.hostname + ": Failed to prefetch artifact " + application.Name + " (" + application.Version + ")"
				application.Errors = append(application.Errors, &Error{Message: message, Command: command})
				application.FailedInstances = append(application.FailedInstances, instance)

				if strings.Compare(environment, "staging") != 0 {
					return command.Error
				}
			} else {
				application.SuccessfulInstances = append(application.SuccessfulInstances, instance)
			}
		}
	}

	if len(application.SuccessfulInstances) == 0 {
		return errors.New(application.Name + ": no successful instance found")
	}
	return nil
}

func (application *Application) Switch(slog *zap.SugaredLogger) error {
	for _, instance := range application.SuccessfulInstances {
		if instance.Connected() && len(instance.Errors) == 0 {
			if command := instance.Switch(application.Name, application.Version); command.Error != nil {
				return command.Error
			}
		}
	}
	return nil
}

type templateData struct {
	Application    string
	Environment    string
	InstanceNumber int
}

func getUsername() string {
	u, err := user.Current()
	if err != nil {
		// TODO error message
		log.Println(err)
		return ""
	}
	return u.Username
}

type Applications []*Application

// String is the method to format the flag's value, part of the flag.Value interface.
// The String method's output will be used in diagnostics.
func (i *Applications) String() string {
	return fmt.Sprint(*i)
}

// Set is the method to set the flag value, part of the flag.Value interface.
// Set's argument is a string to be parsed to set the flag.
// It's a comma-separated list, so we split it.
func (i *Applications) Set(value string) error {
	for _, t := range strings.Split(value, ",") {
		arr := strings.SplitN(t, ":", 2)
		if len(arr) == 2 {
			*i = append(*i, NewApplication(arr[0], arr[1]))
		} else {
			log.Fatal("Option -a, --application must be in format <application>:<version> in \"", t, "\"")
		}
	}
	return nil
}

func (applications *Applications) Close() {
	for _, application := range []*Application(*applications) {
		application.Close()
	}
}

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
package conf

import (
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Entries []*ConfigEntry
}

type ConfigEntry struct {
	Applications []Application `yaml:applications`
	Environments []string      `yaml:environments`
	Instances    []Instance    `yaml:instances`
}

type Instance struct {
	NumberOfInstances int    `yaml:"numberOfInstances"`
	Template          string `yaml:template`
	ReverseOrder      bool   `yaml:"reverseInstanceOrder"`
}

type Application struct {
	Regexp string  `yaml:regexp`
	Name   string  `yaml:name`
	Alias  *string `yaml:user,omitempty`
}

func LoadConfig() *Config {
	var conf []*ConfigEntry
	if filename := findConfigFile(); filename != nil {
		dat, err := ioutil.ReadFile(*filename)
		err = yaml.Unmarshal(dat, &conf)
		if err != nil {
			log.Fatalf("cannot unmarshal data: %v", err)
		}
	} else {
		log.Println("No config file found")
	}
	return &Config{Entries: conf}
}

func findConfigFile() *string {
	usr, _ := user.Current()

	config_path := []string{"./config.yml", filepath.Join(usr.HomeDir, ".config/switchctl/config.yml")}
	for _, path := range config_path {
		if _, err := os.Stat(path); err == nil {
			return &path
		}
	}
	return nil
}

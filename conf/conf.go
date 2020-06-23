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
	Template     string        `yaml:template`
	Instances    int           `yaml:instances`
}

type Application struct {
	Name  string  `yaml:name`
	Alias *string `yaml:user,omitempty`
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

	config_path := []string{"./config.yml", filepath.Join(usr.HomeDir, ".config/switch/config.yml")}
	for _, path := range config_path {
		if _, err := os.Stat(path); err == nil {
			return &path
		}
	}
	return nil
}

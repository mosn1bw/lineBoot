package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"
)

type Configuration struct {
	Bind    string `yaml:"bind"`
	Channel struct {
		Secret string `yaml:"secret"`
		Token  string `yaml:"token"`
	} `yaml:"channel"`
	Source map[string]string
}

var (
	_config *Configuration
)

func init() {
	var err error
	rootDirPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatalf("config: file error: %s", err.Error())
	}
	configPath := filepath.Join(rootDirPath, "app.yml")
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		// config exists
		file, err := ioutil.ReadFile(configPath)
		if err != nil {
			log.Fatalf("config: file error: %s", err.Error())
		}

		_config = newConfiguration()
		err = yaml.Unmarshal(file, &_config)
		if err != nil {
			log.Fatal("config: config error:", err)
		}
	}
}

func newConfiguration() *Configuration {
	return &Configuration{}
}

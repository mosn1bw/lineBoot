package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	yaml "gopkg.in/yaml.v2"
)

type Configuration struct {
	Bind    string `yaml:"bind"`
	Channel struct {
		Secret string `yaml:"secret"`
		Token  string `yaml:"token"`
	} `yaml:"channel"`
	Source     map[string]string
	AppBaseURL string `yaml:"app_base_url"`
}

var (
	_config             *Configuration
	_localZone          *time.Location
	nbaAPIGameURL       string
	nbaAPIGamePlayerURL string
)

func init() {
	var err error
	rootDirPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatalf("config: file error: %s", err.Error())
	}
	configPath := filepath.Join(rootDirPath, "app.yml")
	_config = newConfiguration()
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		// config exists
		file, err := ioutil.ReadFile(configPath)
		if err != nil {
			log.Fatalf("config: file error: %s", err.Error())
		}

		err = yaml.Unmarshal(file, &_config)
		if err != nil {
			log.Fatal("config: config error:", err)
		}
	} else {
		_config.Bind = os.Getenv("PORT")
		_config.Channel.Secret = os.Getenv("ChannelSecret")
		_config.Channel.Token = os.Getenv("ChannelAccessToken")
		_config.Source = map[string]string{
			"nba_game":        os.Getenv("SourceURL"),
			"nba_game_player": os.Getenv("SourcePlayerURL"),
		}
		_config.AppBaseURL = os.Getenv("AppBaseURL")
	}

	var found bool
	nbaAPIGameURL, found = _config.Source["nba_game"]
	if !found {
		panic("config nbaAPIGameURL empty")
	}

	nbaAPIGamePlayerURL, found = _config.Source["nba_game_player"]
	if !found {
		panic("config nbaAPIGamePlayerURL empty")
	}

	_localZone, err = time.LoadLocation("Asia/Taipei")
	if err != nil {
		panic(err)
	}
}

func newConfiguration() *Configuration {
	return &Configuration{}
}

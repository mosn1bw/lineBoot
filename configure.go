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
	_config                  *Configuration
	_localZone               *time.Location
	_fontPath                string
	nbaAPIScoresURL          string
	nbaAPIGameSnapshotURL    string
	nbaAPIGameSnapshotENURL  string
	nbaConferenceStandingAPI string
	nbaBracketURL            string
)

func init() {
	var err error
	rootDirPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatalf("config: file error: %s", err.Error())
	}
	configPath := filepath.Join(rootDirPath, "app.yml")
	_fontPath = filepath.Join(rootDirPath, "font/MicrosoftYaHeiMono-CP950.ttf")
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
			"nba_url": os.Getenv("SourceNBAURL"),
		}
		_config.AppBaseURL = os.Getenv("AppBaseURL")
	}

	var found bool
	var nbaAPIURL string

	nbaAPIURL, found = _config.Source["nba_url"]
	if !found {
		panic("config nba_url empty")
	}

	nbaAPIScoresURL = nbaAPIURL + "/stats2/scores/daily.json?countryCode=TW&locale=zh_TW&tz=%2B8"
	nbaAPIGameSnapshotURL = nbaAPIURL + "/stats2/game/snapshot.json?countryCode=TW&locale=%s&gameId=%s"
	nbaConferenceStandingAPI = nbaAPIURL + "/stats2/season/conferencestanding.json?locale=zh_TW"
	nbaBracketURL = nbaAPIURL + "/stats2/playoff/bracket.json?locale=zh_TW"

	_localZone, err = time.LoadLocation("Asia/Taipei")
	if err != nil {
		panic(err)
	}
}

func newConfiguration() *Configuration {
	return &Configuration{}
}

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	DATE_TIME_LAYOUT    = "01 月02 日 15:04"
	NBA_API_TIME_FORMAT = "2006-01-02"
)

func GetNBAGameByDate(date *time.Time) (*GameInfo, error) {
	nbaquertURL := nbaAPIGameURL + fmt.Sprintf("&gameDate=%s", date.Format(NBA_API_TIME_FORMAT))
	return getNBAGame(nbaquertURL)
}

func GetNBAGameToday() (*GameInfo, error) {
	return getNBAGame(nbaAPIGameURL)
}

func getNBAGame(url string) (*GameInfo, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("error: get error %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Printf("error: get fail %s", resp.Body)
		return nil, fmt.Errorf("status code error %v", resp.Body)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("error: ReadAll error %v", err)
		return nil, err
	}
	data := GameInfo{}
	json.Unmarshal(body, &data)
	return &data, err
}

func GetNBAGamePlayerByGameID(id string) (*GamePlayerInfo, error) {
	url := fmt.Sprintf(nbaAPIGamePlayerURL, id)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("error: get error %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Printf("error: get fail %s", resp.Body)
		return nil, fmt.Errorf("status code error %v", resp.Body)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("error: ReadAll error %v", err)
		return nil, err
	}
	data := GamePlayerInfo{}
	json.Unmarshal(body, &data)
	return &data, err
}

package main

import (
	"log"
	"strings"
	"time"
)

const (
	_prefix = "nba"
)

func TextMessageHandler(receiveMsg string) string {
	sendMsg := ""
	log.Print(receiveMsg)
	cmdMsg, isCmd := parseReceiveMsg(receiveMsg)
	if !isCmd {
		return sendMsg
	}
	switch cmdMsg {
	case "":
		sendMsg = "支援命令: \n   NBA賽事 | NBA今日賽事 | NBA明日賽事"
	case "賽事", "今日賽事":
		data, err := GetNBAameToday()
		if err != nil {
			log.Printf("GetNBAameToday error : %v", err)
		}
		sendMsg = data.ParseToMessage()
	case "明日賽事":
		today, err := GetLocalTimeNow()
		if err != nil {
			log.Printf("GetLocalTimeNow error: %v", err)
		}
		tomorrow := today.Add(24 * time.Hour)
		data, err := GetNBAGameByDate(&tomorrow)
		if err != nil {
			log.Printf("GetNBAGameByDate error :%v, %v", tomorrow, err)
		}
		sendMsg = data.ParseToMessage()
	}
	return sendMsg
}

func parseReceiveMsg(msg string) (string, bool) {
	var recMsg string
	recMsg = strings.Trim(msg, " ")
	recMsg = strings.ToLower(recMsg)
	if len(recMsg) < 3 {
		return "", false
	}
	prefix := recMsg[0:3]
	return recMsg[3:], prefix == _prefix
}

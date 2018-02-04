package main

import (
	"log"
	"strconv"
	"strings"
	"time"
)

func LeftPad2Len(s string, padStr string, overallLen int) string {
	var padCountInt int
	padCountInt = 1 + ((overallLen - len(padStr)) / len(padStr))
	var retStr = strings.Repeat(padStr, padCountInt) + s
	return retStr[(len(retStr) - overallLen):]
}

func UtcMillis2TimeString(utcMillisStr string, timeFormat string) string {
	utcMillis, err := strconv.ParseInt(utcMillisStr, 10, 64)
	if err != nil {
		log.Printf("parse time error: %v", err)
		return ""
	}
	utcTimestamp := utcMillis / 1000
	gameTime := time.Unix(utcTimestamp, 0)
	gameTimeStr := gameTime.Format(timeFormat)
	return gameTimeStr
}

func GetLocalTimeNow() (time.Time, error) {
	now := time.Now()
	local, err := time.LoadLocation("Asia/Taipei")
	if err != nil {
		return now, err
	}
	return now.In(local), nil
}

package main

import "strings"

func ParseReceiveMsg(msg string) string {
	var recMsg string
	recMsg = strings.Trim(msg, " ")
	recMsg = strings.ToLower(recMsg)
	return recMsg
}

func LeftPad2Len(s string, padStr string, overallLen int) string {
	var padCountInt int
	padCountInt = 1 + ((overallLen - len(padStr)) / len(padStr))
	var retStr = strings.Repeat(padStr, padCountInt) + s
	return retStr[(len(retStr) - overallLen):]
}

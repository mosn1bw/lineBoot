package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/line/line-bot-sdk-go/linebot"
)

type NBABotClient struct {
	bot         *linebot.Client
	appBaseURL  string
	downloadDir string
}

func NewNBABotClient(channelSecret, channelToken, appBaseURL string) (*NBABotClient, error) {
	bot, err := linebot.New(
		channelSecret,
		channelToken,
	)
	if err != nil {
		return nil, err
	}
	downloadDir := filepath.Join(filepath.Dir(os.Args[0]), "line-bot")
	_, err = os.Stat(downloadDir)
	if err != nil {
		if err := os.Mkdir(downloadDir, 0777); err != nil {
			return nil, err
		}
	}
	return &NBABotClient{
		bot:         bot,
		appBaseURL:  appBaseURL,
		downloadDir: downloadDir,
	}, nil
}

func (app *NBABotClient) Callback(w http.ResponseWriter, r *http.Request) {
	events, err := app.bot.ParseRequest(r)
	if err != nil {
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(400)
		} else {
			w.WriteHeader(500)
		}
		return
	}
	for _, event := range events {
		// gID := event.Source.GroupID
		// uID := event.Source.UserID
		// rID := event.Source.RoomID
		switch event.Type {
		case linebot.EventTypeMessage:
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				if err := app.handleText(message, event.ReplyToken, event.Source); err != nil {
					log.Print(err)
				}
			}
		}
	}
}

const (
	_cmd_prefix = "nba"
)

func (app *NBABotClient) handleText(message *linebot.TextMessage, replyToken string, source *linebot.EventSource) error {
	sendMsg := ""
	log.Print(message.Text)
	cmdMsg, isCmd := parseReceiveMsg(message.Text)
	if !isCmd {
		return nil
	}
	switch cmdMsg {
	case "":
		imageURL := app.appBaseURL + "/static/buttons/nba.jpg"
		buttons := linebot.NewButtonsTemplate(
			imageURL, "NBA功能列表", "賽事",
			linebot.NewMessageTemplateAction("今日賽事", "NBA今日賽事"),
			linebot.NewMessageTemplateAction("明日賽事", "NBA明日賽事"),
			linebot.NewMessageTemplateAction("昨日賽事", "NBA昨日賽事"),
		)
		if _, err := app.bot.ReplyMessage(
			replyToken,
			linebot.NewTemplateMessage("支援命令: \n   NBA賽事 | NBA今日賽事 | NBA明日賽事 | NBA昨日賽事", buttons),
		).Do(); err != nil {
			return err
		}
		return nil
	// case "profile":
	// 	if source.UserID != "" {
	// 		profile, err := app.bot.GetProfile(source.UserID).Do()
	// 		if err != nil {
	// 			return app.replyText(replyToken, err.Error())
	// 		}
	// 		if _, err := app.bot.ReplyMessage(
	// 			replyToken,
	// 			linebot.NewTextMessage("Display name: "+profile.DisplayName),
	// 			linebot.NewTextMessage("Status message: "+profile.StatusMessage),
	// 		).Do(); err != nil {
	// 			return err
	// 		}
	// 	} else {
	// 		return app.replyText(replyToken, "Bot can't use profile API without user ID")
	// 	}
	case "賽事", "今日賽事":
		data, err := GetNBAameToday()
		if err != nil {
			log.Printf("GetNBAameToday error : %v", err)
		}
		sendMsg = data.ParseToMessage()
	case "明日賽事":
		today, err := GetLocalTime(time.Now())
		if err != nil {
			log.Printf("GetLocalTimeNow error: %v", err)
		}
		tomorrow := today.Add(24 * time.Hour)
		data, err := GetNBAGameByDate(&tomorrow)
		if err != nil {
			log.Printf("GetNBAGameByDate error :%v, %v", tomorrow, err)
		}
		sendMsg = data.ParseToMessage()
	case "昨日賽事":
		today, err := GetLocalTime(time.Now())
		if err != nil {
			log.Printf("GetLocalTimeNow error: %v", err)
		}
		tomorrow := today.Add(-24 * time.Hour)
		data, err := GetNBAGameByDate(&tomorrow)
		if err != nil {
			log.Printf("GetNBAGameByDate error :%v, %v", tomorrow, err)
		}
		sendMsg = data.ParseToMessage()
	}
	if sendMsg != "" {
		if _, err := app.bot.ReplyMessage(
			replyToken,
			linebot.NewTextMessage(sendMsg),
		).Do(); err != nil {
			return err
		}
	}
	return nil
}

func parseReceiveMsg(msg string) (string, bool) {
	var recMsg string
	recMsg = strings.Trim(msg, " ")
	recMsg = strings.ToLower(recMsg)
	if len(recMsg) < 3 {
		return "", false
	}
	prefix := recMsg[0:3]
	return recMsg[3:], prefix == _cmd_prefix
}

func (app *NBABotClient) replyText(replyToken, text string) error {
	if _, err := app.bot.ReplyMessage(
		replyToken,
		linebot.NewTextMessage(text),
	).Do(); err != nil {
		return err
	}
	return nil
}

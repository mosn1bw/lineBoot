package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/line/line-bot-sdk-go/linebot/httphandler"
)

func main() {
	log.Print("Server Start...")
	channelSecret := _config.Channel.Secret
	channelToken := _config.Channel.Token
	handler, err := httphandler.New(
		channelSecret,
		channelToken,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Setup HTTP Server for receiving requests from LINE platform
	handler.HandleEvents(func(events []*linebot.Event, r *http.Request) {
		bot, err := handler.NewClient()
		if err != nil {
			log.Print(err)
			return
		}
		for _, event := range events {
			if event.Type == linebot.EventTypeMessage {
				// gID := event.Source.GroupID
				// uID := event.Source.UserID
				// rID := event.Source.RoomID
				// log.Printf("gid: %s, uid: %s, rid: %s", gID, uID, rID)
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					receiveMsg := message.Text
					sendMsg := ""
					log.Print(receiveMsg)
					cmdMsg := ParseReceiveMsg(receiveMsg)
					switch cmdMsg {
					case "nba":
						data, err := GetNBATodayData()
						if err != nil {
							log.Printf("GetNBATodayData error : %v", err)
						}
						sendMsg = data.ParseToMessage()
					}
					if sendMsg != "" {
						if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(sendMsg)).Do(); err != nil {
							log.Print(err)
						}
					}
					// _, err := bot.PushMessage(uID, linebot.NewTextMessage("21321")).Do()
					// if err != nil {
					// 	log.Printf("PushMessage error : %v", err)
					// }
				}
			}
		}
	})

	http.Handle("/callback", handler)

	// This is just a sample code.
	// For actually use, you must support HTTPS by using `ListenAndServeTLS`, reverse proxy or etc.
	addr := fmt.Sprintf(":%s", _config.Bind)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}

	// data, err := GetNBATodayData()
	// if err != nil {
	// 	log.Printf("GetNBATodayData error : %v", err)
	// }
	// message := data.ParseToMessage()
	// fmt.Printf("msg: \n %s", message)
}

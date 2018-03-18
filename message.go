package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
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
		case linebot.EventTypePostback:
			data := event.Postback.Data
			dataArr := strings.Split(data, "@")
			if len(dataArr) != 2 {
				return
			}
			teamType := dataArr[0]
			gameID := dataArr[1]
			pInfo, err := GetNBAGamePlayerByGameID(gameID)
			if err != nil {
				log.Printf("GetNBAGamePlayerByGameID err: %v", err)
				return
			}
			sendMseeage := " 球員｜位置\n上場時間｜得分｜籃板｜助攻 \n-----------\n"
			messageFmt := "%s | %s\n%s | %d | %d | %d \n-----------\n"
			if teamType == "away" {
				for _, player := range pInfo.Payload.AwayTeam.GamePlayers {
					if player.StatTotal.Mins == 0 {
						continue
					}
					name := fmt.Sprintf("%s-%s", player.Profile.FirstName, player.Profile.LastName)
					position := player.Profile.Position
					upTime := fmt.Sprintf("%d:%d", player.StatTotal.Mins, player.StatTotal.Secs)
					points := player.StatTotal.Points
					rebs := player.StatTotal.Rebs
					assists := player.StatTotal.Assists
					sendMseeage += fmt.Sprintf(messageFmt, name, position, upTime, points, rebs, assists)
				}
			} else {
				for _, player := range pInfo.Payload.HomeTeam.GamePlayers {
					if player.StatTotal.Mins == 0 {
						continue
					}
					name := fmt.Sprintf("%s-%s", player.Profile.FirstName, player.Profile.LastName)
					position := player.Profile.Position
					upTime := fmt.Sprintf("%d:%d", player.StatTotal.Mins, player.StatTotal.Secs)
					points := player.StatTotal.Points
					rebs := player.StatTotal.Rebs
					assists := player.StatTotal.Assists
					sendMseeage += fmt.Sprintf(messageFmt, name, position, upTime, points, rebs, assists)
				}
			}
			if err := app.replyText(event.ReplyToken, sendMseeage); err != nil {
				log.Print(err)
			}
		}
	}
}

const (
	_cmd_prefix = "#"
)

var (
	TodayGameStr                 = "今日賽事"
	TomorrowGameStr              = "明日賽事"
	YesterdayGameStr             = "昨日賽事"
	EasternConferenceStandingStr = "東區戰績"
	WesternConferenceStandingStr = "西區戰績"
	CmdTodayGame                 = _cmd_prefix + TodayGameStr
	CmdTomorrowGame              = _cmd_prefix + TomorrowGameStr
	CmdYesterdayGame             = _cmd_prefix + YesterdayGameStr
	CmdEasternConferenceStanding = _cmd_prefix + EasternConferenceStandingStr
	CmdWesternConferenceStanding = _cmd_prefix + WesternConferenceStandingStr
)

var CmdArray = []string{
	CmdTodayGame,
	CmdTomorrowGame,
	CmdYesterdayGame,
	CmdEasternConferenceStanding,
	CmdWesternConferenceStanding,
}

func (app *NBABotClient) handleText(message *linebot.TextMessage, replyToken string, source *linebot.EventSource) error {
	var sendMsg *linebot.TemplateMessage
	log.Print(message.Text)
	recMsg := strings.Trim(message.Text, " ")
	recMsg = strings.ToUpper(recMsg)
	imageURL := app.appBaseURL + "/static/buttons/nba.png"
	switch recMsg {
	case "NBA":
		column1 := linebot.NewCarouselColumn(
			imageURL, "NBA比分", "賽事即時比分",
			linebot.NewPostbackTemplateAction(TodayGameStr, CmdTodayGame, CmdTodayGame, ""),
			linebot.NewPostbackTemplateAction(TomorrowGameStr, CmdTomorrowGame, CmdTomorrowGame, ""),
			linebot.NewPostbackTemplateAction(YesterdayGameStr, CmdYesterdayGame, CmdYesterdayGame, ""),
		)
		column2 := linebot.NewCarouselColumn(
			imageURL, "NBA戰績", "分區戰績",
			linebot.NewPostbackTemplateAction("分區戰績", "#分區戰績", "#分區戰績", ""),
			linebot.NewPostbackTemplateAction(EasternConferenceStandingStr, CmdEasternConferenceStanding, CmdEasternConferenceStanding, ""),
			linebot.NewPostbackTemplateAction(WesternConferenceStandingStr, CmdWesternConferenceStanding, CmdWesternConferenceStanding, ""),
		)
		columns := []*linebot.CarouselColumn{
			column1,
			column2,
		}
		cmdLine := strings.Join(CmdArray, " | ")

		template := linebot.NewCarouselTemplate(columns...)
		sendMsg = linebot.NewTemplateMessage("支援命令: \n   "+cmdLine, template)
	case "#分區戰績":
		buttons := linebot.NewButtonsTemplate(
			imageURL, "NBA功能列表", "戰績",
			linebot.NewMessageTemplateAction("東區戰績", CmdEasternConferenceStanding),
			linebot.NewMessageTemplateAction("西區戰績", CmdWesternConferenceStanding),
		)
		cmdLine := strings.Join(CmdArray, " | ")
		if _, err := app.bot.ReplyMessage(
			replyToken,
			linebot.NewTemplateMessage("支援命令: \n   "+cmdLine, buttons),
		).Do(); err != nil {
			return err
		}
		return nil
	case CmdTodayGame:
		data, err := GetNBAGameToday()
		if err != nil {
			log.Printf("GetNBAGameToday error : %v", err)
		}
		sendMsg = app.ParseGameInfoToMessage(data)
	case CmdTomorrowGame:
		today, err := GetLocalTime(time.Now())
		if err != nil {
			log.Printf("GetLocalTimeNow error: %v", err)
		}
		tomorrow := today.Add(24 * time.Hour)
		data, err := GetNBAGameByDate(&tomorrow)
		if err != nil {
			log.Printf("GetNBAGameByDate error :%v, %v", tomorrow, err)
		}
		sendMsg = app.ParseGameInfoToMessage(data)
	case CmdYesterdayGame:
		today, err := GetLocalTime(time.Now())
		if err != nil {
			log.Printf("GetLocalTimeNow error: %v", err)
		}
		tomorrow := today.Add(-24 * time.Hour)
		data, err := GetNBAGameByDate(&tomorrow)
		if err != nil {
			log.Printf("GetNBAGameByDate error :%v, %v", tomorrow, err)
		}
		sendMsg = app.ParseGameInfoToMessage(data)

	case CmdEasternConferenceStanding:
		data, err := GetNBAConferenceStanding()
		if err != nil {
			return err
		}
		sendMseeage := app.ParseConferenceStandingToMessage(data, "Eastern")
		if err := app.replyText(replyToken, sendMseeage); err != nil {
			log.Print(err)
		}
	case CmdWesternConferenceStanding:
		data, err := GetNBAConferenceStanding()
		if err != nil {
			return err
		}
		sendMseeage := app.ParseConferenceStandingToMessage(data, "Western")
		if err := app.replyText(replyToken, sendMseeage); err != nil {
			log.Print(err)
		}

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
	}
	if sendMsg != nil {
		if _, err := app.bot.ReplyMessage(
			replyToken,
			sendMsg,
		).Do(); err != nil {
			return err
		}
	}
	return nil
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

func (app *NBABotClient) ParseGameInfoToMessage(data *GameInfo) *linebot.TemplateMessage {
	imageURL := app.appBaseURL + "/static/buttons/nba.png"
	columns := []*linebot.CarouselColumn{}
	message := "     主隊 : 客隊\n"
	for index, val := range data.Payload.Date.Games {
		// ToDo: fixed more than 10
		if index > 9 {
			continue
		}
		var gameInfo string
		homeScore := val.Boxscore.HomeScore
		awayScore := val.Boxscore.AwayScore
		homeTeamName := val.HomeTeam.Profile.Name
		awayTeamName := val.AwayTeam.Profile.Name
		status := val.Boxscore.Status
		switch status {
		case "1": // 未開賽
			gameTimeStr := UtcMillis2TimeString(val.Profile.UtcMillis, DATE_TIME_LAYOUT)
			gameInfo = fmt.Sprintf("未開賽 | %s ", gameTimeStr)
		default: //2: 比賽中, 3: 結束
			gameInfo = fmt.Sprintf(" %3d - %3d | %s %s", homeScore, awayScore, val.Boxscore.StatusDesc, val.Boxscore.PeriodClock)
			// case "3":
			// 	gameInfo = fmt.Sprintf(" %3d : %3d | 結束", homeScore, awayScore)
		}
		teamMessage := fmt.Sprintf("#%d %s vs %s\n      %s", index+1, homeTeamName, awayTeamName, gameInfo)
		message += teamMessage + "\n"

		// template
		teamVS := fmt.Sprintf("#%d %s vs %s", index+1, homeTeamName, awayTeamName)
		btnName1 := fmt.Sprintf("%s 數據統計", homeTeamName)
		btnName2 := fmt.Sprintf("%s 數據統計", awayTeamName)
		column := linebot.NewCarouselColumn(
			imageURL, teamVS, gameInfo,
			linebot.NewPostbackTemplateAction(btnName1, "home@"+val.Profile.GameID, "", ""),
			linebot.NewPostbackTemplateAction(btnName2, "away@"+val.Profile.GameID, "", ""),
		)
		columns = append(columns, column)
	}
	template := linebot.NewCarouselTemplate(columns...)

	return linebot.NewTemplateMessage(message, template)
}

func (app *NBABotClient) ParseConferenceStandingToMessage(data *ConferenceStanding, conference string) string {
	sendMseeage := "排名        ｜  勝負   ｜ 勝差"
	if conference != "Western" {
		sendMseeage = "排名           ｜  勝負   ｜ 勝差"
	}
	msgFormat := "%02d %s｜%s｜ %.1f"
	for _, group := range data.Payload.StandingGroups {
		if strings.ToLower(group.Conference) == strings.ToLower(conference) {
			teams := group.Teams
			sort.Slice(teams, func(i, j int) bool {
				return group.Teams[j].Standings.ConfRank > group.Teams[i].Standings.ConfRank
			})
			teamMsgArr := []string{}
			for _, team := range teams {
				rank := team.Standings.ConfRank
				teamName := parseTeamName(team.Profile.Name, conference)
				winLose := fmt.Sprintf("%2d - %2d", team.Standings.Wins, team.Standings.Losses)
				confGamesBehind := team.Standings.ConfGamesBehind
				msg := fmt.Sprintf(msgFormat, rank, teamName, winLose, confGamesBehind)
				teamMsgArr = append(teamMsgArr, msg)
			}
			teamMsgStr := strings.Join(teamMsgArr, "\n")
			sendMseeage += "\n" + teamMsgStr
			return sendMseeage
		}
	}
	return sendMseeage
}

func parseTeamName(team string, conference string) string {
	if conference == "Western" {
		if len(team) < 9 {
			return team + "   "
		}
	} else {
		if len(team) == 6 {
			return team + "      "
		} else if len(team) < 6 {
			return team + "     "
		}
	}
	return team
}

package main

import (
	"bufio"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/freetype/truetype"
	"github.com/line/line-bot-sdk-go/linebot"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

const (
	_cmd_prefix = "#"
)

var (
	TodayGameStr                 = "今日賽事"
	TomorrowGameStr              = "明日賽事"
	YesterdayGameStr             = "昨日賽事"
	EasternConferenceStandingStr = "東區戰績"
	WesternConferenceStandingStr = "西區戰績"
	GamePlayerBoxExpStr          = "數據統計說明"
	CmdTodayGame                 = _cmd_prefix + TodayGameStr
	CmdTomorrowGame              = _cmd_prefix + TomorrowGameStr
	CmdYesterdayGame             = _cmd_prefix + YesterdayGameStr
	CmdEasternConferenceStanding = _cmd_prefix + EasternConferenceStandingStr
	CmdWesternConferenceStanding = _cmd_prefix + WesternConferenceStandingStr
	CmdGamePlayerBoxExp          = _cmd_prefix + GamePlayerBoxExpStr
)

var CmdArray = []string{
	CmdTodayGame,
	CmdTomorrowGame,
	CmdYesterdayGame,
	CmdEasternConferenceStanding,
	CmdWesternConferenceStanding,
	CmdGamePlayerBoxExp,
}

type NBABotClient struct {
	bot *linebot.Client
	sync.RWMutex
	appBaseURL     string
	standingImgURL string
	allGameImgURL  string
	nbaImgURL      string
	downloadDir    string
	commandCounter map[string]int
	initTime       *time.Time
}

func NewNBABotClient(channelSecret, channelToken, appBaseURL string) (*NBABotClient, error) {
	bot, err := linebot.New(
		channelSecret,
		channelToken,
	)
	now := time.Now().UTC()
	now = now.Add(8 * time.Hour)
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

	cmdCounter := map[string]int{}
	for _, cmd := range CmdArray {
		cmdCounter[cmd] = 0
	}
	imgPath := appBaseURL + "/static/buttons/"
	return &NBABotClient{
		bot:            bot,
		appBaseURL:     appBaseURL,
		downloadDir:    downloadDir,
		commandCounter: cmdCounter,
		initTime:       &now,
		standingImgURL: imgPath + "standing.png",
		allGameImgURL:  imgPath + "allgame.png",
		nbaImgURL:      imgPath + "nba.png",
	}, nil
}

func (app *NBABotClient) Callback(c *gin.Context) {
	r := c.Request
	events, err := app.bot.ParseRequest(r)
	if err != nil {
		if err == linebot.ErrInvalidSignature {
			c.Writer.WriteHeader(400)
		} else {
			c.Writer.WriteHeader(500)
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
			app.handlePostBack(data, event.ReplyToken)
		}
	}
}

func (app *NBABotClient) Statistic(c *gin.Context) {
	app.RLock()
	defer app.RUnlock()
	response := ""
	var keys []string
	for k := range app.commandCounter {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		response += fmt.Sprintf("%s : %d\n", key, app.commandCounter[key])
	}
	response += fmt.Sprintf("統計開始時間： %s", app.initTime.Format(DATE_TIME_LAYOUT))

	fmt.Fprintf(c.Writer, "%s", response)
}

func (app *NBABotClient) CounterIncs(key string) {
	app.Lock()
	app.commandCounter[key] += 1
	app.Unlock()
}

func (app *NBABotClient) handleText(message *linebot.TextMessage, replyToken string, source *linebot.EventSource) error {
	var sendMsg linebot.Message
	var err error
	log.Print(message.Text)
	recMsg := strings.Trim(message.Text, " ")
	recMsg = strings.ToUpper(recMsg)
	recMsgArr := strings.Split(recMsg, "@")
	page := 0
	if len(recMsgArr) > 1 {
		page, err = strconv.Atoi(recMsgArr[1])
		if err != nil {
			return err
		}
	}
	switch recMsgArr[0] {
	case "NBA":
		column1 := linebot.NewCarouselColumn(
			app.allGameImgURL, "NBA比分", "賽事即時比分",
			linebot.NewPostbackTemplateAction(TodayGameStr, CmdTodayGame, CmdTodayGame, ""),
			linebot.NewPostbackTemplateAction(TomorrowGameStr, CmdTomorrowGame, CmdTomorrowGame, ""),
			linebot.NewPostbackTemplateAction(YesterdayGameStr, CmdYesterdayGame, CmdYesterdayGame, ""),
		)
		column2 := linebot.NewCarouselColumn(
			app.standingImgURL, "NBA戰績", "分區戰績",
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
		app.CounterIncs(recMsg)
	case "#分區戰績":
		buttons := linebot.NewButtonsTemplate(
			app.standingImgURL, "NBA功能列表", "戰績",
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
		app.CounterIncs(recMsg)
		return nil
	case CmdTodayGame:
		data, err := GetNBAGameToday()
		if err != nil {
			log.Printf("GetNBAGameToday error : %v", err)
		}
		sInfo := parseGameInfoToGameScoreInfo(data)
		sendMsg = app.ParseGameScoreInfoToMessage(&ParseGameScoreOpt{
			data:     sInfo,
			cmd:      recMsgArr[0],
			page:     page,
			showList: true,
		})
		app.CounterIncs(recMsg)
	case CmdTomorrowGame:
		today, err := GetLocalTime(time.Now())
		if err != nil {
			log.Printf("GetLocalTimeNow error: %v", err)
		}
		tomorrow := today.Add(2 * 24 * time.Hour)
		data, err := GetNBAGameByDate(&tomorrow)
		if err != nil {
			log.Printf("GetNBAGameByDate error :%v, %v", tomorrow, err)
		}
		sInfo := parseGameInfoToGameScoreInfo(data)
		sendMsg = app.ParseGameScoreInfoToMessage(&ParseGameScoreOpt{
			data:     sInfo,
			cmd:      recMsgArr[0],
			page:     page,
			showList: true,
		})
		app.CounterIncs(recMsg)
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
		sInfo := parseGameInfoToGameScoreInfo(data)
		sendMsg = app.ParseGameScoreInfoToMessage(&ParseGameScoreOpt{
			data:     sInfo,
			cmd:      recMsgArr[0],
			page:     page,
			showList: true,
		})
		app.CounterIncs(recMsg)

	case CmdEasternConferenceStanding:
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		imageUrl := app.appBaseURL + "/standing/Eastern?version=" + timestamp
		if _, err := app.bot.ReplyMessage(
			replyToken,
			linebot.NewImageMessage(imageUrl, imageUrl),
		).Do(); err != nil {
			log.Printf("ParseConferenceStandingToMessage imageUrl err: %v", err)
		}
		app.CounterIncs(recMsg)
	case CmdWesternConferenceStanding:
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		imageUrl := app.appBaseURL + "/standing/Western?version=" + timestamp
		if _, err := app.bot.ReplyMessage(
			replyToken,
			linebot.NewImageMessage(imageUrl, imageUrl),
		).Do(); err != nil {
			log.Printf("ParseConferenceStandingToMessage imageUrl err: %v", err)
		}
		app.CounterIncs(recMsg)

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
	case CmdGamePlayerBoxExp:
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		imageUrl := app.appBaseURL + "/gamecol/info?version=" + timestamp
		if _, err := app.bot.ReplyMessage(
			replyToken,
			linebot.NewImageMessage(imageUrl, imageUrl),
		).Do(); err != nil {
			log.Printf("CmdGamePlayerBoxExp imageUrl err: %v", err)
		}
		app.CounterIncs(recMsg)
	default:
		app.CounterIncs("其它")
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

func (app *NBABotClient) handlePostBack(data string, replyToken string) {
	dataArr := strings.Split(data, "@")
	if len(dataArr) != 3 {
		return
	}
	msgType := dataArr[0]
	action := dataArr[1]
	gameID := dataArr[2]

	switch msgType {
	case "player":
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		imageUrl := app.appBaseURL + "/game/" + gameID + "/" + action + "?version=" + timestamp
		if _, err := app.bot.ReplyMessage(
			replyToken,
			linebot.NewImageMessage(imageUrl, imageUrl),
		).Do(); err != nil {
			log.Printf("GetNBAGamePlayerByGameID imageUrl err: %v", err)
		}
		app.CounterIncs("#比賽數據統計")
	case "score":
		pInfo, err := GetNBAGamePlayerByGameID(gameID, "zh_TW")
		if err != nil {
			log.Printf("score GetNBAGamePlayerByGameID err: %v", err)
			return
		}

		sInfo := parseGamePlayerInfoToGameScoreInfo(pInfo)
		sendMsg := app.ParseGameScoreInfoToMessage(&ParseGameScoreOpt{
			data:     sInfo,
			showList: false,
		})
		if sendMsg != nil {
			if _, err := app.bot.ReplyMessage(
				replyToken,
				sendMsg,
			).Do(); err != nil {
				return
			}
		}
		app.CounterIncs("更新比分")
	}

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

type GameScoreInfo struct {
	Boxscore      GameBoxscore
	GameID        string
	GameTime      string
	HomeTeamName  string
	AwayTeamName  string
	HighlightsURL string
}

type ParseGameScoreOpt struct {
	data     []*GameScoreInfo
	page     int
	cmd      string
	showList bool
}

func (app *NBABotClient) ParseGameScoreInfoToMessage(opt *ParseGameScoreOpt) linebot.Message {
	data := opt.data
	gameNum := len(data)
	page := opt.page
	if gameNum == 0 {
		return linebot.NewTextMessage("當日無賽事")
	}
	if page <= 0 {
		page = 1
	}

	columns := []*linebot.CarouselColumn{}

	startIndex := 0
	endIndex := gameNum

	listBtnText := "更新比分"
	listBtnCmd := opt.cmd
	if gameNum > 7 {
		perPage := gameNum / 2
		startIndex = perPage * (page - 1)
		endIndex = perPage * page
		if startIndex >= gameNum {
			app.ParseGameScoreInfoToMessage(&ParseGameScoreOpt{
				data:     data,
				page:     page - 1,
				cmd:      opt.cmd,
				showList: true,
			})
		}
		if endIndex > gameNum {
			endIndex = gameNum
		}
		if page == 1 {
			listBtnText = "下一頁"
			listBtnCmd += fmt.Sprintf("@%d", page+1)
		}
	}

	if opt.showList && page == 1 {
		firstColumn := linebot.NewCarouselColumn(
			app.allGameImgURL, "賽事選單", "賽事選單",
			linebot.NewPostbackTemplateAction(listBtnText, listBtnText, listBtnCmd, ""),
			linebot.NewPostbackTemplateAction("數據統計說明", "數據統計說明", "#數據統計說明", ""),
			linebot.NewPostbackTemplateAction("功能列表", "功能列表", "NBA", ""),
		)
		columns = append(columns, firstColumn)
	}
	message := "     主隊 : 客隊\n"

	for index := startIndex; index < endIndex; index++ {
		val := data[index]
		var gameInfo string
		homeScore := val.Boxscore.HomeScore
		awayScore := val.Boxscore.AwayScore
		homeTeamName := val.HomeTeamName
		awayTeamName := val.AwayTeamName
		status := val.Boxscore.Status

		btnName1 := fmt.Sprintf("%s 數據統計", homeTeamName)
		btnName2 := fmt.Sprintf("%s 數據統計", awayTeamName)
		btnName3 := "更新比分 - "

		btnData1 := "player@home@" + val.GameID
		btnData2 := "player@away@" + val.GameID
		btnData3 := "score@update@" + val.GameID

		var bt3 linebot.TemplateAction
		switch status {
		case "1": // 1: 未開賽
			btnName3 += "未開賽"
			gameInfo = fmt.Sprintf("未開賽 | %s ", val.GameTime)
			bt3 = linebot.NewPostbackTemplateAction(btnName3, btnData3, "", "")
		case "2": // 2: 比賽中
			btnName3 += "進行中"
			gameInfo = fmt.Sprintf(" %3d - %3d | %s %s", homeScore, awayScore, val.Boxscore.StatusDesc, val.Boxscore.PeriodClock)
			bt3 = linebot.NewPostbackTemplateAction(btnName3, btnData3, "", "")
		case "3": // 3: 結束
			btnName3 += "比賽結束"
			gameInfo = fmt.Sprintf(" %3d - %3d | %s %s", homeScore, awayScore, val.Boxscore.StatusDesc, val.Boxscore.PeriodClock)
			bt3 = linebot.NewURITemplateAction("觀看 Highlights", val.HighlightsURL)
		}
		teamMessage := fmt.Sprintf("#%d %s vs %s\n      %s", index+1, homeTeamName, awayTeamName, gameInfo)
		message += teamMessage + "\n"

		// template
		teamVS := fmt.Sprintf("#%d %s vs %s", index+1, homeTeamName, awayTeamName)

		column := linebot.NewCarouselColumn(
			app.nbaImgURL, teamVS, gameInfo,
			linebot.NewPostbackTemplateAction(btnName1, btnData1, "", ""),
			linebot.NewPostbackTemplateAction(btnName2, btnData2, "", ""),
			bt3,
		)
		columns = append(columns, column)
	}

	template := linebot.NewCarouselTemplate(columns...)

	return linebot.NewTemplateMessage(message, template)
}

var PlayerInfoColumn = []string{"球員", "位置", "上場時間", "得分", "籃板", "助攻"}

func (app *NBABotClient) ParsePlayInfoToImgMessage(c *gin.Context, pInfo *GamePlayerInfo, teamType string) {
	title := UtcMillis2TimeString(pInfo.Payload.GameProfile.UtcMillis, DATE_TIME_LAYOUT)

	homeTeamName := pInfo.Payload.HomeTeam.Profile.Name
	awayTeamName := pInfo.Payload.AwayTeam.Profile.Name
	title += fmt.Sprintf("  %s VS %s", homeTeamName, awayTeamName)

	awayMsgArr := playInfoToMsgArr(pInfo, "away")
	awayOpt := &TextToImageOpt{
		SubTitle: "客 - " + awayTeamName,
		TextData: awayMsgArr,
	}

	if len(awayMsgArr) < 2 {
		awayOpt.SubTitle = "未開賽"
		awayOpt.TextData = [][]string{}
	}
	homeMsgArr := playInfoToMsgArr(pInfo, "home")
	homeOpt := &TextToImageOpt{
		Title:    title,
		SubTitle: "主 - " + homeTeamName,
		TextData: homeMsgArr,
	}
	if len(homeMsgArr) < 2 {
		homeOpt.SubTitle = "未開賽"
		homeOpt.TextData = [][]string{}
	}

	convertTextArrToTableImage(c, []*TextToImageOpt{
		homeOpt, awayOpt,
	}, title)
}

func playInfoToMsgArr(pInfo *GamePlayerInfo, teamType string) [][]string {
	messageArr := [][]string{}
	messageArr = append(messageArr, PlayerInfoColumn)
	gamePlayers := []GamePlayers{}
	if teamType == "away" {
		gamePlayers = pInfo.Payload.AwayTeam.GamePlayers
	} else {
		gamePlayers = pInfo.Payload.HomeTeam.GamePlayers
	}

	for _, player := range gamePlayers {
		if player.StatTotal.Mins == 0 {
			continue
		}
		mArr := []string{}
		name := fmt.Sprintf("%s-%s", player.Profile.FirstName, player.Profile.LastName)
		position := player.Profile.Position
		upTime := fmt.Sprintf("%02d:%02d", player.StatTotal.Mins, player.StatTotal.Secs)
		points := strconv.Itoa(player.StatTotal.Points)
		rebs := strconv.Itoa(player.StatTotal.Rebs)
		assists := strconv.Itoa(player.StatTotal.Assists)
		mArr = append(mArr, name, position, upTime, points, rebs, assists)
		messageArr = append(messageArr, mArr)
	}

	return messageArr
}

func (app *NBABotClient) ParsePlayInfoToDetailImgMessage(c *gin.Context, pInfo *GamePlayerInfo, teamType string) {
	title := UtcMillis2TimeString(pInfo.Payload.GameProfile.UtcMillis, DATE_TIME_LAYOUT)

	homeTeamName := pInfo.Payload.HomeTeam.Profile.Name
	awayTeamName := pInfo.Payload.AwayTeam.Profile.Name
	title += fmt.Sprintf("  %s VS %s", homeTeamName, awayTeamName)

	infoOpt := &TextToImageOpt{}
	if teamType == "away" {
		awayMsgArr := playInfoToDetailMsgArr(pInfo, "away")
		infoOpt.SubTitle = awayTeamName
		infoOpt.TextData = awayMsgArr
	} else {
		homeMsgArr := playInfoToDetailMsgArr(pInfo, "home")
		infoOpt.SubTitle = homeTeamName
		infoOpt.TextData = homeMsgArr
	}
	if len(infoOpt.TextData) < 2 {
		infoOpt.SubTitle = "未開賽"
		infoOpt.TextData = [][]string{}
	}

	convertTextArrToTableImage(c, []*TextToImageOpt{
		infoOpt,
	}, title)
}

// ToDo: 暫無 ＢＡ
type StaticsColumn struct {
	CName string
	EName string
}

var PlayerInfoDetailMapColumn = []StaticsColumn{
	{CName: "姓名", EName: "PLAYERS"},
	{CName: "位置", EName: "POS"},
	{CName: "上場時間", EName: "MIN"},
	{CName: "投籃命中-投籃出手", EName: "FGM-A"},
	{CName: "三分球命中數-三分球出手數", EName: "3PM-A"},
	{CName: "罰球命中-罰球次數", EName: "FTM-A"},
	{CName: "+/-", EName: "+/-"},
	{CName: "進攻籃板", EName: "OR"},
	{CName: "防守籃板", EName: "DR"},
	{CName: "籃板", EName: "TR"},
	{CName: "助攻", EName: "AS"},
	{CName: "犯規", EName: "PF"},
	{CName: "抄截", EName: "ST"},
	{CName: "失誤", EName: "TO"},
	{CName: "阻攻", EName: "BS"},
	{CName: "得分", EName: "PTS"},
	{CName: "EFF", EName: "EFF"},
}

func playInfoToDetailMsgArr(pInfo *GamePlayerInfo, teamType string) [][]string {
	messageArr := [][]string{}
	columns := []string{}
	for _, col := range PlayerInfoDetailMapColumn {
		columns = append(columns, col.EName)
	}
	messageArr = append(messageArr, columns)
	gamePlayers := []GamePlayers{}
	if teamType == "away" {
		gamePlayers = pInfo.Payload.AwayTeam.GamePlayers
	} else {
		gamePlayers = pInfo.Payload.HomeTeam.GamePlayers
	}

	for _, player := range gamePlayers {
		if player.StatTotal.Mins == 0 {
			continue
		}
		mArr := []string{}
		name := fmt.Sprintf("%s. %s", player.Profile.FirstInitial, player.Profile.LastName)
		position := player.Profile.Position
		upTime := fmt.Sprintf("%02d:%02d", player.StatTotal.Mins, player.StatTotal.Secs)
		fgmFga := fmt.Sprintf("%d-%d", player.StatTotal.Fgm, player.StatTotal.Fga)
		tpmtpa := fmt.Sprintf("%d-%d", player.StatTotal.Tpm, player.StatTotal.Tpa)
		ftmfta := fmt.Sprintf("%d-%d", player.StatTotal.Ftm, player.StatTotal.Fta)
		plusMinus := player.Boxscore.PlusMinus
		offRebs := player.StatTotal.OffRebs
		defRebs := player.StatTotal.DefRebs
		totalRebs := offRebs + defRebs
		assists := player.StatTotal.Assists
		fouls := player.StatTotal.Fouls
		turnovers := player.StatTotal.Turnovers
		blocks := player.StatTotal.Blocks
		points := player.StatTotal.Points
		steals := player.StatTotal.Steals

		// EFF: (PTS + TRB + AST + STL + BLK) - (FGA-FGM) - (FTA-FTM) - TO
		eff := (points + totalRebs + assists + steals + blocks) - (player.StatTotal.Fga - player.StatTotal.Fgm) - (player.StatTotal.Fta - player.StatTotal.Ftm) - turnovers

		mArr = append(mArr, name, position, upTime, fgmFga, tpmtpa, ftmfta, plusMinus, strconv.Itoa(offRebs), strconv.Itoa(defRebs), strconv.Itoa(totalRebs), strconv.Itoa(assists), strconv.Itoa(fouls), strconv.Itoa(steals), strconv.Itoa(turnovers), strconv.Itoa(blocks), strconv.Itoa(points), strconv.Itoa(eff))
		messageArr = append(messageArr, mArr)
	}

	return messageArr
}

var StandingInfoColumn = []string{"", "排名", "勝負", "勝差"}

func (app *NBABotClient) ParseConferenceStandingToImgMessage(c *gin.Context, data *ConferenceStanding, conference string) {
	messageArr := [][]string{}
	messageArr = append(messageArr, StandingInfoColumn)
	title := ""
	for _, group := range data.Payload.StandingGroups {
		if strings.ToLower(group.Conference) == strings.ToLower(conference) {
			teams := group.Teams
			sort.Slice(teams, func(i, j int) bool {
				return group.Teams[j].Standings.ConfRank > group.Teams[i].Standings.ConfRank
			})
			for _, team := range teams {
				mArr := []string{}
				rank := fmt.Sprintf("%02d", team.Standings.ConfRank)
				teamName := team.Profile.Name
				winLose := fmt.Sprintf("%2d - %2d", team.Standings.Wins, team.Standings.Losses)
				confGamesBehind := fmt.Sprintf("%.1f", team.Standings.ConfGamesBehind)
				mArr = append(mArr, rank, teamName, winLose, confGamesBehind)
				messageArr = append(messageArr, mArr)
			}
		}
	}
	opt := &TextToImageOpt{
		TextData: messageArr,
	}
	if conference == "Eastern" {
		title = "東區戰績"
	} else {
		title = "西區戰績"
	}

	convertTextArrToTableImage(c, []*TextToImageOpt{opt}, title)
}

type TextToImageOpt struct {
	Title    string
	SubTitle string
	TextData [][]string
}

func convertTextArrToTableImage(c *gin.Context, opts []*TextToImageOpt, title string) {
	size := float64(20)
	dpi := float64(72)
	spacing := float64(2)
	y := 20 + int(math.Ceil(size*dpi/72))
	dy := int(math.Ceil(size * spacing * dpi / 72))

	imgTH, imgTW := y+dy, 0

	//get total height / width
	for _, opt := range opts {
		if len(opt.TextData) == 0 && len(title) == 0 {
			continue
		}
		imgH, imgW := getImageHeightWidth(y, dy, opt, title)
		imgTH += imgH
		if imgW > imgTW {
			imgTW = imgW
		}
	}
	if imgTH == 0 || imgTW == 0 {
		return
	}

	// Read the font data.
	fontfile := _fontPath
	fontBytes, err := ioutil.ReadFile(fontfile)
	if err != nil {
		log.Println(err)
		return
	}
	f, err := truetype.Parse(fontBytes)
	if err != nil {
		log.Println(err)
		return
	}
	fg, bg := image.White, image.Black

	rgba := image.NewRGBA(image.Rect(0, 0, imgTW, imgTH))
	draw.Draw(rgba, rgba.Bounds(), bg, image.ZP, draw.Src)
	h := font.HintingNone
	d := &font.Drawer{
		Dst: rgba,
		Src: fg,
		Face: truetype.NewFace(f, &truetype.Options{
			Size:    size,
			DPI:     dpi,
			Hinting: h,
		}),
	}

	// draw title
	yAxisLast := y
	d.Dot = fixed.Point26_6{
		X: (fixed.I(imgTW) - d.MeasureString(title)) / 2,
		Y: fixed.I(yAxisLast),
	}
	d.DrawString(title)

	for _, opt := range opts {
		textData := opt.TextData
		subTitle := opt.SubTitle
		if len(subTitle) > 0 {
			yAxisLast += dy
			d.Dot = fixed.Point26_6{
				X: (fixed.I(imgTW) - d.MeasureString(subTitle)) / 2,
				Y: fixed.I(yAxisLast),
			}
			d.DrawString(subTitle)
		}

		if len(textData) > 0 {
			preTextLen := 0
			xAxis := 20
			yAxis := yAxisLast
			for index := 0; index < len(textData[0]); index++ {
				maxTextlen := 0

				yAxis = yAxisLast
				xAxis += preTextLen*11 + 20
				for _, row := range textData {
					yAxis += dy
					text := row[index]
					textLen := getRealTextLength(text)
					if textLen > maxTextlen {
						maxTextlen = textLen
					}
					d.Dot = fixed.P(xAxis, yAxis)
					d.DrawString(text)
				}
				preTextLen = maxTextlen
			}
			yAxisLast = yAxis
		}
	}

	b := bufio.NewWriter(c.Writer)
	err = png.Encode(b, rgba)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	err = b.Flush()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func (app *NBABotClient) getGamePlayInfo(c *gin.Context) {
	gameID := c.Param("gameid")
	teamType := c.Param("type")

	pInfo, err := GetNBAGamePlayerByGameID(gameID, "zh_TW")
	if err != nil {
		log.Printf("GetNBAGamePlayerByGameID err: %v", err)
		return
	}
	app.CounterIncs("比賽數據圖片")
	app.ParsePlayInfoToImgMessage(c, pInfo, teamType)
}

func (app *NBABotClient) getGamePlayInfoEN(c *gin.Context) {
	gameID := c.Param("gameid")
	teamType := c.Param("type")

	pInfo, err := GetNBAGamePlayerByGameID(gameID, "en")
	if err != nil {
		log.Printf("GetNBAGamePlayerByGameID err: %v", err)
		return
	}
	app.CounterIncs("比賽數據圖片")
	app.ParsePlayInfoToDetailImgMessage(c, pInfo, teamType)
}

func (app *NBABotClient) getStandingInfo(c *gin.Context) {
	conference := c.Param("conference")
	data, err := GetNBAConferenceStanding()
	if err != nil {
		log.Printf("getStandingInfo err: %v", err)
	}
	app.CounterIncs("戰績圖片")
	app.ParseConferenceStandingToImgMessage(c, data, conference)
}

func getRealTextLength(str string) int {
	count := 0
	for _, char := range str {
		charStr := fmt.Sprintf("%c", char)
		if len(charStr) > 2 {
			count += 2
		} else {
			count += 1
		}
	}
	return count
}

func getImageHeightWidth(y int, dy int, opt *TextToImageOpt, title string) (int, int) {
	maxWidth := getRealTextLength(title) * 11
	textData := opt.TextData
	imgH := dy * len(textData)
	if len(opt.SubTitle) > 0 {
		imgH += dy
		SubTitleWidth := getRealTextLength(opt.SubTitle) * 11
		if SubTitleWidth > maxWidth {
			maxWidth = SubTitleWidth
		}
	}

	if len(textData) > 0 {
		preTextLen := 0
		xAxis := 20
		for index := 0; index < len(textData[0]); index++ {
			maxTextlen := 0
			xAxis += preTextLen*11 + 20
			for _, row := range textData {
				text := row[index]
				textLen := getRealTextLength(text)
				if textLen > maxTextlen {
					maxTextlen = textLen
				}
			}
			preTextLen = maxTextlen
		}
		xAxis += preTextLen*11 + 20
		if xAxis > maxWidth {
			maxWidth = xAxis
		}
	}
	maxWidth += 20

	return imgH, maxWidth
}

func parseGameInfoToGameScoreInfo(data *GameInfo) []*GameScoreInfo {
	gameInfoArr := []*GameScoreInfo{}
	for _, game := range data.Payload.Date.Games {
		highlightsURL := ""
		for _, val := range game.Urls {
			if val.Type == "Highlights" {
				highlightsURL = val.Value
				break
			}
		}
		gameInfoArr = append(gameInfoArr, &GameScoreInfo{
			Boxscore:      game.Boxscore,
			GameID:        game.Profile.GameID,
			HomeTeamName:  game.HomeTeam.Profile.Name,
			AwayTeamName:  game.AwayTeam.Profile.Name,
			GameTime:      UtcMillis2TimeString(game.Profile.UtcMillis, DATE_TIME_LAYOUT),
			HighlightsURL: highlightsURL,
		})
	}
	return gameInfoArr
}

func parseGamePlayerInfoToGameScoreInfo(data *GamePlayerInfo) []*GameScoreInfo {
	game := GameScoreInfo{
		Boxscore:     data.Payload.Boxscore,
		GameID:       data.Payload.GameProfile.GameID,
		HomeTeamName: data.Payload.HomeTeam.Profile.Name,
		AwayTeamName: data.Payload.AwayTeam.Profile.Name,
		GameTime:     UtcMillis2TimeString(data.Payload.GameProfile.UtcMillis, DATE_TIME_LAYOUT),
	}
	return []*GameScoreInfo{&game}
}

func (app *NBABotClient) getGameColumnInfo(c *gin.Context) {
	data := [][]string{}
	for _, col := range PlayerInfoDetailMapColumn {
		row := []string{col.EName, col.CName}
		data = append(data, row)
	}
	convertTextArrToTableImage(c, []*TextToImageOpt{
		{
			TextData: data,
		},
	}, "數據統計說明")
}

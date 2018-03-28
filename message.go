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

	cmdCounter := map[string]int{
		"NBA":     0,
		"#分區戰績":   0,
		"#比賽數據統計": 0,
		"#其它":     0,
	}
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
			dataArr := strings.Split(data, "@")
			if len(dataArr) != 2 {
				return
			}
			timestamp := strconv.FormatInt(time.Now().Unix(), 10)
			teamType := dataArr[0]
			gameID := dataArr[1]
			imageUrl := app.appBaseURL + "/game/" + gameID + "/" + teamType + "?version=" + timestamp
			if _, err := app.bot.ReplyMessage(
				event.ReplyToken,
				linebot.NewImageMessage(imageUrl, imageUrl),
			).Do(); err != nil {
				log.Printf("GetNBAGamePlayerByGameID imageUrl err: %v", err)
			}
			app.CounterIncs("比賽數據統計")
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
	var sendMsg *linebot.TemplateMessage
	log.Print(message.Text)
	recMsg := strings.Trim(message.Text, " ")
	recMsg = strings.ToUpper(recMsg)
	switch recMsg {
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
		sendMsg = app.ParseGameInfoToMessage(data)
		app.CounterIncs(recMsg)
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
		sendMsg = app.ParseGameInfoToMessage(data)
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
			app.nbaImgURL, teamVS, gameInfo,
			linebot.NewPostbackTemplateAction(btnName1, "home@"+val.Profile.GameID, "", ""),
			linebot.NewPostbackTemplateAction(btnName2, "away@"+val.Profile.GameID, "", ""),
		)
		columns = append(columns, column)
	}
	template := linebot.NewCarouselTemplate(columns...)

	return linebot.NewTemplateMessage(message, template)
}

func (app *NBABotClient) ParseConferenceStandingToMessage(data *ConferenceStanding, conference string) string {
	sendMseeage := "排名       ｜  勝負   ｜ 勝差"
	msgFormat := "%02d %4s｜%s｜%.1f"
	for _, group := range data.Payload.StandingGroups {
		if strings.ToLower(group.Conference) == strings.ToLower(conference) {
			teams := group.Teams
			sort.Slice(teams, func(i, j int) bool {
				return group.Teams[j].Standings.ConfRank > group.Teams[i].Standings.ConfRank
			})
			teamMsgArr := []string{}
			for _, team := range teams {
				rank := team.Standings.ConfRank
				teamName := team.Profile.Name
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

// func (app *NBABotClient) ParsePlayInfoToTextMessage(pInfo *GamePlayerInfo, teamType string) string {
// 	sendMseeage := " 球員｜位置\n上場時間｜得分｜籃板｜助攻 \n-----------\n"
// 	messageFmt := "%s | %s\n%s | %d | %d | %d \n-----------\n"
// 	if teamType == "away" {
// 		for _, player := range pInfo.Payload.AwayTeam.GamePlayers {
// 			if player.StatTotal.Mins == 0 {
// 				continue
// 			}
// 			name := fmt.Sprintf("%s-%s", player.Profile.FirstName, player.Profile.LastName)
// 			position := player.Profile.Position
// 			upTime := fmt.Sprintf("%d:%d", player.StatTotal.Mins, player.StatTotal.Secs)
// 			points := player.StatTotal.Points
// 			rebs := player.StatTotal.Rebs
// 			assists := player.StatTotal.Assists
// 			sendMseeage += fmt.Sprintf(messageFmt, name, position, upTime, points, rebs, assists)
// 		}
// 	} else {
// 		for _, player := range pInfo.Payload.HomeTeam.GamePlayers {
// 			if player.StatTotal.Mins == 0 {
// 				continue
// 			}
// 			name := fmt.Sprintf("%s-%s", player.Profile.FirstName, player.Profile.LastName)
// 			position := player.Profile.Position
// 			upTime := fmt.Sprintf("%d:%d", player.StatTotal.Mins, player.StatTotal.Secs)
// 			points := player.StatTotal.Points
// 			rebs := player.StatTotal.Rebs
// 			assists := player.StatTotal.Assists
// 			sendMseeage += fmt.Sprintf(messageFmt, name, position, upTime, points, rebs, assists)
// 		}
// 	}
// 	return sendMseeage
// }

var PlayerInfoColumn = []string{"球員", "位置", "上場時間", "得分", "籃板", "助攻"}

func (app *NBABotClient) ParsePlayInfoToImgMessage(c *gin.Context, pInfo *GamePlayerInfo, teamType string) {
	messageArr := [][]string{}
	messageArr = append(messageArr, PlayerInfoColumn)
	title := UtcMillis2TimeString(pInfo.Payload.GameProfile.UtcMillis, DATE_TIME_LAYOUT)
	subTitle := ""
	homeTeamName := pInfo.Payload.HomeTeam.Profile.Name
	awayTeamName := pInfo.Payload.AwayTeam.Profile.Name
	title += fmt.Sprintf("  %s VS %s", homeTeamName, awayTeamName)
	if teamType == "away" {
		subTitle += "客 - " + awayTeamName
		for _, player := range pInfo.Payload.AwayTeam.GamePlayers {
			if player.StatTotal.Mins == 0 {
				continue
			}
			mArr := []string{}
			name := fmt.Sprintf("%s-%s", player.Profile.FirstName, player.Profile.LastName)
			position := player.Profile.Position
			upTime := fmt.Sprintf("%d:%d", player.StatTotal.Mins, player.StatTotal.Secs)
			points := strconv.Itoa(player.StatTotal.Points)
			rebs := strconv.Itoa(player.StatTotal.Rebs)
			assists := strconv.Itoa(player.StatTotal.Assists)
			mArr = append(mArr, name, position, upTime, points, rebs, assists)
			messageArr = append(messageArr, mArr)
		}
	} else {
		subTitle += "主 - " + homeTeamName
		for _, player := range pInfo.Payload.HomeTeam.GamePlayers {
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
	}
	if len(messageArr) < 2 {
		subTitle = "未開賽"
		messageArr = [][]string{}
	}

	convertTextArrToimage(c, &TextToImageOpt{
		Title:    title,
		SubTitle: subTitle,
		TextData: messageArr,
		ImgWidth: 720,
	})
}

var StandingInfoColumn = []string{"", "排名", "勝負", "勝差"}

func (app *NBABotClient) ParseConferenceStandingToImgMessage(c *gin.Context, data *ConferenceStanding, conference string) {
	messageArr := [][]string{}
	messageArr = append(messageArr, StandingInfoColumn)
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
		ImgWidth: 370,
	}
	if conference == "Eastern" {
		opt.Title = "東區戰績"
	} else {
		opt.Title = "西區戰績"
	}

	convertTextArrToimage(c, opt)
}

type TextToImageOpt struct {
	Title    string
	SubTitle string
	TextData [][]string
	ImgWidth int
}

func convertTextArrToimage(c *gin.Context, opt *TextToImageOpt) {
	textData := opt.TextData
	title := opt.Title
	subTitle := opt.SubTitle

	if len(textData) == 0 && len(title) == 0 {
		return
	}
	size := float64(20)
	dpi := float64(72)
	spacing := float64(2)
	fontfile := _fontPath
	// Read the font data.
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

	y := 20 + int(math.Ceil(size*dpi/72))
	dy := int(math.Ceil(size * spacing * dpi / 72))
	imgH := y + dy*(len(textData)+1)
	if len(subTitle) > 0 {
		imgH += dy
	}
	imgW := opt.ImgWidth
	rgba := image.NewRGBA(image.Rect(0, 0, imgW, imgH))
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
	d.Dot = fixed.Point26_6{
		X: (fixed.I(imgW) - d.MeasureString(title)) / 2,
		Y: fixed.I(y),
	}

	d.DrawString(title)

	if len(subTitle) > 0 {
		d.Dot = fixed.Point26_6{
			X: (fixed.I(imgW) - d.MeasureString(subTitle)) / 2,
			Y: fixed.I(dy + y),
		}
		d.DrawString(subTitle)
	}

	if len(textData) > 0 {
		preTextLen := 0
		xAxis := 20
		for index := 0; index < len(textData[0]); index++ {
			maxTextlen := 0
			yAxis := y
			if len(subTitle) > 0 {
				yAxis += dy
			}
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

	pInfo, err := GetNBAGamePlayerByGameID(gameID)
	if err != nil {
		log.Printf("GetNBAGamePlayerByGameID err: %v", err)
		return
	}
	app.CounterIncs("比賽數據圖片")
	app.ParsePlayInfoToImgMessage(c, pInfo, teamType)
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

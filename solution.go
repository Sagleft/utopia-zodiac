package main

import (
	"encoding/json"
	"errors"
	"flag"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/IvanMenshykov/MoonPhase"
	"github.com/go-stack/stack"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/nfnt/resize"
	"github.com/yanzay/tbot/v2"
)

const (
	timeLayoutUS    = "January 2"
	timeLayoutMonth = "January"
	timeLayoutDay   = "Monday"
	timeLayoutYear  = "2006"
	postImageWidth  = 800
	postImageHeight = 450
	postImageInput  = "img/post_template.png"
	postImageOutput = "post.png"

	postMaxLength = 4096

	moonCornerPosX = 32
	moonCornerPosY = 18
	moonSize       = 60
	moonTitlePosX  = 100
	moonTitlePosY  = 76
)

type solution struct {
	Config  solutionParams
	APIBase string
	API     apiVariants
}

type solutionParams struct {
	BotToken        string
	ChannelChatID   string
	TimeVariant     string
	ReplaceWordFrom string
	RaplaceWordTo   string
	DebugMode       bool
}

type apiVariants map[string]string

type sunsignData struct {
	Tag  string
	Icon string
}

var sunSigns = []sunsignData{
	sunsignData{Tag: "aquarius", Icon: "♒️"},
	sunsignData{Tag: "pisces", Icon: "♓️"},
	sunsignData{Tag: "aries", Icon: "♈️"},
	sunsignData{Tag: "taurus", Icon: "♉️"},
	sunsignData{Tag: "gemini", Icon: "♊️"},
	sunsignData{Tag: "cancer", Icon: "♋️"},
	sunsignData{Tag: "leo", Icon: "♌️"},
	sunsignData{Tag: "virgo", Icon: "♍️"},
	sunsignData{Tag: "libra", Icon: "♎️"},
	sunsignData{Tag: "scorpio", Icon: "♏️"},
	sunsignData{Tag: "sagittarius", Icon: "♐️"},
	sunsignData{Tag: "capricorn", Icon: "♑️"},
}

//https://api.telegram.org/bot<id>/getChat?chat_id=@<chat_tag>
func newSolution() *solution {
	stack.Caller(0) //init error stack
	return &solution{
		Config: solutionParams{
			BotToken: "",
			//ChannelChatID: "@horoscopes_for_you",
			ChannelChatID:   "@testgozod",
			TimeVariant:     "day",
			ReplaceWordFrom: "Ganesha",
			RaplaceWordTo:   "a wise man",
		},
		APIBase: "http://horoscope-api.herokuapp.com/horoscope/",
		API: apiVariants{
			"today": "", "week": "", "month": "", "year": "",
		},
	}
}

func (sol *solution) parseArgs() error {
	timeVariant := flag.String("variant", "today", "today/week/month/year")
	isDebugMode := flag.Bool("debug", false, "debug mode disable notification & show debug log")
	flag.Parse()
	if timeVariant == nil {
		return errors.New("time variant is not set or invalid")
	}
	if isDebugMode != nil {
		sol.Config.DebugMode = *isDebugMode
	}
	//fmt.Println("is debug: ", sol.Config.DebugMode)
	sol.Config.TimeVariant = *timeVariant
	if !sol.isTimeVariantExists() {
		return errors.New("unknown time variant given")
	}
	return nil
}

func (sol *solution) isTimeVariantExists() bool {
	_, exists := sol.API[sol.Config.TimeVariant]
	return exists
}

/*func (sol *solution) getAPIurl() string {
	if !sol.isTimeVariantExists() {
		return ""
	}
	return sol.API[sol.Config.TimeVariant]
}*/

type horoscopeResponse struct {
	Date    string `json:"date"`
	Text    string `json:"horoscope"`
	Sunsign string `json:"sunsign"`
}

func (sol *solution) getZodiacForecast(sunsign string) (*horoscopeResponse, error) {
	URL := sol.APIBase + sol.Config.TimeVariant + "/" + sunsign
	//fmt.Println(URL)
	resp, err := http.Get(URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	//fmt.Println(string(body))
	hObj := horoscopeResponse{}
	parseErr := json.Unmarshal(body, &hObj)
	if parseErr != nil {
		return nil, parseErr
	}
	//feels Genesha -> sage tells
	hObj.Text = strings.Replace(hObj.Text, "feels "+sol.Config.ReplaceWordFrom, sol.Config.RaplaceWordTo+" tells", -1)
	//Genesha -> sage
	hObj.Text = strings.Replace(hObj.Text, sol.Config.ReplaceWordFrom, sol.Config.RaplaceWordTo, -1)
	//. sage -> . A sage
	hObj.Text = strings.Replace(hObj.Text, ". "+sol.Config.RaplaceWordTo, ". "+strings.Title(sol.Config.RaplaceWordTo), -1)
	return &hObj, nil
}

func (sol *solution) run() error {
	err := sol.parseArgs()
	if err != nil {
		panic(err)
	}

	err = sol.makePost()
	if err != nil {
		return err
	}

	err = sol.createPostImage(time.Now())
	if err != nil {
		return err
	}

	err = sol.sendPostImage()
	if err != nil {
		return err
	}
	return nil
}

func (sol *solution) makePost() error {
	var postText string = ""
	if !sol.isTimeVariantExists() {
		return errors.New("failed to get api url for timevariant")
	}
	switch sol.Config.TimeVariant {
	case "today":
		timeFormated := time.Now().Format(timeLayoutUS)
		postText += "Horoscope for " + timeFormated + "\n\n"
	case "month":
		timeFormated := time.Now().Format(timeLayoutMonth)
		postText += "Horoscope for " + timeFormated + "\n\n"
	case "year":
		timeFormated := time.Now().Format(timeLayoutYear)
		postText += "Horoscope for " + timeFormated + "year\n\n"
	}

	for i, sunsignInfo := range sunSigns {
		sunsign := sunsignInfo.Tag
		forecastResponse, err := sol.getZodiacForecast(sunsign)
		if err != nil {
			return err
		}
		newPostPart := sunsignInfo.Icon + " " + strings.Title(sunsign) + "\n✨ " + forecastResponse.Text + "\n\n"
		//check post length
		if len(postText+newPostPart) > postMaxLength || i == len(sunSigns)-1 {
			//send post part or full post (if it is last post part)
			if i == len(sunSigns)-1 {
				postText += sol.Config.ChannelChatID
			}
			err := sol.sendPost(postText)
			if err != nil {
				return err
			}
			postText = ""
		} else {
			postText += newPostPart
		}
	}
	//fmt.Println(postText)
	return nil
}

type telegramResponse struct {
	OK          bool                   `json:"ok"`
	Result      telegramResponseResult `json:"result"`
	Description string                 `json:"description"`
}

type telegramResponseResult struct {
	MessageID int64 `json:"message_id"`
	Date      int64 `json:"date"`
}

func (sol *solution) sendPost(postText string) error {
	//https://api.telegram.org/bot<token>/sendMessage?chat_id=<...>&text=<...>
	tgAPIURL := "https://api.telegram.org/bot" + sol.Config.BotToken +
		"/sendMessage?chat_id=" + sol.Config.ChannelChatID +
		"&text=" + url.QueryEscape(postText)

	if sol.Config.DebugMode {
		tgAPIURL += "&disable_notification=true"
	}
	resp, err := http.Get(tgAPIURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	tResp := telegramResponse{}
	parseErr := json.Unmarshal(body, &tResp)
	if parseErr != nil {
		return parseErr
	}
	if !tResp.OK {
		return errors.New(tResp.Description)
	}

	//pin post if time variant "month" given
	if sol.Config.TimeVariant == "month" {
		err := sol.pinChatMessage(tResp.Result.MessageID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (sol *solution) pinChatMessage(msgID int64) error {
	tgAPIURL := "https://api.telegram.org/bot" + sol.Config.BotToken +
		"/pinChatMessage?chat_id=" + sol.Config.ChannelChatID +
		"&message_id=" + strconv.FormatInt(msgID, 10) + "&disable_notification=true"

	resp, err := http.Get(tgAPIURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	tResp := telegramResponse{}
	parseErr := json.Unmarshal(body, &tResp)
	if parseErr != nil {
		return parseErr
	}
	if !tResp.OK {
		return errors.New(tResp.Description)
	}
	return nil
}

func loadFont(fontFilePath string) (*truetype.Font, error) {
	b, err := ioutil.ReadFile(fontFilePath)
	if err != nil {
		return nil, err
	}
	fontHandler, err := truetype.Parse(b)
	if err != nil {
		return nil, err
	}
	return fontHandler, nil
}

func addLabel(img *image.RGBA, x, y int, fontSize float64, label string, fontHandler *truetype.Font) error {
	fontContext := freetype.NewContext()
	fontContext.SetDPI(72)
	fontContext.SetFont(fontHandler)
	fontContext.SetFontSize(fontSize)
	fontContext.SetDst(img)
	fontContext.SetClip(img.Bounds())
	fontContext.SetSrc(image.NewUniform(color.Gray16{0x3030}))
	pt := freetype.Pt(x, y+int(fontContext.PointToFixed(fontSize)>>6))
	//pt := freetype.Pt(x, y)

	if _, err := fontContext.DrawString(label, pt); err != nil {
		return err
	}
	return nil
}

func (sol *solution) createPostImage(timeData time.Time) error {
	// get moon phase
	time := time.Now()
	moonPhaseData := MoonPhase.New(time)

	// get file io
	f, err := os.Open(postImageInput)
	if err != nil {
		return nil
	}
	defer f.Close()

	// load image
	img, _, err := image.Decode(f)
	if err != nil {
		return err
	}

	// draw moon phase
	var moonPhaseIndex int = int(math.Round(moonPhaseData.Phase()))
	if moonPhaseIndex == 8 {
		moonPhaseIndex = 7 //index fix
	}
	moonImagePath := "img/moon" + strconv.Itoa(moonPhaseIndex) + ".png"
	moonImgFileHandler, err := os.Open(moonImagePath)
	if err != nil {
		return err
	}
	moonImg, _, err := image.Decode(moonImgFileHandler)
	defer f.Close()
	if err != nil {
		return err
	}
	moonPoint := image.Pt(moonCornerPosX-moonSize, moonCornerPosY-moonSize)
	imageResult := image.NewRGBA(img.Bounds())
	draw.Draw(imageResult, img.Bounds(), img, image.Point{0, 0}, draw.Over)

	resizedMoon := resize.Resize(moonSize, moonSize, moonImg, resize.Lanczos3)
	draw.Draw(imageResult, moonImg.Bounds(), resizedMoon, moonPoint, draw.Over)

	// load fonts
	fontRegular, err := loadFont("fonts/Akrobat-Regular.ttf")
	if err != nil {
		return err
	}
	fontBold, err := loadFont("fonts/Akrobat-Bold.ttf")
	if err != nil {
		return err
	}

	// draw moon phase info
	// drawStringPoint := freetype.Pt(10, 10+int(fontRegularContext.PointToFixed(24)>>6))
	// drawStringPoint := freetype.Pt(10, 10)
	addLabel(imageResult, moonTitlePosX, moonTitlePosY, 20, moonPhaseData.PhaseName(), fontRegular)

	// draw day info
	addLabel(imageResult, 650, 40, 28, timeData.Format(timeLayoutUS), fontBold)
	addLabel(imageResult, 680, 80, 20, timeData.Format(timeLayoutDay), fontRegular)

	// save image
	f, err = os.Create(postImageOutput)
	if err != nil {
		return err
	}
	err = png.Encode(f, imageResult)
	defer f.Close()
	return err
}

func (sol *solution) sendPostImage() error {
	bot := tbot.New(sol.Config.BotToken)
	client := bot.Client()
	_, err := client.SendPhotoFile(sol.Config.ChannelChatID, "post.png")
	return err
}

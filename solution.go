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
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/IvanMenshykov/MoonPhase"
	swissknife "github.com/Sagleft/swiss-knife"
	utopiago "github.com/Sagleft/utopialib-go/v2"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/nfnt/resize"
	"github.com/yanzay/tbot/v2"
)

func main() {
	cfg := config{}
	if err := swissknife.ParseStructFromJSONFile(configFilePath, &cfg); err != nil {
		log.Fatalln(err)
	}

	sol := newSolution(cfg)

	if err := sol.parseArgs(); err != nil {
		log.Fatalln(err)
	}

	if err := sol.utopiaConnect(); err != nil {
		log.Fatalln(err)
	}

	if err := sol.makePost(); err != nil {
		log.Fatalln(err)
	}

	if err := sol.createPostImage(time.Now()); err != nil {
		log.Fatalln(err)
	}

	if err := sol.sendPostImage(); err != nil {
		log.Fatalln(err)
	}
}

func newSolution(cfg config) *solution {
	return &solution{
		Config: cfg,
		Utopia: utopiago.NewUtopiaClient(cfg.Utopia),
	}
}

func (app *solution) utopiaConnect() error {
	if !app.Utopia.CheckClientConnection() {
		return errors.New("failed to open connection to Utopia")
	}
	return nil
}

func (sol *solution) parseArgs() error {
	// TODO: move to config

	timeVariant := flag.String("variant", "today", "today/week/month/year")
	isDebugMode := flag.Bool("debug", false, "debug mode disable notification & show debug log")
	flag.Parse()
	if timeVariant == nil {
		return errors.New("time variant is not set or invalid")
	}
	if isDebugMode != nil {
		sol.Config.DebugMode = *isDebugMode
	}

	sol.Config.TimeVariant = *timeVariant
	if !sol.isTimeVariantExists() {
		return errors.New("unknown time variant given")
	}
	return nil
}

func (sol *solution) isTimeVariantExists() bool {
	_, exists := APITimeVariants[sol.Config.TimeVariant]
	return exists
}

type horoscopeResponse struct {
	Date    string `json:"date"`
	Text    string `json:"horoscope"`
	Sunsign string `json:"sunsign"`
}

func (sol *solution) getZodiacForecast(sunsign string) (*horoscopeResponse, error) {
	URL := APIBaseURL + sol.Config.TimeVariant + "/" + sunsign
	resp, err := http.Get(URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	hObj := horoscopeResponse{}
	parseErr := json.Unmarshal(body, &hObj)
	if parseErr != nil {
		return nil, parseErr
	}

	//feels Genesha -> sage tells
	hObj.Text = strings.Replace(
		hObj.Text, "feels "+sol.Config.ReplaceWordFrom,
		sol.Config.RaplaceWordTo+" tells", -1,
	)
	//Genesha -> sage
	hObj.Text = strings.Replace(
		hObj.Text, sol.Config.ReplaceWordFrom,
		sol.Config.RaplaceWordTo, -1,
	)
	//. sage -> . A sage
	hObj.Text = strings.Replace(hObj.Text, ". "+sol.Config.RaplaceWordTo, ". "+
		strings.ToTitle(sol.Config.RaplaceWordTo), -1)
	return &hObj, nil
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
		newPostPart := sunsignInfo.Icon + " " + strings.ToTitle(sunsign) + "\n✨ " +
			forecastResponse.Text + "\n\n"

		//check post length
		if len(postText+newPostPart) > postMaxLength || i == len(sunSigns)-1 {
			//send post part or full post (if it is last post part)
			if i == len(sunSigns)-1 {
				postText += sol.Config.ChannelID
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

func (app *solution) sendPost(postText string) error {
	//https://api.telegram.org/bot<token>/sendMessage?chat_id=<...>&text=<...>
	/*tgAPIURL := "https://api.telegram.org/bot" + app.Config.ChannelID +
	"/sendMessage?chat_id=" + app.Config.ChannelID +
	"&text=" + url.QueryEscape(postText)*/

	if app.Config.DebugMode {
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

	// pin post if time variant "month" given
	if app.Config.TimeVariant == "month" {
		err := app.pinChatMessage(tResp.Result.MessageID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (app *solution) pinChatMessage(msgID int64) error {
	/*tgAPIURL := "https://api.telegram.org/bot" + sol.Config.BotToken +
	"/pinChatMessage?chat_id=" + sol.Config.ChannelID +
	"&message_id=" + strconv.FormatInt(msgID, 10) + "&disable_notification=true"*/

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
	_, err := client.SendPhotoFile(sol.Config.ChannelID, "post.png")
	return err
}

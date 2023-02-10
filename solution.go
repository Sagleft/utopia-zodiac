package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/IvanMenshykov/MoonPhase"
	swissknife "github.com/Sagleft/swiss-knife"
	utopiago "github.com/Sagleft/utopialib-go/v2"
	"github.com/nfnt/resize"
)

func main() {
	cfg := config{}
	if err := swissknife.ParseStructFromJSONFile(configFilePath, &cfg); err != nil {
		log.Fatalln(err)
	}

	sol := newSolution(cfg)

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

func (sol *solution) isTimeVariantExists() bool {
	_, exists := APITimeVariants[sol.Config.TimeVariant]
	return exists
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

	// feels Genesha -> sage tells
	hObj.Text = strings.Replace(
		hObj.Text, "feels "+sol.Config.WordFilter.ReplaceWordFrom,
		sol.Config.WordFilter.RaplaceWordTo+" tells", -1,
	)
	// Genesha -> sage
	hObj.Text = strings.Replace(
		hObj.Text, sol.Config.WordFilter.ReplaceWordFrom,
		sol.Config.WordFilter.RaplaceWordTo, -1,
	)
	// . sage -> . A sage
	hObj.Text = strings.Replace(hObj.Text, ". "+sol.Config.WordFilter.RaplaceWordTo, ". "+
		strings.ToTitle(sol.Config.WordFilter.RaplaceWordTo), -1)
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

		// check post length
		if len(postText+newPostPart) > postMaxLength || i == len(sunSigns)-1 {
			// send post part or full post (if it is last post part)
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

func (app *solution) sendPost(postText string) error {
	msg := url.QueryEscape(postText)
	_, err := app.Utopia.SendChannelMessage(app.Config.ChannelID, msg)
	return err

	// pin post if time variant "month" given
	/*if app.Config.TimeVariant == "month" {
		err := app.pinChatMessage(tResp.Result.MessageID)
		if err != nil {
			return err
		}
	}*/
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

func (app *solution) sendPostImage() error {
	imageBytes, err := swissknife.ReadFileToBytes(postImageOutput)
	if err != nil {
		return fmt.Errorf("read post image: %w", err)
	}

	imgEncoded := base64.StdEncoding.EncodeToString(imageBytes)

	_, err = app.Utopia.SendChannelPicture(app.Config.ChannelID, imgEncoded, "", postImageFilename)
	return err
}

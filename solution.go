package main

import (
	"errors"
	"log"
	"strings"
	"time"

	swissknife "github.com/Sagleft/swiss-knife"
	utopiago "github.com/Sagleft/utopialib-go/v2"
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

func (sol *solution) isTimeVariantExists() bool {
	_, exists := APITimeVariants[sol.Config.TimeVariant]
	return exists
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

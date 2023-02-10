package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func (app *solution) getZodiacForecast(sunsign string) (*horoscopeResponse, error) {
	url := fmt.Sprintf(
		"%s?day=%s&sunsign=%s",
		APIBaseURL, app.Config.TimeVariant, sunsign,
	)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create new request: %w", err)
	}

	req.Header.Add("X-RapidAPI-Key", app.Config.APIKey)
	req.Header.Add("X-RapidAPI-Host", APIDomain)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	hObj := horoscopeResponse{}
	parseErr := json.Unmarshal(body, &hObj)
	if parseErr != nil {
		return nil, parseErr
	}

	// feels Genesha -> sage tells
	hObj.Text = strings.Replace(
		hObj.Text, "feels "+app.Config.WordFilter.ReplaceWordFrom,
		app.Config.WordFilter.RaplaceWordTo+" tells", -1,
	)
	// Genesha -> sage
	hObj.Text = strings.Replace(
		hObj.Text, app.Config.WordFilter.ReplaceWordFrom,
		app.Config.WordFilter.RaplaceWordTo, -1,
	)
	// . sage -> . A sage
	hObj.Text = strings.Replace(hObj.Text, ". "+app.Config.WordFilter.RaplaceWordTo, ". "+
		strings.ToTitle(app.Config.WordFilter.RaplaceWordTo), -1)
	return &hObj, nil
}

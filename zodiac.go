package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
)

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

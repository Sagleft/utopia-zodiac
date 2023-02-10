package main

import utopiago "github.com/Sagleft/utopialib-go/v2"

type solution struct {
	Config config
	Utopia utopiago.Client
}

type config struct {
	ChannelID   string          `json:"channelID"`
	TimeVariant string          `json:"timeVariant"`
	WordFilter  wordFilter      `json:"wordReplace"`
	DebugMode   bool            `json:"debug"`
	Utopia      utopiago.Config `json:"utopia"`
}

type wordFilter struct {
	ReplaceWordFrom string `json:"from"`
	RaplaceWordTo   string `json:"to"`
}

type apiVariants map[string]string

type sunsignData struct {
	Tag  string
	Icon string
}

type horoscopeResponse struct {
	Date    string `json:"date"`
	Text    string `json:"horoscope"`
	Sunsign string `json:"sunsign"`
}

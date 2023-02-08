package main

type solution struct {
	Config solutionParams
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

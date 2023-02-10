package main

const (
	timeLayoutUS            = "January 2"
	timeLayoutMonth         = "January"
	timeLayoutDay           = "Monday"
	timeLayoutYear          = "2006"
	postImageWidth          = 800
	postImageHeight         = 450
	postImageInput          = "img/post_template.png"
	postImageOutput         = "post.png"
	moonPhaseImageExtension = "png"
	moonPhaseImagesDir      = "img/moon"
	postImageFilename       = postImageOutput
	configFilePath          = "config.json"
	fontRegularPath         = "fonts/Akrobat-Regular.ttf"
	fontBoldPath            = "fonts/Akrobat-Bold.ttf"

	postMaxLength = 4096

	moonCornerPosX = 32
	moonCornerPosY = 18
	moonSize       = 60
	moonTitlePosX  = 100
	moonTitlePosY  = 76
)

const APIDomain = "horoscope-astrology.p.rapidapi.com"
const APIBaseURL = "https://" + APIDomain + "/horoscope"

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

var APITimeVariants = apiVariants{
	"today": "", "week": "", "month": "", "year": "",
}

package main

import (
	"image"
	"image/draw"
	"image/png"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/IvanMenshykov/MoonPhase"
	"github.com/nfnt/resize"
)

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
	moonImagePath := moonPhaseImagesDir + strconv.Itoa(moonPhaseIndex) + "." + moonPhaseImageExtension
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
	fontRegular, err := loadFont(fontRegularPath)
	if err != nil {
		return err
	}
	fontBold, err := loadFont(fontBoldPath)
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

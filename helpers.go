package main

import (
	"image"
	"image/color"
	"io/ioutil"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
)

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

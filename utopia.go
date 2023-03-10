package main

import (
	"encoding/base64"
	"errors"
	"fmt"

	swissknife "github.com/Sagleft/swiss-knife"
)

func (app *solution) utopiaConnect() error {
	if !app.Utopia.CheckClientConnection() {
		return errors.New("failed to open connection to Utopia. check host, token or port")
	}
	return nil
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

func (app *solution) sendPost(postText string) error {
	_, err := app.Utopia.SendChannelMessage(app.Config.ChannelID, postText)
	return err

	// TBD: pin post if time variant "month" given
}

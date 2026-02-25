package main

import (
	"context"
	"fmt"
	"os"

	"github.com/coze-dev/coze-go"
)

func main() {
	// Get an access_token through personal access token or oauth.
	token := os.Getenv("COZE_API_TOKEN")
	authCli := coze.NewTokenAuth(token)

	// Init the Coze client through the access_token.
	cozeCli := coze.NewCozeAPI(authCli, coze.WithBaseURL(os.Getenv("COZE_API_BASE")))

	botID := os.Getenv("COZE_BOT_ID")
	voiceID := os.Getenv("COZE_VOICE_ID")

	ctx := context.Background()
	resp, err := cozeCli.Audio.Rooms.Create(ctx, &coze.CreateAudioRoomsReq{
		BotID:   botID,
		VoiceID: voiceID,
	})
	if err != nil {
		fmt.Println("Error creating rooms:", err)
		return
	}

	fmt.Println(resp)
	fmt.Println("Room ID:", resp.RoomID)
	fmt.Println("Log ID:", resp.Response().LogID())
}

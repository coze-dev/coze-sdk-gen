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

	// saveFilePath := os.Getenv("SAVE_FILE_PATH")
	voiceID := os.Getenv("COZE_VOICE_ID")
	content := "Come and try it out"

	ctx := context.Background()
	resp, err := cozeCli.Audio.Speech.Create(ctx, &coze.CreateAudioSpeechReq{
		Input:   content,
		VoiceID: voiceID,
	})
	if err != nil {
		fmt.Println("Error creating speech:", err)
		return
	}

	fmt.Println(resp)
	fmt.Println(resp.LogID())

	if err := resp.WriteToFile(os.Getenv("SAVE_SPEECH_PATH")); err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
}

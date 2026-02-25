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

	ctx := context.Background()
	// you can use iterator to automatically retrieve next page
	voices, err := cozeCli.Audio.Voices.List(ctx, &coze.ListAudioVoicesReq{PageSize: 10})
	if err != nil {
		fmt.Println("Error fetching voices:", err)
		return
	}
	for voices.Next() {
		fmt.Println(voices.Current())
	}

	if voices.Err() != nil {
		fmt.Println("Error fetching voices:", voices.Err())
		return
	}

	// the page result will return followed information
	fmt.Println("has_more:", voices.HasMore())
}

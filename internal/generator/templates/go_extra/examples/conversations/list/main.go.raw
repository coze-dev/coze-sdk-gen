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
	botID := os.Getenv("PUBLISHED_BOT_ID")

	// Init the Coze client through the access_token.
	cozeCli := coze.NewCozeAPI(authCli, coze.WithBaseURL(os.Getenv("COZE_API_BASE")))

	ctx := context.Background()
	// you can use iterator to automatically retrieve next page
	conversations, err := cozeCli.Conversations.List(ctx, &coze.ListConversationsReq{PageSize: 2, BotID: botID})
	if err != nil {
		fmt.Println("Error fetching conversations:", err)
		return
	}
	for conversations.Next() {
		fmt.Println(conversations.Current())
	}
	if conversations.Err() != nil {
		fmt.Println("Error fetching conversations:", conversations.Err())
		return
	}

	// the page result will return followed information
	fmt.Println("has_more:", conversations.HasMore())
}

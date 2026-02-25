package main

import (
	"context"
	"fmt"
	"os"

	"github.com/coze-dev/coze-go"
)

// This examples is for describing how to retrieve a bot, fetch published bot list from the API.
// The document for those interface:
func main() {
	// Get an access_token through personal access token or oauth.
	token := os.Getenv("COZE_API_TOKEN")
	botID := os.Getenv("PUBLISHED_BOT_ID")
	authCli := coze.NewTokenAuth(token)

	// Init the Coze client through the access_token.
	cozeCli := coze.NewCozeAPI(authCli, coze.WithBaseURL(os.Getenv("COZE_API_BASE")))

	ctx := context.Background()

	//
	// retrieve a bot

	botInfo, err := cozeCli.Bots.Retrieve(ctx, &coze.RetrieveBotsReq{
		BotID: botID,
	})
	if err != nil {
		fmt.Println("Error retrieving bot:", err)
		return
	}
	fmt.Println(botInfo.Bot)
	fmt.Println("Log ID:", botInfo.LogID())

	//
	// get published bot list

	pageNum := 1
	workspaceID := os.Getenv("WORKSPACE_ID")
	botList, err := cozeCli.Bots.List(ctx, &coze.ListBotsReq{
		SpaceID:  workspaceID,
		PageNum:  pageNum,
		PageSize: 4,
	})
	if err != nil {
		fmt.Println("Error listing bots:", err)
		return
	}

	// you can use iterator to automatically retrieve next page
	for botList.Next() {
		fmt.Println(botList.Current())
	}

	if botList.Err() != nil {
		fmt.Println("Error listing bots:", botList.Err())
		return
	}

	// the page result will return followed information
	fmt.Println("total:", botList.Total())
	fmt.Println("has_more:", botList.HasMore())
}

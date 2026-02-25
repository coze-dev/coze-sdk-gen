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
	conversationID := os.Getenv("CONVERSATION_ID")

	// Init the Coze client through the access_token.
	cozeCli := coze.NewCozeAPI(authCli, coze.WithBaseURL(os.Getenv("COZE_API_BASE")))

	ctx := context.Background()
	// you can use iterator to automatically retrieve next page
	message, err := cozeCli.Conversations.Messages.List(ctx, &coze.ListConversationsMessagesReq{Limit: 2, ConversationID: conversationID})
	if err != nil {
		fmt.Println("Error fetching message:", err)
		return
	}
	for message.Next() {
		fmt.Println(message.Current())
	}
	if message.Err() != nil {
		fmt.Println("Error fetching message:", message.Err())
		return
	}

	// the page result will return followed information
	fmt.Println("has_more:", message.HasMore())
}

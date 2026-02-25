package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/coze-dev/coze-go"
)

func main() {
	// Get an access_token through personal access token or oauth.
	token := os.Getenv("COZE_API_TOKEN")
	authCli := coze.NewTokenAuth(token)
	botID := os.Getenv("COZE_BOT_ID")

	// Init the Coze client through the access_token.
	cozeCli := coze.NewCozeAPI(authCli, coze.WithBaseURL(os.Getenv("COZE_API_BASE")))

	ctx := context.Background()
	// Create a new conversation
	resp, err := cozeCli.Conversations.Create(ctx, &coze.CreateConversationsReq{BotID: botID})
	if err != nil {
		fmt.Println("Error creating conversation:", err)
		return
	}
	fmt.Println("create conversations:", resp.Conversation)
	fmt.Println(resp.LogID())

	conversationID := resp.Conversation.ID

	// Retrieve the conversation
	getResp, err := cozeCli.Conversations.Retrieve(ctx, &coze.RetrieveConversationsReq{ConversationID: conversationID})
	if err != nil {
		fmt.Println("Error retrieving conversation:", err)
		return
	}
	fmt.Println("retrieve conversations:", getResp)
	fmt.Println(getResp.LogID())

	// you can manually create message for conversation
	createMessageReq := &coze.CreateMessageReq{}
	createMessageReq.Role = coze.MessageRoleAssistant
	createMessageReq.ConversationID = conversationID
	createMessageReq.SetObjectContext([]*coze.MessageObjectString{
		coze.NewFileMessageObjectByURL(os.Getenv("IMAGE_FILE_PATH")),
		coze.NewTextMessageObject("hello"),
		coze.NewImageMessageObjectByURL(os.Getenv("IMAGE_FILE_PATH")),
	})
	time.Sleep(time.Second)

	msgs, err := cozeCli.Conversations.Messages.Create(ctx, createMessageReq)
	if err != nil {
		fmt.Println("Error creating message:", err)
		return
	}
	fmt.Println(msgs)
	fmt.Println(msgs.LogID())

	// Clear the conversation
	clearResp, err := cozeCli.Conversations.Clear(ctx, &coze.ClearConversationsReq{ConversationID: conversationID})
	if err != nil {
		fmt.Println("Error clearing conversation:", err)
		return
	}
	fmt.Println(clearResp)
	fmt.Println(clearResp.LogID())
}

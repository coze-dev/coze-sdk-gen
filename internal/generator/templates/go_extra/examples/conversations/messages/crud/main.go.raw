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
	conversationID := os.Getenv("CONVERSATION_ID")

	//
	// create message to specific conversation

	createReq := &coze.CreateMessageReq{
		ConversationID: conversationID,
		Role:           coze.MessageRoleUser,
		Content:        "message count",
		ContentType:    coze.MessageContentTypeText,
	}
	messageResp, err := cozeCli.Conversations.Messages.Create(ctx, createReq)
	if err != nil {
		fmt.Println("Error creating message:", err)
		return
	}
	message := messageResp.Message
	fmt.Println(message)
	fmt.Println(messageResp.LogID())

	//
	// retrieve message

	retrievedMsgResp, err := cozeCli.Conversations.Messages.Retrieve(ctx, &coze.RetrieveConversationsMessagesReq{
		ConversationID: conversationID,
		MessageID:      message.ID,
	})
	if err != nil {
		fmt.Println("Error retrieving message:", err)
		return
	}
	retrievedMsg := retrievedMsgResp.Message
	fmt.Println(retrievedMsg)
	fmt.Println(retrievedMsgResp.LogID())

	//
	// update message

	updateReq := &coze.UpdateConversationMessagesReq{
		ConversationID: conversationID,
		MessageID:      message.ID,
		Content:        fmt.Sprintf("modified message content:%s", message.Content),
		ContentType:    coze.MessageContentTypeText,
	}
	updateResp, err := cozeCli.Conversations.Messages.Update(ctx, updateReq)
	if err != nil {
		fmt.Println("Error updating message:", err)
		return
	}
	updatedMsg := updateResp.Message
	fmt.Println(updatedMsg)
	fmt.Println(updateResp.LogID())

	//
	// delete message

	deletedMsgResp, err := cozeCli.Conversations.Messages.Delete(ctx, &coze.DeleteConversationsMessagesReq{
		ConversationID: conversationID,
		MessageID:      message.ID,
	})
	if err != nil {
		fmt.Println("Error deleting message:", err)
		return
	}
	deletedMsg := deletedMsgResp.Message
	fmt.Println(deletedMsg)
	fmt.Println(deletedMsgResp.LogID())
}

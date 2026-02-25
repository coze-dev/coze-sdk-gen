package main

import (
	"context"
	"fmt"
	"os"

	"github.com/coze-dev/coze-go"
)

// This examples describes how to use the workflow interface to chats.
func main() {
	// Get an access_token through personal access token or oauth.
	token := os.Getenv("COZE_API_TOKEN")
	authCli := coze.NewTokenAuth(token)

	// Init the Coze client through the access_token.
	cozeCli := coze.NewCozeAPI(authCli, coze.WithBaseURL(os.Getenv("COZE_API_BASE")))

	ctx := context.Background()
	workflowID := os.Getenv("WORKFLOW_ID")

	// if your workflow need input params, you can send them by map
	data := map[string]interface{}{
		"date": "param values",
	}

	req := &coze.RunWorkflowsReq{
		WorkflowID: workflowID,
		Parameters: data,
		IsAsync:    true,
	}

	resp, err := cozeCli.Workflows.Runs.Create(ctx, req)
	if err != nil {
		fmt.Println("Error running workflow:", err)
		return
	}
	fmt.Println(resp)
	fmt.Println(resp.LogID())
}

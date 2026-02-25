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
		IsAsync:    true, // if you want the workflow runs asynchronously, you must set isAsync to true.
	}

	resp, err := cozeCli.Workflows.Runs.Create(ctx, req)
	if err != nil {
		fmt.Println("Error running workflow:", err)
		return
	}
	fmt.Println("Start async workflow runs:", resp.ExecuteID)
	fmt.Println(resp.LogID())

	executeID := resp.ExecuteID
	isFinished := false

	for !isFinished {
		historyResp, err := cozeCli.Workflows.Runs.Histories.Retrieve(ctx, &coze.RetrieveWorkflowsRunsHistoriesReq{
			WorkflowID: workflowID,
			ExecuteID:  executeID,
		})
		if err != nil {
			fmt.Println("Error retrieving history:", err)
			return
		}
		fmt.Println(historyResp)
		fmt.Println(historyResp.LogID())

		history := historyResp.Histories[0]
		switch history.ExecuteStatus {
		case coze.WorkflowExecuteStatusFail:
			fmt.Println("Workflow runs failed, reason:", history.ErrorMessage)
			isFinished = true
		case coze.WorkflowExecuteStatusRunning:
			fmt.Println("Workflow runs is running")
			time.Sleep(time.Second)
		default:
			fmt.Println("Workflow runs success:", history.Output)
			isFinished = true
		}
	}
}

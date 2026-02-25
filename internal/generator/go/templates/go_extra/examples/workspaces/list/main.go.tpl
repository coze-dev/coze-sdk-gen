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
	workspaces, err := cozeCli.Workspaces.List(ctx, &coze.ListWorkspaceReq{PageSize: 2})
	if err != nil {
		fmt.Println("Error fetching workspaces:", err)
		return
	}
	for workspaces.Next() {
		fmt.Println(workspaces.Current())
	}
	if workspaces.Err() != nil {
		fmt.Println("Error fetching workspaces:", workspaces.Err())
		return
	}

	// the page result will return followed information
	fmt.Println("total:", workspaces.Total())
	fmt.Println("has_more:", workspaces.HasMore())
}

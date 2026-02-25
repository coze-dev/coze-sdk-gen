package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/coze-dev/coze-go"
)

func main() {
	// Get an access_token through personal access token or oauth.
	token := os.Getenv("COZE_API_TOKEN")
	authCli := coze.NewTokenAuth(token)
	datasetID, _ := strconv.ParseInt(os.Getenv("DATASET_ID"), 10, 64)

	// Init the Coze client through the access_token.
	cozeCli := coze.NewCozeAPI(authCli, coze.WithBaseURL(os.Getenv("COZE_API_BASE")))

	ctx := context.Background()
	// you can use iterator to automatically retrieve next page
	documents, err := cozeCli.Datasets.Documents.List(ctx, &coze.ListDatasetsDocumentsReq{Size: 2, DatasetID: datasetID})
	if err != nil {
		fmt.Println("Error fetching documents:", err)
		return
	}
	for documents.Next() {
		fmt.Println(documents.Current())
	}
	if documents.Err() != nil {
		fmt.Println("Error fetching documents:", documents.Err())
		return
	}

	// the page result will return followed information
	fmt.Println("has_more:", documents.HasMore())
}

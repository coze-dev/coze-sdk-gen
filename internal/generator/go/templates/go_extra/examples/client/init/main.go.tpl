package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/coze-dev/coze-go"
)

// This examples demonstrates how to initialize the Coze client with different configurations.
func main() {
	// Get an access_token through personal access token or oauth.
	token := os.Getenv("COZE_API_TOKEN")
	authCli := coze.NewTokenAuth(token)

	// 1. Initialize with default configuration
	cozeCli1 := coze.NewCozeAPI(authCli)
	fmt.Println("client 1:", cozeCli1)

	// 2. Initialize with custom base URL
	cozeAPIBase := os.Getenv("COZE_API_BASE")
	cozeCli2 := coze.NewCozeAPI(authCli, coze.WithBaseURL(cozeAPIBase))
	fmt.Println("client 2:", cozeCli2)

	// 3. Initialize with custom HTTP client
	customClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
	}
	cozeCli3 := coze.NewCozeAPI(authCli,
		coze.WithBaseURL(cozeAPIBase),
		coze.WithHttpClient(customClient),
	)
	fmt.Println("client 3:", cozeCli3)
}

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/coze-dev/coze-go"
)

// How to effectuate OpenAPI authorization through the OAuth authorization code method.
//
// Firstly, users need to access https://www.coze.com/open/oauth/apps. For the cn environment,
// users need to access https://www.coze.cn/open/oauth/apps to create an OAuth App of the type
// of Web application.
// The specific creation process can be referred to in the document:
// https://www.coze.com/docs/developer_guides/oauth_code. For the cn environment, it can be
// accessed at https://www.coze.cn/docs/developer_guides/oauth_code.
// After the creation is completed, the client ID, client secret, and redirect link, can be
// obtained. For the client secret, users need to keep it securely to avoid leakage.

func main() {
	redirectURI := os.Getenv("COZE_WEB_OAUTH_REDIRECT_URI")
	clientSecret := os.Getenv("COZE_WEB_OAUTH_CLIENT_SECRET")
	clientID := os.Getenv("COZE_WEB_OAUTH_CLIENT_ID")
	ctx := context.Background()

	// The default access is api.coze.com, but if you need to access api.coze.cn,
	// please use base_url to configure the api endpoint to access
	cozeAPIBase := os.Getenv("COZE_API_BASE")
	if cozeAPIBase == "" {
		cozeAPIBase = coze.ComBaseURL
	}

	// The sdk offers the WebOAuthClient class to establish an authorization for Web OAuth.
	// Firstly, it is required to initialize the WebOAuthApp with the client ID and client secret.
	oauth, err := coze.NewWebOAuthClient(clientID, clientSecret, coze.WithAuthBaseURL(cozeAPIBase))
	if err != nil {
		fmt.Printf("Failed to create OAuth client: %v\n", err)
		return
	}

	// Generate the authorization link and direct the user to open it.
	oauthURL := oauth.GetOAuthURL(ctx, &coze.GetWebOAuthURLReq{
		RedirectURI: redirectURI,
		State:       "state",
	})
	fmt.Println(oauthURL)

	// The space permissions for which the Access Token is granted can be specified. As following codes:
	// oauthURL = oauth.GetOAuthURL(&coze.GetWebOAuthURLReq{
	// 	RedirectURI: redirectURI,
	// 	State:       "state",
	// 	WorkspaceID: &workspaceID,
	// })
	// fmt.Println(oauthURL)

	// After the user clicks the authorization consent button, the coze web page will redirect
	// to the redirect address configured in the authorization link and carry the authorization
	// code and state parameters in the address via the query string.
	//
	// Get from the query of the redirect interface: query.get('code')
	code := "mock code"

	// After obtaining the code after redirection, the interface to exchange the code for a
	// token can be invoked to generate the coze access_token of the authorized user.
	resp, err := oauth.GetAccessToken(ctx, &coze.GetWebOAuthAccessTokenReq{
		Code:        code,
		RedirectURI: redirectURI,
	})
	if err != nil {
		fmt.Printf("Failed to get access token: %v\n", err)
		return
	}
	fmt.Println(resp)

	// When the token expires, you can also refresh and re-obtain the token
	resp, err = oauth.RefreshToken(ctx, resp.RefreshToken)
	if err != nil {
		fmt.Printf("Failed to refresh token: %v\n", err)
		return
	}

	fmt.Printf("%+v\n", resp)

	// you can get request log by getLogID method
	fmt.Println(resp.LogID())

	// use the access token to init Coze client
	cozeCli := coze.NewCozeAPI(coze.NewTokenAuth(resp.AccessToken))
	_ = cozeCli
}

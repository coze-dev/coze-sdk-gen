package main

import (
	"context"
	"fmt"
	"os"

	"github.com/coze-dev/coze-go"
)

// How to effectuate OpenAPI authorization through the OAuth Proof Key for Code Exchange method.
//
// PKCE stands for Proof Key for Code Exchange, and it's an extension to the OAuth 2.0 authorization
// code flow designed to enhance security for public clients, such as mobile and single-page
// applications.
//
// Firstly, users need to access https://www.coze.com/open/oauth/apps. For the cn environment,
// users need to access https://www.coze.cn/open/oauth/apps to create an OAuth App of the type
// of Mobile/PC/Single-page application.
// The specific creation process can be referred to in the document:
// https://www.coze.com/docs/developer_guides/oauth_pkce. For the cn environment, it can be
// accessed at https://www.coze.cn/docs/developer_guides/oauth_pkce.
// After the creation is completed, the client ID can be obtained.

func main() {
	redirectURI := os.Getenv("COZE_PKCE_OAUTH_REDIRECT_URI")
	clientID := os.Getenv("COZE_PKCE_OAUTH_CLIENT_ID")

	//
	// The default access is api.coze.com, but if you need to access api.coze.cn,
	// please use base_url to configure the api endpoint to access

	cozeAPIBase := os.Getenv("COZE_API_BASE")
	if cozeAPIBase == "" {
		cozeAPIBase = coze.ComBaseURL
	}
	ctx := context.Background()

	oauth, err := coze.NewPKCEOAuthClient(clientID, coze.WithAuthBaseURL(cozeAPIBase))
	if err != nil {
		fmt.Printf("Failed to create OAuth client: %v\n", err)
		return
	}

	// In the SDK, we have wrapped up the code_challenge process of PKCE.
	// Developers only need to select the code_challenge_method.
	oauthURL, err := oauth.GetOAuthURL(ctx, &coze.GetPKCEOAuthURLReq{
		RedirectURI: redirectURI,
		State:       "state",
		Method:      coze.CodeChallengeMethodS256.Ptr(),
	})
	if err != nil {
		fmt.Printf("Failed to generate OAuth URL: %v\n", err)
		return
	}
	fmt.Println(oauthURL.AuthorizationURL)
	fmt.Println(oauthURL.CodeVerifier)

	//
	// The space permissions for which the Access Token is granted can be specified. As following codes:
	// oauthURL, err := oauth.GenOAuthURL(&coze.GetPKCEOAuthURLReq{
	//			RedirectURI: redirectURI, State: "state",
	//			Method: coze.CodeChallengeMethodS256.Ptr(),
	//			WorkspaceID: utils.ptr("workspace_id"),
	//		})
	// if err != nil {
	// 	fmt.Printf("Failed to generate OAuth URL with workspaces: %v\n", err)
	// 	return
	// }
	// fmt.Println(oauthURL.AuthorizationURL)
	//	fmt.Println(oauthURL.CodeVerifier)

	// After the user clicks the authorization consent button,
	// the coze web page will redirect to the redirect address configured in the authorization link
	// and carry the authorization code and state parameters in the address via the query string.
	// Get from the query of the redirect interface : query.get('code')
	code := "mock code"
	codeVerifier := oauthURL.CodeVerifier
	// After obtaining the code after redirection, the interface to exchange the code for a
	// token can be invoked to generate the coze access_token of the authorized user.
	// The developer should use code verifier returned by genOAuthURL() method
	resp, err := oauth.GetAccessToken(ctx, &coze.GetPKCEAccessTokenReq{
		Code:         code,
		RedirectURI:  redirectURI,
		CodeVerifier: codeVerifier,
	})
	if err != nil {
		fmt.Printf("Failed to get access token: %v\n", err)
		return
	}
	fmt.Printf("%+v\n", resp)

	// use the access token to init Coze client
	cozeCli := coze.NewCozeAPI(coze.NewTokenAuth(resp.AccessToken), coze.WithBaseURL(cozeAPIBase))
	fmt.Println(cozeCli)

	// When the token expires, you can also refresh and re-obtain the token
	resp, err = oauth.RefreshToken(ctx, resp.RefreshToken)
	if err != nil {
		fmt.Printf("Failed to refresh token: %v\n", err)
		return
	}
	fmt.Println(resp)
}

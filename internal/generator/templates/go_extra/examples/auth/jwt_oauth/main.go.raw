package main

import (
	"context"
	"fmt"
	"os"

	"github.com/coze-dev/coze-go"
)

//
// This examples is about how to use the service jwt oauth process to acquire user authorization.
//
// Firstly, users need to access https://www.coze.com/open/oauth/apps. For the cn environment,
// users need to access https://www.coze.cn/open/oauth/apps to create an OAuth App of the type
// of Service application.
// The specific creation process can be referred to in the document:
// https://www.coze.com/docs/developer_guides/oauth_jwt. For the cn environment, it can be
// accessed at https://www.coze.cn/docs/developer_guides/oauth_jwt.
// After the creation is completed, the client ID, private key, and public key id, can be obtained.
// For the client secret and public key id, users need to keep it securely to avoid leakage.
//

func main() {
	// The default access is api.coze.com, but if you need to access api.coze.cn,
	// please use base_url to configure the api endpoint to access
	cozeAPIBase := os.Getenv("COZE_API_BASE")
	jwtOauthClientID := os.Getenv("COZE_JWT_OAUTH_CLIENT_ID")
	jwtOauthPrivateKey := os.Getenv("COZE_JWT_OAUTH_PRIVATE_KEY")
	jwtOauthPrivateKeyFilePath := os.Getenv("COZE_JWT_OAUTH_PRIVATE_KEY_FILE_PATH")
	jwtOauthPublicKeyID := os.Getenv("COZE_JWT_OAUTH_PUBLIC_KEY_ID")

	// Read private key from file
	privateKeyBytes, err := os.ReadFile(jwtOauthPrivateKeyFilePath)
	if err != nil {
		fmt.Printf("Error reading private key file: %v\n", err)
		return
	}
	jwtOauthPrivateKey = string(privateKeyBytes)

	// The jwt oauth type requires using private to be able to issue a jwt token, and through the jwt token,
	// apply for an access_token from the coze service.The sdk encapsulates this procedure,
	// and only needs to use get_access_token to obtain the access_token under the jwt oauth process.
	// Generate the authorization token The default ttl is 900s, and developers can customize the expiration time,
	// which can be set up to 24 hours at most.
	oauth, err := coze.NewJWTOAuthClient(coze.NewJWTOAuthClientParam{
		ClientID: jwtOauthClientID, PublicKey: jwtOauthPublicKeyID, PrivateKeyPEM: jwtOauthPrivateKey,
	}, coze.WithAuthBaseURL(cozeAPIBase))
	if err != nil {
		fmt.Printf("Error creating JWT OAuth client: %v\n", err)
		return
	}
	ctx := context.Background()

	resp, err := oauth.GetAccessToken(ctx, nil)
	if err != nil {
		fmt.Printf("Error getting access token: %v\n", err)
		return
	}
	fmt.Printf("Access token response: %+v\n", resp)
	fmt.Println(resp.LogID())

	// The jwt oauth process does not support refreshing tokens. When the token expires,
	// just directly call get_access_token to generate a new token.
	cozeCli := coze.NewCozeAPI(
		coze.NewJWTAuth(oauth, nil),
		coze.WithBaseURL(cozeAPIBase),
	)
	_ = cozeCli
}

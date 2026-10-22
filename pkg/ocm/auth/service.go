// Copyright Contributors to the Open Cluster Management project

package auth

import (
	"fmt"
)

var authBaseURL = "https://sso.redhat.com"

type authClient struct{}

type TokenGetter interface {
	GetToken(AuthRequest) (string, error)
}

var (
	AuthClient TokenGetter = &authClient{}
)

func (client authClient) GetToken(request AuthRequest) (string, error) {
	if request.BaseURL == "" {
		request.BaseURL = authBaseURL
	}
	response, err := AuthProvider.GetToken(request)

	if err != nil {
		return "", fmt.Errorf("%s: %v", "couldn't get token", err)
	}

	if response.AccessToken == "" {
		return "", fmt.Errorf("missing `access_token` in response")
	}
	return response.AccessToken, nil
}

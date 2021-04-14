// Copyright Contributors to the Open Cluster Management project

package auth

import (
	"errors"
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
		if errors.Is(err.Error, ErrInvalidToken) {
			// something wasn't found
		}
		return "", fmt.Errorf("%s: %w", "couldn't get token", err.Error)
	}
	if response.AccessToken == "" {
		return "", fmt.Errorf("missing `access_token` in response")
	}
	return response.AccessToken, nil
}

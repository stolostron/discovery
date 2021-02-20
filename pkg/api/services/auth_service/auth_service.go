// Copyright Contributors to the Open Cluster Management project

package auth_service

import (
	"fmt"

	"github.com/open-cluster-management/discovery/pkg/api/domain/auth_domain"
	"github.com/open-cluster-management/discovery/pkg/api/providers/auth_provider"
)

var authBaseURL = "https://sso.redhat.com"

type authClient struct{}

type TokenGetter interface {
	GetToken(auth_domain.AuthRequest) (string, error)
}

var (
	AuthClient TokenGetter = &authClient{}
)

func (client authClient) GetToken(request auth_domain.AuthRequest) (string, error) {
	if request.BaseURL == "" {
		request.BaseURL = authBaseURL
	}
	response, err := auth_provider.AuthProvider.GetToken(request)

	if err != nil {
		return "", fmt.Errorf("Couldn't get token: %s", err.Description)
	}
	if response.AccessToken == "" {
		return "", fmt.Errorf("missing `access_token` in response")
	}
	return response.AccessToken, nil
}

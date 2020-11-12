package auth_service

import (
	"fmt"
	"os"

	"github.com/open-cluster-management/discovery/pkg/api/domain/auth_domain"
	"github.com/open-cluster-management/discovery/pkg/api/providers/auth_provider"
)

var authBaseURL = "https://sso.redhat.com"

func init() {
	if val, ok := os.LookupEnv("OCM_URL"); ok {
		authBaseURL = val
	}
}

type authClient struct{}

type TokenGetter interface {
	GetToken(string) (string, error)
}

var (
	AuthClient TokenGetter = &authClient{}
)

func (client authClient) GetToken(token string) (string, error) {
	response, err := auth_provider.AuthProvider.GetToken(auth_domain.AuthRequest{
		Token:   token,
		BaseURL: authBaseURL,
	})
	if err != nil {
		return "", fmt.Errorf("Couldn't get token: %s", err.Description)
	}
	if response.AccessToken == "" {
		return "", fmt.Errorf("missing `access_token` in response")
	}
	return response.AccessToken, nil
}

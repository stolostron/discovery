package auth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// ErrorResponse ...
type ErrorResponse struct {
	Kind   string `yaml:"kind,omitempty"`
	ID     string `yaml:"id,omitempty"`
	Href   string `yaml:"href,omitempty"`
	Code   string `yaml:"code,omitempty"`
	Reason string `yaml:"reason,omitempty"`
}

// AuthTokenResponse ...
type AuthTokenResponse struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	IDToken          string `json:"id_token"`
	NotBeforePolicy  int    `json:"not-before-policy"`
	SessionState     string `json:"session_state"`
	Scope            string `json:"scope"`
}

func NewAccessToken(ocmToken string) (string, error) {
	requestURL := "https://sso.redhat.com/auth/realms/redhat-external/protocol/openid-connect/token"

	data := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {"cloud-services"},
		"refresh_token": {ocmToken},
	}
	resp, err := http.PostForm(requestURL, data)
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	authTokenResponse := AuthTokenResponse{}
	err = json.Unmarshal(body, &authTokenResponse)
	if err != nil {
		return "", err
	}
	if authTokenResponse.AccessToken == "" {
		return "", fmt.Errorf("Error reading access token from response: %v", string(body))
	}

	return authTokenResponse.AccessToken, nil
}

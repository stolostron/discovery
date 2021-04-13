// Copyright Contributors to the Open Cluster Management project

package auth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

const (
	authEndpoint = "%s/auth/realms/redhat-external/protocol/openid-connect/token"
)

var (
	httpClient   AuthPostInterface = &authRestClient{}
	AuthProvider IAuthProvider     = &authProvider{}
)

type AuthPostInterface interface {
	Post(url string, data url.Values) (resp *http.Response, err error)
}

type authRestClient struct{}

func (c *authRestClient) Post(url string, data url.Values) (resp *http.Response, err error) {
	return http.PostForm(url, data) // #nosec G107 (url needs to be configurable to target mock servers)
}

type IAuthProvider interface {
	GetToken(request AuthRequest) (*AuthTokenResponse, *AuthError)
}

type authProvider struct{}

func (a *authProvider) GetToken(request AuthRequest) (*AuthTokenResponse, *AuthError) {
	postUrl := fmt.Sprintf(authEndpoint, request.BaseURL)
	data := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {"cloud-services"},
		"refresh_token": {request.Token},
	}

	response, err := httpClient.Post(postUrl, data)
	if err != nil {
		return nil, &AuthError{
			Error: err,
		}
	}
	defer response.Body.Close()

	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, &AuthError{
			Error: err,
		}
	}

	// The api owner can decide to change datatypes, etc. When this happen, it might affect the error format returned
	if response.StatusCode > 299 {
		var errResponse AuthError
		if err := json.Unmarshal(bytes, &errResponse); err != nil {
			return nil, &AuthError{
				Error:    err,
				Response: bytes,
				Code:     response.StatusCode,
			}
		}
		errResponse.Code = response.StatusCode

		if errResponse.ErrorMessage == "" || errResponse.Description == "" {
			errResponse.Error = fmt.Errorf("invalid json response body")
			errResponse.Response = bytes
		}
		return nil, &errResponse
	}

	var result AuthTokenResponse
	if err := json.Unmarshal(bytes, &result); err != nil {
		return nil, &AuthError{
			Error:    fmt.Errorf("error unmarshaling response"),
			Response: bytes,
			Code:     response.StatusCode,
		}
	}

	return &result, nil
}

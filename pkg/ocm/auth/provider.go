// Copyright Contributors to the Open Cluster Management project

package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	authEndpoint = "%s/auth/realms/redhat-external/protocol/openid-connect/token"
)

var logf = log.Log.WithName("auth-provider")

var (
	httpClient   AuthPostInterface = &authRestClient{}
	AuthProvider IAuthProvider     = &authProvider{}

	ErrInvalidToken       = errors.New("invalid token")
	ErrInvalidClient      = errors.New("invalid_client")
	ErrUnauthorizedClient = errors.New("unauthorized_client")
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

func (a *authProvider) GetToken(request AuthRequest) (retRes *AuthTokenResponse, retErr *AuthError) {
	postUrl := fmt.Sprintf(authEndpoint, request.BaseURL)

	var data url.Values
	switch request.AuthMethod {
	case "service-account":
		data = url.Values{
			"grant_type":    {"client_credentials"},
			"client_id":     {request.ID},
			"client_secret": {request.Secret},
		}

	default:
		data = url.Values{
			"grant_type":    {"refresh_token"},
			"client_id":     {"cloud-services"},
			"refresh_token": {request.Token},
		}
	}

	response, err := httpClient.Post(postUrl, data)
	if err != nil {
		return nil, &AuthError{
			Error: err,
		}
	}

	defer func() {
		err := response.Body.Close()
		if err != nil && retErr == nil {
			retErr = &AuthError{
				Error: fmt.Errorf("%s: %w", "error closing response body", err),
			}
		}
	}()

	retRes, retErr = parseResponse(response)
	return
}

func parseResponse(response *http.Response) (*AuthTokenResponse, *AuthError) {
	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, &AuthError{
			Error: err,
		}
	}

	if response.StatusCode > 299 {
		var errResponse AuthError
		if err := json.Unmarshal(bytes, &errResponse); err != nil {
			logf.V(1).Info("Auth API error response", "status", response.StatusCode, "body", string(bytes))
			return nil, &AuthError{
				Error:    fmt.Errorf("authentication failed"),
				Response: bytes,
				Code:     response.StatusCode,
			}
		}
		errResponse.Code = response.StatusCode

		logf.V(1).Info("Auth API error", "status", response.StatusCode, "error", errResponse.ErrorMessage, "description", errResponse.Description)
		errResponse.Response = bytes

		if errResponse.ErrorMessage == "" || errResponse.Description == "" {
			errResponse.Error = fmt.Errorf("authentication failed")
		} else if errResponse.Description == "Invalid refresh token" {
			errResponse.Error = ErrInvalidToken
		} else {
			errResponse.Error = fmt.Errorf("authentication failed")
		}

		return nil, &errResponse
	}

	var result AuthTokenResponse
	if err := json.Unmarshal(bytes, &result); err != nil {
		logf.V(1).Info("Failed to unmarshal auth response", "status", response.StatusCode, "body", string(bytes))
		return nil, &AuthError{
			Error:    fmt.Errorf("failed to parse authentication response"),
			Response: bytes,
			Code:     response.StatusCode,
		}
	}

	return &result, nil
}

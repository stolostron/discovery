package auth_provider

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"

	"github.com/open-cluster-management/discovery/pkg/api/clients/restclient"
	"github.com/open-cluster-management/discovery/pkg/api/domain/auth_domain"
)

const (
	// authURL      = "https://sso.redhat.com/auth/realms/redhat-external/protocol/openid-connect/token"
	authEndpoint = "%s/auth/realms/redhat-external/protocol/openid-connect/token"
)

type authProvider struct{}

type IAuthProvider interface {
	GetToken(request auth_domain.AuthRequest) (*auth_domain.AuthTokenResponse, *auth_domain.AuthError)
}

var (
	AuthProvider IAuthProvider = &authProvider{}
)

func (a *authProvider) GetToken(request auth_domain.AuthRequest) (*auth_domain.AuthTokenResponse, *auth_domain.AuthError) {
	postUrl := fmt.Sprintf(authEndpoint, request.BaseURL)
	data := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {"cloud-services"},
		"refresh_token": {request.Token},
	}

	response, err := restclient.AuthHTTPClient.Post(postUrl, data)
	if err != nil {
		return nil, &auth_domain.AuthError{
			Error: err,
		}
	}
	defer response.Body.Close()

	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, &auth_domain.AuthError{
			Error: err,
		}
	}

	// The api owner can decide to change datatypes, etc. When this happen, it might affect the error format returned
	if response.StatusCode > 299 {
		var errResponse auth_domain.AuthError
		if err := json.Unmarshal(bytes, &errResponse); err != nil {
			return nil, &auth_domain.AuthError{
				Error:    err,
				Response: bytes,
			}
		}
		errResponse.Code = response.StatusCode

		if errResponse.ErrorMessage == "" || errResponse.Description == "" {
			errResponse.Error = fmt.Errorf("invalid json response body")
			errResponse.Response = bytes
		}
		return nil, &errResponse
	}

	var result auth_domain.AuthTokenResponse
	if err := json.Unmarshal(bytes, &result); err != nil {
		return nil, &auth_domain.AuthError{
			Error:    fmt.Errorf("error unmarshaling response"),
			Response: bytes,
		}
	}

	return &result, nil
}

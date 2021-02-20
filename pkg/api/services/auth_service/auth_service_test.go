// Copyright Contributors to the Open Cluster Management project

package auth_service

import (
	"fmt"
	"testing"

	"github.com/open-cluster-management/discovery/pkg/api/domain/auth_domain"
	"github.com/open-cluster-management/discovery/pkg/api/providers/auth_provider"
	"github.com/stretchr/testify/assert"
)

var (
	getTokenFunc func(request auth_domain.AuthRequest) (*auth_domain.AuthTokenResponse, *auth_domain.AuthError)
)

// Mocking the TokenGetter interface
type authProviderMock struct{}

func (cm *authProviderMock) GetToken(request auth_domain.AuthRequest) (*auth_domain.AuthTokenResponse, *auth_domain.AuthError) {
	return getTokenFunc(request)
}

//When the everything is good
func TestGetTokenNoError(t *testing.T) {
	getTokenFunc = func(request auth_domain.AuthRequest) (*auth_domain.AuthTokenResponse, *auth_domain.AuthError) {
		return &auth_domain.AuthTokenResponse{
			AccessToken: "new_access_token",
		}, nil
	}
	auth_provider.AuthProvider = &authProviderMock{} //without this line, the real api is fired

	response, err := AuthClient.GetToken(auth_domain.AuthRequest{
		Token: "this_is_my_token",
	})
	assert.Nil(t, err)
	assert.EqualValues(t, "new_access_token", response)
}

// recieved an AuthTokenResponse but it's missing and `access_token`
func TestGetTokenMissingAccessToken(t *testing.T) {
	getTokenFunc = func(request auth_domain.AuthRequest) (*auth_domain.AuthTokenResponse, *auth_domain.AuthError) {
		return &auth_domain.AuthTokenResponse{}, nil
	}
	auth_provider.AuthProvider = &authProviderMock{} //without this line, the real api is fired

	response, err := AuthClient.GetToken(auth_domain.AuthRequest{
		Token: "this_is_my_token",
	})
	assert.NotNil(t, err)
	assert.EqualValues(t, "", response)
}

// recieved an error caused by unmarshalling rather than from the API
func TestGetTokenInvalidErrorResponse(t *testing.T) {
	getTokenFunc = func(request auth_domain.AuthRequest) (*auth_domain.AuthTokenResponse, *auth_domain.AuthError) {
		return nil, &auth_domain.AuthError{
			Error:    fmt.Errorf("invalid json response body"),
			Response: []byte(`{"code": 405, "message":"RESTEASY003650: No resource method found for GET, return 405 with Allow header"}`),
		}
	}
	auth_provider.AuthProvider = &authProviderMock{} //without this line, the real api is fired

	response, err := AuthClient.GetToken(auth_domain.AuthRequest{
		Token: "this_is_my_token",
	})
	assert.NotNil(t, err)
	assert.EqualValues(t, "", response)
}

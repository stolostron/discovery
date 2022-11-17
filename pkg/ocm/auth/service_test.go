// Copyright Contributors to the Open Cluster Management project

package auth

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	getTokenFunc func(request AuthRequest) (*AuthTokenResponse, *AuthError)
)

// Mocking the TokenGetter interface
type authProviderMock struct{}

func (cm *authProviderMock) GetToken(request AuthRequest) (*AuthTokenResponse, *AuthError) {
	return getTokenFunc(request)
}

// When the everything is good
func TestProviderGetTokenNoError(t *testing.T) {
	getTokenFunc = func(request AuthRequest) (*AuthTokenResponse, *AuthError) {
		return &AuthTokenResponse{
			AccessToken: "new_access_token",
		}, nil
	}
	AuthProvider = &authProviderMock{} //without this line, the real api is fired

	response, err := AuthClient.GetToken(AuthRequest{
		Token: "this_is_my_token",
	})
	assert.Nil(t, err)
	assert.EqualValues(t, "new_access_token", response)
}

// recieved an AuthTokenResponse but it's missing and `access_token`
func TestGetTokenMissingAccessToken(t *testing.T) {
	getTokenFunc = func(request AuthRequest) (*AuthTokenResponse, *AuthError) {
		return &AuthTokenResponse{}, nil
	}
	AuthProvider = &authProviderMock{} //without this line, the real api is fired

	response, err := AuthClient.GetToken(AuthRequest{
		Token: "this_is_my_token",
	})
	assert.NotNil(t, err)
	assert.EqualValues(t, "", response)
}

// recieved an error caused by unmarshalling rather than from the API
func TestGetTokenInvalidErrorResponse(t *testing.T) {
	getTokenFunc = func(request AuthRequest) (*AuthTokenResponse, *AuthError) {
		return nil, &AuthError{
			Error:    fmt.Errorf("invalid json response body"),
			Response: []byte(`{"code": 405, "message":"RESTEASY003650: No resource method found for GET, return 405 with Allow header"}`),
		}
	}
	AuthProvider = &authProviderMock{} //without this line, the real api is fired

	response, err := AuthClient.GetToken(AuthRequest{
		Token: "this_is_my_token",
	})
	assert.NotNil(t, err)
	assert.EqualValues(t, "", response)
}

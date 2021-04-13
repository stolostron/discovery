// Copyright Contributors to the Open Cluster Management project

package auth

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	postRequestFunc func(url string, data url.Values) (*http.Response, error)
)

// Mocking the AuthPostInterface
type postClientMock struct{}

func (cm *postClientMock) Post(request string, data url.Values) (*http.Response, error) {
	return postRequestFunc(request, data)
}

//When the everything is good
func TestGetTokenNoError(t *testing.T) {
	postRequestFunc = func(url string, data url.Values) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader(`{"access_token":"ephemeral_access_token","not-before-policy":0,"session_state":"random-session-state","scope":"openid offline_access"}`)),
		}, nil
	}
	httpClient = &postClientMock{} //without this line, the real api is fired

	response, err := AuthProvider.GetToken(AuthRequest{Token: "this_is_my_token"})
	assert.NotNil(t, response)
	assert.Nil(t, err)
	assert.EqualValues(t, "ephemeral_access_token", response.AccessToken)
}

func TestGetTokenInvalidApiKey(t *testing.T) {
	t.Run("Bad token", func(t *testing.T) {
		postRequestFunc = func(url string, data url.Values) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       ioutil.NopCloser(strings.NewReader(`{"error":"invalid_grant","error_description":"Invalid refresh token"}`)),
			}, nil
		}
		httpClient = &postClientMock{} //without this line, the real api is fired

		response, err := AuthProvider.GetToken(AuthRequest{Token: "this_is_an_invalid_token"})
		assert.NotNil(t, err)
		assert.Nil(t, response)
		assert.EqualValues(t, http.StatusBadRequest, err.Code)
		assert.EqualValues(t, "invalid_grant", err.ErrorMessage)
		assert.EqualValues(t, "Invalid refresh token", err.Description)
	})

	t.Run("Empty string token", func(t *testing.T) {
		postRequestFunc = func(url string, data url.Values) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       ioutil.NopCloser(strings.NewReader(`{"error":"invalid_grant","error_description":"Invalid refresh token"}`)),
			}, nil
		}
		httpClient = &postClientMock{} //without this line, the real api is fired

		response, err := AuthProvider.GetToken(AuthRequest{Token: ""})
		assert.NotNil(t, err)
		assert.Nil(t, response)
		assert.EqualValues(t, http.StatusBadRequest, err.Code)
		assert.EqualValues(t, "invalid_grant", err.ErrorMessage)
		assert.EqualValues(t, "Invalid refresh token", err.Description)
	})
}

func TestGetTokenMissingFormData(t *testing.T) {
	t.Run("Missing grant_type", func(t *testing.T) {
		postRequestFunc = func(url string, data url.Values) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       ioutil.NopCloser(strings.NewReader(`{"error":"invalid_request","error_description":"Missing form parameter: grant_type"}`)),
			}, nil
		}
		httpClient = &postClientMock{} //without this line, the real api is fired

		response, err := AuthProvider.GetToken(AuthRequest{Token: "this_is_my_token"})
		assert.NotNil(t, err)
		assert.Nil(t, response)
		assert.EqualValues(t, http.StatusBadRequest, err.Code)
		assert.EqualValues(t, "invalid_request", err.ErrorMessage)
	})

	t.Run("Missing client_id", func(t *testing.T) {
		postRequestFunc = func(url string, data url.Values) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       ioutil.NopCloser(strings.NewReader(`{"error":"unauthorized_client","error_description":"INVALID_CREDENTIALS: Invalid client credentials"}`)),
			}, nil
		}
		httpClient = &postClientMock{} //without this line, the real api is fired

		response, err := AuthProvider.GetToken(AuthRequest{Token: "this_is_my_token"})
		assert.NotNil(t, err)
		assert.Nil(t, response)
		assert.EqualValues(t, http.StatusBadRequest, err.Code)
		assert.EqualValues(t, "unauthorized_client", err.ErrorMessage)
	})
}

//When the error response is invalid, here the code is supposed to be an integer, but a string was given.
//This can happen when the api owner changes some data types in the api
func TestGetTokenInvalidErrorInterface(t *testing.T) {
	unexpectedJSONResponse := `{"code": 405, "message":"RESTEASY003650: No resource method found for GET, return 405 with Allow header"}`
	postRequestFunc = func(url string, data url.Values) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusMethodNotAllowed,
			Body:       ioutil.NopCloser(strings.NewReader(unexpectedJSONResponse)),
		}, nil
	}
	httpClient = &postClientMock{} //without this line, the real api is fired

	response, err := AuthProvider.GetToken(AuthRequest{Token: "this_is_my_token"})
	assert.NotNil(t, err)
	assert.Nil(t, response)
	assert.NotNil(t, err.Error)
	assert.EqualValues(t, unexpectedJSONResponse, err.Response)
}

func TestGetTokenInvalidResponseInterface(t *testing.T) {
	// access_token returned is a number instead of a string
	unexpectedJSONResponse := `{"access_token":12345,"not-before-policy":0,"session_state":"random-session-state","scope":"openid offline_access"}`
	postRequestFunc = func(url string, data url.Values) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader(unexpectedJSONResponse)),
		}, nil
	}
	httpClient = &postClientMock{} //without this line, the real api is fired

	response, err := AuthProvider.GetToken(AuthRequest{Token: "this_is_my_token"})
	assert.NotNil(t, err)
	assert.Nil(t, response)
	assert.NotNil(t, err.Error)
	assert.EqualValues(t, unexpectedJSONResponse, err.Response)
}

// If the response does not indicate any error but doesn't give an access token then we may be in hot water
func TestGetTokenNoAccessToken(t *testing.T) {
	missingTokenJSONResponse := `{"not-before-policy":0,"session_state":"random-session-state","scope":"openid offline_access"}`
	postRequestFunc = func(url string, data url.Values) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader(missingTokenJSONResponse)),
		}, nil
	}
	httpClient = &postClientMock{} //without this line, the real api is fired

	response, err := AuthProvider.GetToken(AuthRequest{Token: "this_is_my_token"})
	assert.NotNil(t, response)
	assert.Nil(t, err)

}

// Copyright Contributors to the Open Cluster Management project

package common

import (
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	v1 "github.com/stolostron/discovery/api/v1"
	"github.com/stolostron/discovery/pkg/ocm/cluster"
	sub "github.com/stolostron/discovery/pkg/ocm/subscription"
	"github.com/stretchr/testify/assert"
)

var (
	getRequestFunc func(*http.Request) (*http.Response, error)
)

// Mocking the HTTPRequester interface
type getClientMock struct{}

func (cm *getClientMock) Get(request *http.Request) (*http.Response, error) {
	return getRequestFunc(request)
}

func TestProviderGetResourcesNoError(t *testing.T) {
	tests := []struct {
		name         string
		endpointURL  string
		testFilePath string
	}{
		{
			name:         "Should get cluster resources with no error",
			endpointURL:  cluster.GetClusterURL(),
			testFilePath: "../cluster/testdata/clusters_mgmt_mock.json",
		},
		{
			name:         "Should get subscription resources with no error",
			endpointURL:  sub.GetSubscriptionURL(),
			testFilePath: "../subscription/testdata/accounts_mgmt_mock.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getRequestFunc = func(*http.Request) (*http.Response, error) {
				file, err := os.Open(tt.testFilePath)
				if err != nil {
					t.Error(err)
				}

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(file),
				}, nil
			}
			httpClient = &getClientMock{} //without this line, the real api is fired

			response, err := provider.GetResources(Request{}, tt.endpointURL)
			assert.NotNil(t, response)
			assert.Nil(t, err)
			assert.EqualValues(t, 1, len(response.Items))
		})
	}
}

func TestParseResponse(t *testing.T) {
	tests := []struct {
		name     string
		response *http.Response
		want     bool
	}{
		{
			name: "Should parse response with no error",
			response: &http.Response{
				Body:       io.NopCloser(strings.NewReader(`{"kind":"example","page":1,"size":10,"total":100,"items":[{"id":1,"name":"item1"},{"id":2,"name":"item2"}]}`)),
				StatusCode: 200,
			},
			want: true,
		},
		{
			name: "Should parse response with error",
			response: &http.Response{
				Body:       io.NopCloser(strings.NewReader(`{"reason":"example reason"}`)),
				StatusCode: 400,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseResponse(tt.response)

			if tt.response.StatusCode > 299 {
				if ok := got == nil; !ok {
					t.Errorf("parseResponse(response) got (%v,  %v), want %v", got, err, tt.want)
				}

			} else {
				if ok := err == nil; !ok {
					t.Errorf("parseResponse(response) got (%v, %v), want %v", got, err, tt.want)
				}
			}
		})
	}
}

func TestPrepareRequest(t *testing.T) {
	tests := []struct {
		name        string
		endpointURL string
		request     Request
		want        bool
	}{
		{
			name:        "Should prepare request for getting cluster resources",
			endpointURL: cluster.GetClusterURL(),
			request: Request{
				BaseURL: ocmBaseURL,
				Token:   "test-token",
				Filter: v1.Filter{
					LastActive: 2,
				},
			},
			want: true,
		},
		{
			name:        "Should prepare request for getting subscription resources",
			endpointURL: sub.GetSubscriptionURL(),
			request: Request{
				BaseURL: ocmBaseURL,
				Token:   "test-token",
				Filter: v1.Filter{
					LastActive: 2,
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := prepareRequest(tt.request, tt.endpointURL); err != nil {
				t.Errorf("prepareRequest(request, endpointURL), got %v, want %v", got, tt.want)
			}
		})
	}
}

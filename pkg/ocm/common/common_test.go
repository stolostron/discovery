// Copyright Contributors to the Open Cluster Management project

package common

import (
	"io"
	"net/http"
	"os"
	"testing"

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
		// want         bool
	}{
		{
			name:         "Should get cluster resources with no error",
			endpointURL:  cluster.GetClusterURL(),
			testFilePath: "../cluster/testdata/clusters_mgmt_mock.json",
			// want:         true,
		},
		{
			name:         "Should get subscription resources with no error",
			endpointURL:  sub.GetSubscriptionURL(),
			testFilePath: "../subscription/testdata/accounts_mgmt_mock.json",
			// want:         false,
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

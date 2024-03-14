// Copyright Contributors to the Open Cluster Management project

package common

import (
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	// "github.com/stolostron/discovery/pkg/ocm/cluster"
	// discovery "github.com/stolostron/discovery/api/v1"
	// "github.com/stolostron/discovery/pkg/ocm/cluster"
)

var (
	getRequestFunc func(*http.Request) (*http.Response, error)
	endpointURL    = "http://sample"
)

// Mocking the HTTPRequester interface
type getClientMock struct{}

func (cm *getClientMock) Get(request *http.Request) (*http.Response, error) {
	return getRequestFunc(request)
}

// When the everything is good
func TestProviderGetResourcesNoError(t *testing.T) {
	getRequestFunc = func(*http.Request) (*http.Response, error) {
		file, err := os.Open("testdata/ocm_mock.json")
		if err != nil {
			t.Error(err)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(file),
		}, nil
	}

	// Create a new ClusterRequest object
	httpClient = &getClientMock{} //without this line, the real api is fired

	// Create a new instance of the Provider
	_ = Provider{}

	response, err := provider.GetResources(Request{}, endpointURL)
	assert.NotNil(t, response)
	assert.Nil(t, err)
	assert.EqualValues(t, 1, len(response.Items))
}

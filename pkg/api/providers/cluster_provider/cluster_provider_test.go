package cluster_provider

import (
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	discoveryv1 "github.com/open-cluster-management/discovery/api/v1"
	"github.com/open-cluster-management/discovery/pkg/api/clients/restclient"
	"github.com/open-cluster-management/discovery/pkg/api/domain/cluster_domain"
	"github.com/stretchr/testify/assert"
)

var (
	getRequestFunc func(*http.Request) (*http.Response, error)
)

// Mocking the ClusterGetInterface
type getClientMock struct{}

func (cm *getClientMock) Get(request *http.Request) (*http.Response, error) {
	return getRequestFunc(request)
}

//When the everything is good
func TestGetClustersNoError(t *testing.T) {
	getRequestFunc = func(*http.Request) (*http.Response, error) {
		file, err := os.Open("ocm_mock.json")
		if err != nil {
			t.Error(err)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(file),
		}, nil
	}
	restclient.ClusterHTTPClient = &getClientMock{} //without this line, the real api is fired

	response, err := ClusterProvider.GetClusters(cluster_domain.ClusterRequest{})
	assert.NotNil(t, response)
	assert.Nil(t, err)
	assert.EqualValues(t, 3, len(response.Items))
}

func TestGetClustersInvalidAccessToken(t *testing.T) {
	t.Run("Bad token", func(t *testing.T) {
		json := `{
		"kind": "Error",
		"id": "401",
		"href": "/api/clusters_mgmt/v1/errors/401",
		"code": "CLUSTERS-MGMT-401",
		"reason": "Signature of bearer token isn't valid"
		}`
		getRequestFunc = func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusUnauthorized,
				Body:       ioutil.NopCloser(strings.NewReader(json)),
			}, nil
		}
		restclient.ClusterHTTPClient = &getClientMock{} //without this line, the real api is fired

		response, err := ClusterProvider.GetClusters(cluster_domain.ClusterRequest{})
		assert.NotNil(t, err)
		assert.Nil(t, response)
		assert.NotEmpty(t, err.Reason)
		assert.EqualValues(t, "Error", err.Kind)
	})

	t.Run("Expired token", func(t *testing.T) {
		json := `{
			"kind": "Error",
			"id": "401",
			"href": "/api/clusters_mgmt/v1/errors/401",
			"code": "CLUSTERS-MGMT-401",
			"reason": "Bearer token is expired"
		}`
		getRequestFunc = func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusUnauthorized,
				Body:       ioutil.NopCloser(strings.NewReader(json)),
			}, nil
		}
		restclient.ClusterHTTPClient = &getClientMock{} //without this line, the real api is fired

		response, err := ClusterProvider.GetClusters(cluster_domain.ClusterRequest{})
		assert.NotNil(t, err)
		assert.Nil(t, response)
		assert.NotEmpty(t, err.Reason)
		assert.EqualValues(t, "Error", err.Kind)
	})

}

func TestGetClustersInvalidFilter(t *testing.T) {
	json := `{
		"kind": "Error",
		"id": "400",
		"href": "/api/clusters_mgmt/v1/errors/400",
		"code": "CLUSTERS-MGMT-400",
		"reason": "foo is not a valid field name",
		"operation_id": "1gt8hln464eqhcfga8ktc0sasf056sdk"
	}`
	getRequestFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       ioutil.NopCloser(strings.NewReader(json)),
		}, nil
	}
	restclient.ClusterHTTPClient = &getClientMock{} //without this line, the real api is fired

	response, err := ClusterProvider.GetClusters(cluster_domain.ClusterRequest{})
	assert.NotNil(t, err)
	assert.Nil(t, response)
	assert.NotEmpty(t, err.Reason)
	assert.EqualValues(t, "Error", err.Kind)
}

func TestGetClustersInvalidResponseInterface(t *testing.T) {
	// this is an auth error response
	json := `{"error":"invalid_grant","error_description":"Invalid refresh token"}`
	getRequestFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       ioutil.NopCloser(strings.NewReader(json)),
		}, nil
	}
	restclient.ClusterHTTPClient = &getClientMock{} //without this line, the real api is fired

	response, err := ClusterProvider.GetClusters(cluster_domain.ClusterRequest{})
	assert.NotNil(t, err)
	assert.Nil(t, response)
	assert.NotNil(t, err.Error)
	assert.EqualValues(t, json, err.Response)
}

// When we apply a filter that filters all clusters. This is a valid response
func TestGetClustersNoMatchingClusters(t *testing.T) {
	json := `{
		"kind": "ClusterList",
		"page": 0,
		"size": 0,
		"total": 0,
		"items": []
	}`
	getRequestFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader(json)),
		}, nil
	}
	restclient.ClusterHTTPClient = &getClientMock{} //without this line, the real api is fired

	response, err := ClusterProvider.GetClusters(cluster_domain.ClusterRequest{
		Filter: discoveryv1.Filter{
			Age: 1,
		},
	})
	assert.NotNil(t, response)
	assert.Nil(t, err)
}

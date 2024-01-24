// Copyright Contributors to the Open Cluster Management project

package cluster

import (
	"fmt"
	"testing"

	discovery "github.com/stolostron/discovery/api/v1"
	"github.com/stretchr/testify/assert"
)

var (
	getClustersFunc func(request ClusterRequest) (*ClusterResponse, *ClusterError)
)

// Mocking the IClusterProvider interface
type clusterProviderMock struct{}

func (cm *clusterProviderMock) GetClusters(request ClusterRequest) (*ClusterResponse, *ClusterError) {
	return getClustersFunc(request)
}

func TestGetClustersBadFormat(t *testing.T) {
	getClustersFunc = func(request ClusterRequest) (*ClusterResponse, *ClusterError) {
		return nil, &ClusterError{
			Error:    fmt.Errorf("invalid json response body"),
			Response: []byte(`{"code": 405, "message":"RESTEASY003650: No resource method found for GET, return 405 with Allow header"}`),
		}
	}
	ClusterProvider = &clusterProviderMock{} //without this line, the real api is fired

	clusterClient := NewClient(ClusterRequest{
		Token:  "access_token",
		Filter: discovery.Filter{LastActive: 1000000000},
	})

	response, err := clusterClient.GetClusters()
	assert.Nil(t, response)
	assert.NotNil(t, err)
}

func TestGetClustersNoError(t *testing.T) {
	getClustersFunc = func(request ClusterRequest) (*ClusterResponse, *ClusterError) {
		return &ClusterResponse{
			Kind:  "ClusterList",
			Page:  1,
			Size:  1,
			Total: 1,
			Items: []Cluster{
				{
					Kind:        "Cluster",
					ID:          "123abc",
					Href:        "/api/clusters_mgmt/v1/clusters/123abc",
					Name:        "mycluster",
					ExternalID:  "mycluster",
					DisplayName: "mycluster",
				},
			},
		}, nil
	}
	ClusterProvider = &clusterProviderMock{} //without this line, the real api is fired

	clusterClient := NewClient(ClusterRequest{
		Token:  "access_token",
		Filter: discovery.Filter{LastActive: 1000000000},
	})

	response, err := clusterClient.GetClusters()
	assert.Nil(t, err)
	assert.NotNil(t, response)
}

func TestNewClient(t *testing.T) {
	clusterRequestConfig := ClusterRequest{
		Token:   "test",
		BaseURL: "testURL",
	}
	clusterClient := ClusterClientGenerator.NewClient(clusterRequestConfig)
	assert.NotNil(t, clusterClient)
}

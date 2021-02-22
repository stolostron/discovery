// Copyright Contributors to the Open Cluster Management project

package cluster_service

import (
	"fmt"
	"testing"

	discoveryv1 "github.com/open-cluster-management/discovery/api/v1"
	"github.com/open-cluster-management/discovery/pkg/api/domain/cluster_domain"
	"github.com/open-cluster-management/discovery/pkg/api/providers/cluster_provider"
	"github.com/stretchr/testify/assert"
)

var (
	getClustersFunc func(request cluster_domain.ClusterRequest) (*cluster_domain.ClusterResponse, *cluster_domain.ClusterError)
)

// Mocking the IClusterProvider interface
type clusterProviderMock struct{}

func (cm *clusterProviderMock) GetClusters(request cluster_domain.ClusterRequest) (*cluster_domain.ClusterResponse, *cluster_domain.ClusterError) {
	return getClustersFunc(request)
}

//When the everything is good
func TestGetClustersBadFormat(t *testing.T) {
	getClustersFunc = func(request cluster_domain.ClusterRequest) (*cluster_domain.ClusterResponse, *cluster_domain.ClusterError) {
		return nil, &cluster_domain.ClusterError{
			Error:    fmt.Errorf("invalid json response body"),
			Response: []byte(`{"code": 405, "message":"RESTEASY003650: No resource method found for GET, return 405 with Allow header"}`),
		}
	}
	cluster_provider.ClusterProvider = &clusterProviderMock{} //without this line, the real api is fired

	clusterClient := NewClient(cluster_domain.ClusterRequest{
		Token:  "access_token",
		Filter: discoveryv1.Filter{},
	})

	response, err := clusterClient.GetClusters()
	assert.Nil(t, response)
	assert.NotNil(t, err)
}

func TestGetClustersNoError(t *testing.T) {
	getClustersFunc = func(request cluster_domain.ClusterRequest) (*cluster_domain.ClusterResponse, *cluster_domain.ClusterError) {
		return &cluster_domain.ClusterResponse{
			Kind:  "ClusterList",
			Page:  1,
			Size:  1,
			Total: 1,
			Items: []cluster_domain.Cluster{
				{
					Kind:        "Cluster",
					ID:          "123abc",
					Href:        "/api/clusters_mgmt/v1/clusters/123abc",
					Name:        "mycluster",
					ExternalID:  "mycluster",
					DisplayName: "mycluster",
				},
			},
			Reason: "200",
		}, nil
	}
	cluster_provider.ClusterProvider = &clusterProviderMock{} //without this line, the real api is fired

	clusterClient := NewClient(cluster_domain.ClusterRequest{
		Token:  "access_token",
		Filter: discoveryv1.Filter{},
	})

	response, err := clusterClient.GetClusters()
	assert.Nil(t, err)
	assert.NotNil(t, response)
}

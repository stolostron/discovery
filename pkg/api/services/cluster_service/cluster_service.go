package cluster_service

import (
	"fmt"

	"github.com/open-cluster-management/discovery/pkg/api/domain/cluster_domain"
	"github.com/open-cluster-management/discovery/pkg/api/providers/cluster_provider"
)

var (
	ocmClusterBaseURL  = "https://api.openshift.com"
	clusterRequestSize = 1000
)

var (
	ClusterClientGenerator ClientGenerator = &clientGenerator{}
)

type clientGenerator struct{}

type ClientGenerator interface {
	NewClient(config cluster_domain.ClusterRequest) ClusterGetter
}

func (client *clientGenerator) NewClient(config cluster_domain.ClusterRequest) ClusterGetter {
	return NewClient(config)
}

func NewClient(config cluster_domain.ClusterRequest) ClusterGetter {
	client := &clusterClient{
		Config: config,
	}
	if client.Config.BaseURL == "" {
		client.Config.BaseURL = ocmClusterBaseURL
	}
	if client.Config.Size == 0 {
		client.Config.Size = clusterRequestSize
	}
	return client
}

type clusterClient struct {
	Config cluster_domain.ClusterRequest
}

type ClusterGetter interface {
	GetClusters() ([]cluster_domain.Cluster, error)
}

func (client *clusterClient) GetClusters() ([]cluster_domain.Cluster, error) {
	discovered := []cluster_domain.Cluster{}
	request := client.Config

	request.Page = 1
	for {
		discoveredList, err := cluster_provider.ClusterProvider.GetClusters(request)
		if err != nil {
			return nil, fmt.Errorf(err.Reason)
		}
		discovered = append(discovered, discoveredList.Items...)
		if len(discoveredList.Items) < request.Size {
			break
		}
		request.Page++
	}
	return discovered, nil
}

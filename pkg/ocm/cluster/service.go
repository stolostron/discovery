// Copyright Contributors to the Open Cluster Management project

package cluster

import (
	"fmt"
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
	NewClient(config ClusterRequest) ClusterGetter
}

func (client *clientGenerator) NewClient(config ClusterRequest) ClusterGetter {
	return NewClient(config)
}

func NewClient(config ClusterRequest) ClusterGetter {
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
	Config ClusterRequest
}

type ClusterGetter interface {
	GetClusters() ([]Cluster, error)
}

func (client *clusterClient) GetClusters() ([]Cluster, error) {
	discovered := []Cluster{}
	request := client.Config

	request.Page = 1
	for {
		discoveredList, err := ClusterProvider.GetClusters(request)
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

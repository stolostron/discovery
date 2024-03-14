// Copyright Contributors to the Open Cluster Management project

package common

import (
	"fmt"

	"github.com/stolostron/discovery/pkg/ocm/cluster"
)

var (
	ocmBaseURL  = "https://api.openshift.com"
	requestSize = 1000
)

var (
	ocmClientGenerator ClientGenerator = &clientGenerator{}
)

type clientGenerator struct{}

func (client *clientGenerator) NewClient(config ClusterRequest) ClusterGetter {
	return NewClient(config)
}

type ClientGenerator interface {
	NewClient(config Request) ClusterGetter
}

func NewClient(config Request) ClusterGetter {
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

type Client struct {
	Config Request
}

type ResourceGetter interface {
	GetResources() ([]cluster.Cluster, error)
}

func (client *Client) GetResources() ([]Cluster, error) {
	discovered := []Cluster{}
	request = client.Config

	request.Page = 1
	for {
		discoveredList, err := provider.GetResources(request)
		if err != nil {
			return nil, fmt.Errorf(err.Reason)
		}

		filteredList := Filter(discoveredList.Items, client.Config.Filter)
		for _, fc := range filteredList {
			discovered = append(discovered, fc)
		}

		if len(discoveredList.Items) < request.Size {
			break
		}
		request.Page++
	}

	return discovered, nil
}

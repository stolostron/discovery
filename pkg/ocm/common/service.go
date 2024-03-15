// Copyright Contributors to the Open Cluster Management project

package common

import (
	"fmt"

	"github.com/stolostron/discovery/pkg/ocm/cluster"
	"github.com/stolostron/discovery/pkg/ocm/subscription"
)

var (
	ocmBaseURL  = "https://api.openshift.com"
	requestSize = 1000
)

var (
	OCMClientGenerator ClientGenerator = &clientGenerator{}
)

type clientGenerator struct{}

type ClientGenerator interface {
	NewClient(config Request) ResourceGetter
}

func (client *clientGenerator) NewClient(config Request) ResourceGetter {
	return NewClient(config)
}

func NewClient(config Request) ResourceGetter {
	client := &Client{
		Config: config,
	}

	if client.Config.BaseURL == "" {
		client.Config.BaseURL = ocmBaseURL
	}
	if client.Config.Size == 0 {
		client.Config.Size = requestSize
	}

	return client
}

type Client struct {
	Config Request
}

type ResourceGetter interface {
	GetClusters() ([]cluster.Cluster, error)
	GetSubscriptions() ([]subscription.Subscription, error)
}

func (client *Client) GetClusters() ([]cluster.Cluster, error) {
	discovered := []cluster.Cluster{}
	request := client.Config

	request.Page = 1
	for {
		discoveredList, err := provider.GetResources(request, cluster.GetClusterURL())
		if err != nil {
			return nil, fmt.Errorf(err.Reason)
		}

		filteredList := FilterResources(discoveredList.Items, client.Config.Filter)
		discovered = append(discovered, filteredList.([]cluster.Cluster)...)

		if len(discoveredList.Items) < request.Size {
			break
		}
		request.Page++
	}

	return discovered, nil
}

func (client *Client) GetSubscriptions() ([]subscription.Subscription, error) {
	discovered := []subscription.Subscription{}
	request := client.Config

	request.Page = 1
	for {
		discoveredList, err := provider.GetResources(request, subscription.GetSubscriptionURL())
		if err != nil {
			return nil, fmt.Errorf(err.Reason)
		}

		filteredList := FilterResources(discoveredList.Items, client.Config.Filter)
		discovered = append(discovered, filteredList.([]subscription.Subscription)...)

		if len(discoveredList.Items) < request.Size {
			break
		}
		request.Page++
	}

	return discovered, nil
}

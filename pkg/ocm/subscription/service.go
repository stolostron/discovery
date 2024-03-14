// Copyright Contributors to the Open Cluster Management project

package subscription

import (
	"fmt"

	"github.com/stolostron/discovery/pkg/ocm/common"
)

var (
	ocmSubscriptionBaseURL  = "https://api.openshift.com"
	subscriptionRequestSize = 1000
)

var (
	SubscriptionClientGenerator ClientGenerator = &clientGenerator{}
)

type clientGenerator struct{}

type ClientGenerator interface {
	NewClient(config SubscriptionRequest) SubscriptionGetter
}

func (client *clientGenerator) NewClient(config SubscriptionRequest) SubscriptionGetter {
	return NewClient(config)
}

func NewClient(config SubscriptionRequest) SubscriptionGetter {
	client := &subscriptionClient{
		Config: config,
	}
	if client.Config.Request.BaseURL == "" {
		client.Config.Request.BaseURL = ocmSubscriptionBaseURL
	}
	if client.Config.Request.Size == 0 {
		client.Config.Request.Size = subscriptionRequestSize
	}
	return client
}

type subscriptionClient struct {
	Config SubscriptionRequest
}

type SubscriptionGetter interface {
	GetSubscriptions() ([]Subscription, error)
}

func (client *subscriptionClient) GetSubscriptions() ([]Subscription, error) {
	discovered := []Subscription{}
	request := client.Config

	request.Request.Page = 1
	for {
		discoveredList, err := SubscriptionProvider.GetSubscriptions(request)
		if err != nil {
			return nil, fmt.Errorf(err.ErrorResponse.Error.Error())
		}

		filteredSubs := common.Filter(discoveredList.Items, client.Config.Request.Filter)
		for _, sub := range filteredSubs {
			discovered = append(discovered, sub)
		}

		if len(discoveredList.Items) < request.Request.Size {
			break
		}
		request.Request.Page++
	}
	return discovered, nil
}

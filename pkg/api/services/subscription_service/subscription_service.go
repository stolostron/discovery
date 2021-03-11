package subscription_service

import (
	"fmt"

	"github.com/open-cluster-management/discovery/pkg/api/domain/subscription_domain"
	"github.com/open-cluster-management/discovery/pkg/api/providers/subscription_provider"
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
	NewClient(config subscription_domain.SubscriptionRequest) SubscriptionGetter
}

func (client *clientGenerator) NewClient(config subscription_domain.SubscriptionRequest) SubscriptionGetter {
	return NewClient(config)
}

func NewClient(config subscription_domain.SubscriptionRequest) SubscriptionGetter {
	client := &subscriptionClient{
		Config: config,
	}
	if client.Config.BaseURL == "" {
		client.Config.BaseURL = ocmSubscriptionBaseURL
	}
	if client.Config.Size == 0 {
		client.Config.Size = subscriptionRequestSize
	}
	return client
}

type subscriptionClient struct {
	Config subscription_domain.SubscriptionRequest
}

type SubscriptionGetter interface {
	GetSubscriptions() ([]subscription_domain.Subscription, error)
}

func (client *subscriptionClient) GetSubscriptions() ([]subscription_domain.Subscription, error) {
	discovered := []subscription_domain.Subscription{}
	request := client.Config

	request.Page = 1
	for {
		discoveredList, err := subscription_provider.SubscriptionProvider.GetSubscriptions(request)
		if err != nil {
			return nil, fmt.Errorf(err.Reason)
		}

		for _, sub := range discoveredList.Items {
			// Filter archived clusters
			if sub.Status == "Archived" {
				continue
			}
			discovered = append(discovered, sub)
		}

		if len(discoveredList.Items) < request.Size {
			break
		}
		request.Page++
	}
	return discovered, nil
}

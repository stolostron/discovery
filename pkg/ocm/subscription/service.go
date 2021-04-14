package subscription

import (
	"fmt"
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
	if client.Config.BaseURL == "" {
		client.Config.BaseURL = ocmSubscriptionBaseURL
	}
	if client.Config.Size == 0 {
		client.Config.Size = subscriptionRequestSize
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

	request.Page = 1
	for {
		discoveredList, err := SubscriptionProvider.GetSubscriptions(request)
		if err != nil {
			return nil, fmt.Errorf(err.Error.Error())
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

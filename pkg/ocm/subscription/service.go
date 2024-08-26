// Copyright Contributors to the Open Cluster Management project

package subscription

import (
	"errors"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/log"
)

var logf = log.Log.WithName("subscription")

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

	logf.V(1).Info("Starting subscription retrieval", "BaseURL", client.Config.BaseURL, "Size", client.Config.Size)

	for {
		// Logging request details
		logf.V(2).Info("Sending subscription request", "Page", request.Page, "Size", request.Size)

		// Fetch the subscriptions
		discoveredList, err := SubscriptionProvider.GetSubscriptions(request)

		if err != nil {
			if err.Error == nil && err.Reason != "" {
				err.Error = errors.New(err.Reason)
			}

			logf.Error(err.Error, "Failed to retrieve subscriptions", "Page", request.Page,
				"BaseURL", client.Config.BaseURL)

			switch err.StatusCode {
			case 401, 403:
				logf.Info(fmt.Sprintf("%v error occurred, returning empty subscription list", err.StatusCode))
				return []Subscription{}, nil

			case 404:
				logf.Info("404 error occurred, returning empty subscription list")
				return []Subscription{}, nil

			case 429:
				logf.Info("429 Too Many Requests: Rate limit hit, returning empty subscription list")
				return []Subscription{}, nil

			case 500, 502, 503, 504:
				logf.Info(fmt.Sprintf("Server error occurred (Code %v), returning empty subscription list",
					err.StatusCode))

				return []Subscription{}, nil

			default:
				return nil, err.Error
			}
		}

		// Handle empty discovered list
		if discoveredList == nil || len(discoveredList.Items) == 0 {
			logf.V(3).Info("Discovered list returned empty or is nil for subscriptions", "Page", request.Page)
			break
		} else {
			logf.V(3).Info("Received subscription response", "Page", request.Page, "Items", len(discoveredList.Items))
		}

		// Filter and append the subscriptions
		filteredSubs := Filter(discoveredList.Items, client.Config.Filter)
		logf.V(3).Info("Filtered subscriptions", "FilteredItems", len(filteredSubs))

		discovered = append(discovered, filteredSubs...)

		if len(discoveredList.Items) < request.Size {
			logf.V(1).Info("Finished retrieving subscriptions", "TotalSubscriptions", len(discovered))
			break
		}
		request.Page++
	}
	return discovered, nil
}

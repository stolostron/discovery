package restclient

import (
	"net/http"
)

type subscriptionClient struct{}

type SubscriptionGetInterface interface {
	Get(*http.Request) (*http.Response, error)
}

var (
	SubscriptionHTTPClient ClusterGetInterface = &clusterClient{}
)

func (c *subscriptionClient) Get(request *http.Request) (*http.Response, error) {
	client := http.Client{}
	return client.Do(request)
}

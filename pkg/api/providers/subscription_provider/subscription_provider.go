package subscription_provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/open-cluster-management/discovery/pkg/api/clients/restclient"
	"github.com/open-cluster-management/discovery/pkg/api/domain/subscription_domain"
)

const (
	subscriptionURL = "%s/api/accounts_mgmt/v1/subscriptions"
)

type subscriptionProvider struct{}

type ISubscriptionProvider interface {
	GetSubscriptions(request subscription_domain.SubscriptionRequest) (*subscription_domain.SubscriptionResponse, *subscription_domain.SubscriptionError)
}

var (
	SubscriptionProvider ISubscriptionProvider = &subscriptionProvider{}
)

func (c *subscriptionProvider) GetSubscriptions(request subscription_domain.SubscriptionRequest) (*subscription_domain.SubscriptionResponse, *subscription_domain.SubscriptionError) {
	getURL := fmt.Sprintf(subscriptionURL, request.BaseURL)
	query := &url.Values{}
	query.Add("size", fmt.Sprintf("%d", request.Size))
	query.Add("page", fmt.Sprintf("%d", request.Page))

	getRequest, err := http.NewRequest("GET", getURL, nil)
	getRequest.URL.RawQuery = query.Encode()
	getRequest.Header.Add("Authorization", fmt.Sprintf("Bearer %s", request.Token))
	getRequest = getRequest.WithContext(context.Background())

	response, err := restclient.SubscriptionHTTPClient.Get(getRequest)
	if err != nil {
		return nil, &subscription_domain.SubscriptionError{
			Error: err,
		}
	}
	defer response.Body.Close()

	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, &subscription_domain.SubscriptionError{
			Error: err,
		}
	}

	// The api owner can decide to change datatypes, etc. When this happen, it might affect the error format returned
	if response.StatusCode > 299 {
		var errResponse subscription_domain.SubscriptionError
		if err := json.Unmarshal(bytes, &errResponse); err != nil {
			return nil, &subscription_domain.SubscriptionError{
				Error:    err,
				Response: bytes}
		}

		if errResponse.Reason == "" {
			errResponse.Error = fmt.Errorf("invalid json response body")
			errResponse.Response = bytes
		}
		return nil, &errResponse
	}

	var result subscription_domain.SubscriptionResponse
	if err := json.Unmarshal(bytes, &result); err != nil {
		return nil, &subscription_domain.SubscriptionError{
			Error:    fmt.Errorf("error unmarshaling response"),
			Response: bytes,
		}
	}

	return &result, nil
}

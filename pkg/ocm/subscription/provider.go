// Copyright Contributors to the Open Cluster Management project

package subscription

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	discovery "github.com/open-cluster-management/discovery/api/v1alpha1"
)

const (
	subscriptionURL = "%s/api/accounts_mgmt/v1/subscriptions"
)

var (
	httpClient           SubscriptionGetInterface = &subscriptionRestClient{}
	SubscriptionProvider ISubscriptionProvider    = &subscriptionProvider{}
)

type SubscriptionGetInterface interface {
	Get(*http.Request) (*http.Response, error)
}

type subscriptionRestClient struct{}

func (c *subscriptionRestClient) Get(request *http.Request) (*http.Response, error) {
	client := http.Client{}
	return client.Do(request)
}

type ISubscriptionProvider interface {
	GetSubscriptions(request SubscriptionRequest) (*SubscriptionResponse, *SubscriptionError)
}

type subscriptionProvider struct{}

func (c *subscriptionProvider) GetSubscriptions(request SubscriptionRequest) (*SubscriptionResponse, *SubscriptionError) {
	getRequest, err := prepareRequest(request)
	if err != nil {
		return nil, &SubscriptionError{
			Error: fmt.Errorf("%s: %w", "error forming request", err),
		}
	}

	response, err := httpClient.Get(getRequest)
	if err != nil {
		return nil, &SubscriptionError{
			Error: fmt.Errorf("%s: %w", "error during request", err),
		}
	}
	defer response.Body.Close()

	return parseResponse(response)
}

func prepareRequest(request SubscriptionRequest) (*http.Request, error) {
	getURL := fmt.Sprintf(subscriptionURL, request.BaseURL)
	query := &url.Values{}
	query.Add("size", fmt.Sprintf("%d", request.Size))
	query.Add("page", fmt.Sprintf("%d", request.Page))
	applyPreFilters(query, request.Filter)

	getRequest, err := http.NewRequest("GET", getURL, nil)
	if err != nil {
		return nil, err
	}
	getRequest.URL.RawQuery = query.Encode()
	getRequest.Header.Add("Authorization", fmt.Sprintf("Bearer %s", request.Token))
	getRequest = getRequest.WithContext(context.Background())
	return getRequest, nil
}

func parseResponse(response *http.Response) (*SubscriptionResponse, *SubscriptionError) {
	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, &SubscriptionError{
			Error: fmt.Errorf("%s: %w", "couldn't read response body", err),
		}
	}

	if response.StatusCode > 299 {
		var errResponse SubscriptionError
		if err := json.Unmarshal(bytes, &errResponse); err != nil {
			return nil, &SubscriptionError{
				Error:    fmt.Errorf("%s: %w", "couldn't unmarshal subscription error response", err),
				Response: bytes,
			}
		}

		if errResponse.Reason == "" {
			errResponse.Error = fmt.Errorf("unexpected json response body")
			errResponse.Response = bytes
		}
		return nil, &errResponse
	}

	var result SubscriptionResponse
	if err := json.Unmarshal(bytes, &result); err != nil {
		return nil, &SubscriptionError{
			Error:    fmt.Errorf("%s: %w", "couldn't unmarshal subscription response", err),
			Response: bytes,
		}
	}

	return &result, nil
}

// applyPreFilters adds fields to the http query to limit the number of items returned
func applyPreFilters(query *url.Values, filters discovery.Filter) {
	if filters.LastActive != 0 {
		layoutISO := "2006-01-02T15:04:05"
		query.Add("search", fmt.Sprintf("updated_at >= '%s'", lastActiveDateTime(time.Now(), filters.LastActive).Format(layoutISO)))
	}
}

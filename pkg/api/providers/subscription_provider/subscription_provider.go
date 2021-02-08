package subscription_provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	discoveryv1 "github.com/open-cluster-management/discovery/api/v1"
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
	applyPreFilters(query, request.Filter)

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

// applyPreFilters adds fields to the http query to limit the number of items returned
func applyPreFilters(query *url.Values, filters discoveryv1.Filter) {
	if filters.Age != 0 {
		query.Add("search", fmt.Sprintf("creation_timestamp >= '%s'", ageDate(time.Now(), filters.Age)))
	}
}

// return the date that is `daysAgo` days before `currentDate` in 'YYYY-MM-DD' format
func ageDate(currentDate time.Time, daysAgo int) string {
	if daysAgo < 0 {
		daysAgo = 0
	}
	cutoffDay := currentDate.AddDate(0, 0, -daysAgo)
	return cutoffDay.Format("2006-01-02")
}

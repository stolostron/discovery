// Copyright Contributors to the Open Cluster Management project

package common

import (
	"fmt"
	"testing"

	discovery "github.com/stolostron/discovery/api/v1"
	"github.com/stolostron/discovery/pkg/ocm/cluster"
	sub "github.com/stolostron/discovery/pkg/ocm/subscription"
	"github.com/stolostron/discovery/pkg/ocm/utils"
	"github.com/stretchr/testify/assert"
)

var (
	getClustersFunc      func() ([]cluster.Cluster, error)
	getSubscriptionsFunc func() ([]sub.Subscription, error)
	resourceGetter       = resourceGetterMock{}
)

// The mocks the GetClusters request to return a select few clusters without connection
// to an external datasource
type resourceGetterMock struct{}

func (m *resourceGetterMock) GetClusters() ([]cluster.Cluster, error) {
	fmt.Printf("in clusters")
	return getClustersFunc()
}

func (m *resourceGetterMock) GetSubscriptions() ([]sub.Subscription, error) {
	fmt.Printf("in subscriptions")
	return getSubscriptionsFunc()
}

// This mocks the NewClient function and returns an instance of the subscriptionGetterMock
type ClientGeneratorMock struct{}

func (m *ClientGeneratorMock) NewClient(config Request) ResourceGetter {
	return &resourceGetter
}

func TestGetSubscriptionsBadFormat(t *testing.T) {
	OCMClientGenerator = &ClientGeneratorMock{}

	getSubscriptionsFunc = func() ([]sub.Subscription, error) {
		return nil, fmt.Errorf("invalid json response body")
	}

	subscriptionClient := OCMClientGenerator.NewClient(Request{
		Token:  "access_token",
		Filter: discovery.Filter{LastActive: 1000000000},
	})

	response, err := subscriptionClient.GetSubscriptions()
	fmt.Printf("response: %v - err: %v", response, err)
	assert.Nil(t, response)
	assert.NotNil(t, err)
}

func TestGetSubscriptionsNoError(t *testing.T) {
	OCMClientGenerator = &ClientGeneratorMock{} //without this line, the real api is fired
	getSubscriptionsFunc = func() ([]sub.Subscription, error) {
		return []sub.Subscription{
			{
				Kind:    "Subscription",
				ID:      "123abc",
				Href:    "/api/accounts_mgmt/v1/subscriptions/123abc",
				Creator: utils.StandardKind{},
				Status:  "Active",
			},
		}, nil
	}

	mockClient := OCMClientGenerator.NewClient(Request{
		Token:  "access_token",
		Filter: discovery.Filter{LastActive: 1000000000},
	})

	response, err := mockClient.GetSubscriptions()
	assert.Nil(t, err)
	assert.NotNil(t, response)
}

func TestNewClient(t *testing.T) {
	requestConfig := Request{
		Token:   "test",
		BaseURL: "testURL",
	}

	OCMClientGenerator = &ClientGeneratorMock{} //without this line, the real api is fired
	assert.NotNil(t, OCMClientGenerator.NewClient(requestConfig))
}

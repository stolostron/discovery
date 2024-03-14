// // Copyright Contributors to the Open Cluster Management project

// package subscription

// import (
// 	"fmt"
// 	"testing"

// 	discovery "github.com/stolostron/discovery/api/v1"
// 	"github.com/stretchr/testify/assert"
// )

// var (
// 	getSubscriptionsFunc func(request SubscriptionRequest) (*SubscriptionResponse, *SubscriptionError)
// )

// // Mocking the ISubscriptionProvider interface
// type subscriptionProviderMock struct{}

// func (cm *subscriptionProviderMock) GetSubscriptions(request SubscriptionRequest) (*SubscriptionResponse, *SubscriptionError) {
// 	return getSubscriptionsFunc(request)
// }

// func TestGetSubscriptionsBadFormat(t *testing.T) {
// 	getSubscriptionsFunc = func(request SubscriptionRequest) (*SubscriptionResponse, *SubscriptionError) {
// 		return nil, &SubscriptionError{
// 			Error:    fmt.Errorf("invalid json response body"),
// 			Response: []byte(`{"code": 405, "message":"RESTEASY003650: No resource method found for GET, return 405 with Allow header"}`),
// 		}
// 	}
// 	SubscriptionProvider = &subscriptionProviderMock{} //without this line, the real api is fired

// 	subscriptionClient := NewClient(SubscriptionRequest{
// 		Token:  "access_token",
// 		Filter: discovery.Filter{LastActive: 1000000000},
// 	})

// 	response, err := subscriptionClient.GetSubscriptions()
// 	assert.Nil(t, response)
// 	assert.NotNil(t, err)
// }

// func TestGetSubscriptionsNoError(t *testing.T) {
// 	getSubscriptionsFunc = func(request SubscriptionRequest) (*SubscriptionResponse, *SubscriptionError) {
// 		return &SubscriptionResponse{
// 			Kind:  "SubscriptionList",
// 			Page:  1,
// 			Size:  1,
// 			Total: 1,
// 			Items: []Subscription{
// 				{
// 					Kind:    "Subscription",
// 					ID:      "123abc",
// 					Href:    "/api/accounts_mgmt/v1/subscriptions/123abc",
// 					Creator: StandardKind{},
// 					Status:  "Active",
// 				},
// 			},
// 		}, nil
// 	}
// 	SubscriptionProvider = &subscriptionProviderMock{} //without this line, the real api is fired

// 	subscriptionClient := NewClient(SubscriptionRequest{
// 		Token:  "access_token",
// 		Filter: discovery.Filter{LastActive: 1000000000},
// 	})

// 	response, err := subscriptionClient.GetSubscriptions()
// 	assert.Nil(t, err)
// 	assert.NotNil(t, response)
// }

// func TestNewClient(t *testing.T) {
// 	subscriptionRequestConfig := SubscriptionRequest{
// 		Token:   "test",
// 		BaseURL: "testURL",
// 	}
// 	subscriptionClient := SubscriptionClientGenerator.NewClient(subscriptionRequestConfig)
// 	assert.NotNil(t, subscriptionClient)
// }

package subscription_service

import (
	"github.com/open-cluster-management/discovery/pkg/api/domain/subscription_domain"
)

var (
	getSubscriptionsFunc func(request subscription_domain.SubscriptionRequest) (*subscription_domain.SubscriptionResponse, *subscription_domain.SubscriptionError)
)

// Mocking the ISubscriptionProvider interface
type subscriptionProviderMock struct{}

func (cm *subscriptionProviderMock) GetSubscriptions(request subscription_domain.SubscriptionRequest) (*subscription_domain.SubscriptionResponse, *subscription_domain.SubscriptionError) {
	return getSubscriptionsFunc(request)
}

// //When the everything is good
// func TestGetSubscriptionsBadFormat(t *testing.T) {
// 	getSubscriptionsFunc = func(request subscription_domain.SubscriptionRequest) (*subscription_domain.SubscriptionResponse, *subscription_domain.SubscriptionError) {
// 		return nil, &subscription_domain.SubscriptionError{
// 			Error:    fmt.Errorf("invalid json response body"),
// 			Response: []byte(`{"code": 405, "message":"RESTEASY003650: No resource method found for GET, return 405 with Allow header"}`),
// 		}
// 	}
// 	subscription_provider.SubscriptionProvider = &subscriptionProviderMock{} //without this line, the real api is fired

// 	subscriptionClient := NewClient(subscription_domain.SubscriptionRequest{
// 		Token:  "access_token",
// 		Filter: discoveryv1.Filter{},
// 	})

// 	response, err := subscriptionClient.GetSubscriptions()
// 	assert.Nil(t, response)
// 	assert.NotNil(t, err)
// }

// func TestGetSubscriptionsNoError(t *testing.T) {
// 	getSubscriptionsFunc = func(request subscription_domain.SubscriptionRequest) (*subscription_domain.SubscriptionResponse, *subscription_domain.SubscriptionError) {
// 		return &subscription_domain.SubscriptionResponse{
// 			Kind:  "SubscriptionList",
// 			Page:  1,
// 			Size:  1,
// 			Total: 1,
// 			Items: []subscription_domain.Subscription{
// 				{
// 					Kind:        "Subscription",
// 					ID:          "123abc",
// 					Href:        "/api/subscriptions_mgmt/v1/subscriptions/123abc",
// 					Name:        "mysubscription",
// 					ExternalID:  "mysubscription",
// 					DisplayName: "mysubscription",
// 				},
// 			},
// 			Reason: "200",
// 		}, nil
// 	}
// 	subscription_provider.SubscriptionProvider = &subscriptionProviderMock{} //without this line, the real api is fired

// 	subscriptionClient := NewClient(subscription_domain.SubscriptionRequest{
// 		Token:  "access_token",
// 		Filter: discoveryv1.Filter{},
// 	})

// 	response, err := subscriptionClient.GetSubscriptions()
// 	assert.Nil(t, err)
// 	assert.NotNil(t, response)
// }

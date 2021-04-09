package ocm

import (
	discoveryv1 "github.com/open-cluster-management/discovery/api/v1"
	"github.com/open-cluster-management/discovery/pkg/ocm/domain/auth_domain"
	"github.com/open-cluster-management/discovery/pkg/ocm/domain/subscription_domain"
	"github.com/open-cluster-management/discovery/pkg/ocm/services/auth_service"
	"github.com/open-cluster-management/discovery/pkg/ocm/services/subscription_service"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DiscoverClusters returns a list of DiscoveredClusters found in both the accounts_mgmt and
// accounts_mgmt apis with the given filters
func DiscoverClusters(token string, baseURL string, filters discoveryv1.Filter) ([]discoveryv1.DiscoveredCluster, error) {
	// Request ephemeral access token with user token. This will be used for OCM requests
	authRequest := auth_domain.AuthRequest{
		Token:   token,
		BaseURL: baseURL,
	}
	accessToken, err := auth_service.AuthClient.GetToken(authRequest)
	if err != nil {
		return nil, err
	}

	// Get subscriptions from accounts_mgmt api
	subscriptionRequestConfig := subscription_domain.SubscriptionRequest{
		Token:   accessToken,
		BaseURL: baseURL,
		Filter:  filters,
	}
	subscriptionClient := subscription_service.SubscriptionClientGenerator.NewClient(subscriptionRequestConfig)
	subscriptions, err := subscriptionClient.GetSubscriptions()
	if err != nil {
		return nil, err
	}

	var discoveredClusters []discoveryv1.DiscoveredCluster
	for _, sub := range subscriptions {
		// Build a DiscoveredCluster object from the subscription information
		if dc, valid := formatCluster(sub); valid {
			discoveredClusters = append(discoveredClusters, dc)
		}
	}

	return discoveredClusters, nil
}

// formatCluster converts a cluster from OCM form to DiscoveredCluster form, or returns false if it is not valid
func formatCluster(sub subscription_domain.Subscription) (discoveryv1.DiscoveredCluster, bool) {
	discoveredCluster := discoveryv1.DiscoveredCluster{}
	if len(sub.Metrics) == 0 {
		return discoveredCluster, false
	}
	if sub.ExternalClusterID == "" {
		return discoveredCluster, false
	}
	discoveredCluster = discoveryv1.DiscoveredCluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "operator.open-cluster-management.io/v1",
			Kind:       "DiscoveredCluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: sub.ExternalClusterID,
		},
		Spec: discoveryv1.DiscoveredClusterSpec{
			Name:              sub.ExternalClusterID,
			Console:           sub.ConsoleURL,
			CreationTimestamp: sub.CreatedAt,
			ActivityTimestamp: sub.UpdatedAt,
			OpenshiftVersion:  sub.Metrics[0].OpenShiftVersion,
			CloudProvider:     sub.CloudProviderID,
			Status:            sub.Status,
		},
	}
	return discoveredCluster, true
}

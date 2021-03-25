package ocm

import (
	discoveryv1 "github.com/open-cluster-management/discovery/api/v1"
	"github.com/open-cluster-management/discovery/pkg/ocm/domain/auth_domain"
	"github.com/open-cluster-management/discovery/pkg/ocm/domain/cluster_domain"
	"github.com/open-cluster-management/discovery/pkg/ocm/domain/subscription_domain"
	"github.com/open-cluster-management/discovery/pkg/ocm/services/auth_service"
	"github.com/open-cluster-management/discovery/pkg/ocm/services/cluster_service"
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
	newSubs, err := subscriptionClient.GetSubscriptions()
	if err != nil {
		return nil, err
	}
	subscriptionSpecs := make(map[string]discoveryv1.SubscriptionSpec, len(newSubs))
	for _, sub := range newSubs {
		if sub.ExternalClusterID != "" {
			subscriptionSpecs[sub.ExternalClusterID] = formatSubscription(sub)
		}
	}

	// Get clusters from clusters_mgmt api
	requestConfig := cluster_domain.ClusterRequest{
		Token:   accessToken,
		BaseURL: baseURL,
		Filter:  filters,
	}
	clusterClient := cluster_service.ClusterClientGenerator.NewClient(requestConfig)
	newClusters, err := clusterClient.GetClusters()
	if err != nil {
		return nil, err
	}

	var discoveredClusters []discoveryv1.DiscoveredCluster
	for _, cluster := range newClusters {
		subSpec, hasSubscription := subscriptionSpecs[cluster.ExternalID]
		if !hasSubscription {
			continue
		}

		// Build a DiscoveredCluster object from the cluster information
		dc := formatCluster(cluster)
		dc.Spec.Subscription = subSpec
		discoveredClusters = append(discoveredClusters, dc)
	}

	return discoveredClusters, nil
}

// formatSubscription converts a Subscription to a SubscriptionSpec
func formatSubscription(sub subscription_domain.Subscription) discoveryv1.SubscriptionSpec {
	return discoveryv1.SubscriptionSpec{
		Status:       sub.Status,
		SupportLevel: sub.SupportLevel,
		Managed:      sub.Managed,
		CreatorID:    sub.ClusterID,
	}
}

// formatCluster converts a cluster from OCM form to DiscoveredCluster form
func formatCluster(cluster cluster_domain.Cluster) discoveryv1.DiscoveredCluster {
	return discoveryv1.DiscoveredCluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "operator.open-cluster-management.io/v1",
			Kind:       "DiscoveredCluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.ID,
			Namespace: "open-cluster-management",
		},
		Spec: discoveryv1.DiscoveredClusterSpec{
			Console:           cluster.Console.URL,
			CreationTimestamp: cluster.CreationTimestamp,
			ActivityTimestamp: cluster.ActivityTimestamp,
			// ActivityTimestamp: metav1.NewTime(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)),
			OpenshiftVersion: cluster.OpenShiftVersion,
			Name:             cluster.Name,
			Region:           cluster.Region.ID,
			CloudProvider:    cluster.CloudProvider.ID,
			HealthState:      cluster.HealthState,
			State:            cluster.State,
			Product:          cluster.Product.ID,
			// IsManagedCluster:  managedClusterNames[cluster.Name],
			// APIURL: apiurl,
		},
	}
}

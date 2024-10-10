// Copyright Contributors to the Open Cluster Management project

package ocm

import (
	"fmt"
	"strings"

	discovery "github.com/stolostron/discovery/api/v1"
	"github.com/stolostron/discovery/pkg/ocm/auth"
	"github.com/stolostron/discovery/pkg/ocm/subscription"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DiscoverClusters returns a list of DiscoveredClusters found in both the accounts_mgmt and
// accounts_mgmt apis with the given filters
func DiscoverClusters(authRequest auth.AuthRequest, filters discovery.Filter) ([]discovery.DiscoveredCluster, error) {
	// Request ephemeral access token with user token. This will be used for OCM requests
	accessToken, err := auth.AuthClient.GetToken(authRequest)
	if err != nil {
		return nil, err
	}

	// Get subscriptions from accounts_mgmt api
	subscriptionRequestConfig := subscription.SubscriptionRequest{
		Token:   accessToken,
		BaseURL: authRequest.BaseURL,
		Filter:  filters,
	}

	subscriptionClient := subscription.SubscriptionClientGenerator.NewClient(subscriptionRequestConfig)
	subscriptions, err := subscriptionClient.GetSubscriptions()
	if err != nil {
		return nil, err
	}

	var discoveredClusters []discovery.DiscoveredCluster
	for _, sub := range subscriptions {
		// Build a DiscoveredCluster object from the subscription information
		if dc, valid := formatCluster(sub); valid {
			discoveredClusters = append(discoveredClusters, dc)
		}
	}

	return discoveredClusters, nil
}

// formatCluster converts a cluster from OCM form to DiscoveredCluster form, or returns false if it is not valid
func formatCluster(sub subscription.Subscription) (discovery.DiscoveredCluster, bool) {
	discoveredCluster := discovery.DiscoveredCluster{}
	// TODO: consider refactoring to "filter" clusters ouside this function to retain function clarity
	if len(sub.Metrics) == 0 {
		return discoveredCluster, false
	}
	if sub.ExternalClusterID == "" {
		return discoveredCluster, false
	}
	if sub.Status == "Reserved" {
		return discoveredCluster, false
	}
	discoveredCluster = discovery.DiscoveredCluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "operator.open-cluster-management.io/v1",
			Kind:       "DiscoveredCluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: sub.ExternalClusterID,
		},
		Spec: discovery.DiscoveredClusterSpec{
			APIURL:            computeApiUrl(sub),
			ActivityTimestamp: sub.LastTelemetryDate,
			CloudProvider:     sub.CloudProviderID,
			Console:           sub.ConsoleURL,
			CreationTimestamp: sub.CreatedAt,
			DisplayName:       computeDisplayName(sub),
			Name:              sub.ExternalClusterID,
			OCPClusterID:      sub.ExternalClusterID,
			OpenshiftVersion:  sub.Metrics[0].OpenShiftVersion,
			Owner:             sub.Creator.UserName,
			Region:            sub.RegionID,
			RHOCMClusterID:    sub.ClusterID,
			Status:            sub.Status,
			Type:              computeType(sub),
		},
	}
	return discoveredCluster, true
}

// IsUnauthorizedClient returns true if the specified error is unauthorized client side error.
func IsUnauthorizedClient(err error) bool {
	return strings.Contains(err.Error(), auth.ErrUnauthorizedClient.Error())
}

// IsUnrecoverable returns true if the specified error is not temporary
// and will continue to occur with the current state.
func IsUnrecoverable(err error) bool {
	return strings.Contains(err.Error(), auth.ErrInvalidToken.Error())
}

// computeDisplayName tries to provide a more user-friendly name if set
// to a cluster ID
func computeDisplayName(sub subscription.Subscription) string {
	// displayName is custom
	if sub.DisplayName != sub.ExternalClusterID && sub.DisplayName != "" {
		return sub.DisplayName
	}
	// use consoleURL for displayName
	if strings.HasPrefix(sub.ConsoleURL, "https://console-openshift-console.apps.") {
		// trim common prefix
		hostport := strings.TrimPrefix(sub.ConsoleURL, "https://console-openshift-console.apps.")
		// trim port if present
		i := strings.LastIndex(hostport, ":")
		if i > -1 {
			hostport = hostport[:i]
		}
		// replace '.' with '-'
		hostport = strings.ReplaceAll(hostport, ".", "-")

		return hostport
	}
	// Use GUID as backup
	return sub.ExternalClusterID
}

// computeApiUrl calculates the Kubernetes api endpoint from a subscription's consoleURL
func computeApiUrl(sub subscription.Subscription) string {
	consolePrefix := "https://console-openshift-console.apps."
	apiPrefix, apiPort := "https://api", "6443"

	if strings.HasPrefix(sub.ConsoleURL, consolePrefix) {
		// trim common prefix
		name := strings.TrimPrefix(sub.ConsoleURL, consolePrefix)
		// trim port if present
		i := strings.LastIndex(name, ":")
		if i > -1 {
			name = name[:i]
		}
		return fmt.Sprintf("%s.%s:%s", apiPrefix, name, apiPort)
	}
	// doesn't match common pattern
	return ""
}

// computeType calculates the type of the cluster based on subscription.plan.id
func computeType(sub subscription.Subscription) string {
	switch sub.Plan.ID {
	case "MOA", "MOA-HostedControlPlane": // API returns MOA but this is displayed in OCM as ROSA
		return "ROSA"

	default:
		return sub.Plan.ID
	}
}

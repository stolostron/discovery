// Copyright Contributors to the Open Cluster Management project

package ocm

import (
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	discovery "github.com/stolostron/discovery/api/v1"
	"github.com/stolostron/discovery/pkg/ocm/auth"
	"github.com/stolostron/discovery/pkg/ocm/cluster"
	"github.com/stolostron/discovery/pkg/ocm/subscription"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	// Default OCM API base URL
	defaultOCMBaseURL = "https://api.openshift.com"
)

// DiscoverClusters returns a list of DiscoveredClusters found in both the accounts_mgmt and
// clusters_mgmt apis with the given filters
func DiscoverClusters(authRequest auth.AuthRequest, filters discovery.Filter) ([]discovery.DiscoveredCluster, error) {
	log := logf.Log.WithName("ocm-discovery")

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

	// Determine OCM API base URL (default if not set via annotation)
	ocmBaseURL := authRequest.BaseURL
	if ocmBaseURL == "" {
		ocmBaseURL = defaultOCMBaseURL
	}

	// Create cluster client for querying individual ROSA clusters
	clusterClient := cluster.NewClient(ocmBaseURL, accessToken)

	var discoveredClusters []discovery.DiscoveredCluster
	for _, sub := range subscriptions {
		// Build a DiscoveredCluster object from the subscription information
		if dc, valid := formatCluster(sub, clusterClient, log); valid {
			discoveredClusters = append(discoveredClusters, dc)
		}
	}

	return discoveredClusters, nil
}

// formatCluster converts a cluster from OCM form to DiscoveredCluster form, or returns false if it is not valid
func formatCluster(sub subscription.Subscription, clusterClient cluster.Client, log logr.Logger) (discovery.DiscoveredCluster, bool) {
	discoveredCluster := discovery.DiscoveredCluster{}
	// TODO: consider refactoring to "filter" clusters ouside this function to retain function clarity
	if len(sub.Metrics) == 0 {
		return discoveredCluster, false
	}

	// Determine API URL - use cluster_mgmt API for ROSA clusters, heuristic for others
	apiURL := getAPIURL(sub, clusterClient, log)

	discoveredCluster = discovery.DiscoveredCluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "operator.open-cluster-management.io/v1",
			Kind:       "DiscoveredCluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: sub.ExternalClusterID,
		},
		Spec: discovery.DiscoveredClusterSpec{
			APIURL:            apiURL,
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

// getAPIURL determines the API URL for a cluster. For ROSA clusters, it queries the cluster_mgmt
// API to get the actual API URL. For other clusters, it uses the heuristic computation.
// Falls back to heuristic if cluster_mgmt API query fails.
func getAPIURL(sub subscription.Subscription, clusterClient cluster.Client, log logr.Logger) string {
	// Check if this is a ROSA cluster
	if !isROSA(sub.Plan.ID) {
		// Use heuristic for non-ROSA clusters
		return computeApiUrl(sub)
	}

	// For ROSA clusters, try to get the actual API URL from cluster_mgmt API
	if sub.ClusterID == "" {
		log.V(1).Info("ROSA cluster missing ClusterID, using heuristic", "externalID", sub.ExternalClusterID)
		return computeApiUrl(sub)
	}

	clusterInfo, err := clusterClient.GetClusterByID(sub.ClusterID)
	if err != nil {
		// Log the error but don't fail - fall back to heuristic
		log.V(1).Info("Failed to get cluster info from cluster_mgmt API, using heuristic",
			"clusterID", sub.ClusterID,
			"externalID", sub.ExternalClusterID,
			"error", err.Error())
		return computeApiUrl(sub)
	}

	if clusterInfo.API.URL != "" {
		return clusterInfo.API.URL
	}

	// Cluster info retrieved but API URL is empty - fall back to heuristic
	log.V(1).Info("Cluster info missing API URL, using heuristic",
		"clusterID", sub.ClusterID,
		"externalID", sub.ExternalClusterID)
	return computeApiUrl(sub)
}

// isROSA checks if a plan ID represents a ROSA cluster type
func isROSA(planID string) bool {
	switch planID {
	case "MOA", "MOA-HostedControlPlane", "ROSA", "ROSA-HyperShift":
		return true
	default:
		return false
	}
}

// IsInvalidClient returns true if the specified error is invalid client side error.
func IsInvalidClient(err error) bool {
	if err != nil {
		return strings.Contains(err.Error(), auth.ErrInvalidClient.Error())
	}
	return false
}

// IsUnauthorizedClient returns true if the specified error is unauthorized client side error.
func IsUnauthorizedClient(err error) bool {
	if err != nil {
		return strings.Contains(err.Error(), auth.ErrUnauthorizedClient.Error())
	}
	return false
}

// IsUnrecoverable returns true if the specified error is not temporary
// and will continue to occur with the current state.
func IsUnrecoverable(err error) bool {
	if err != nil {
		return strings.Contains(err.Error(), auth.ErrInvalidToken.Error())
	}
	return false
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

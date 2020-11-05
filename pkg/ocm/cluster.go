package ocm

import (
	"fmt"

	discoveryv1 "github.com/open-cluster-management/discovery/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// "github.com/ghodss/yaml"
)

// ClusterList ...
type ClusterList struct {
	Kind   string    `json:"kind"`
	Page   string    `json:"page"`
	Size   int       `json:"size"`
	Total  int       `json:"total"`
	Items  []Cluster `json:"items"`
	Reason string    `json:"reason"`
}

// Cluster ...
type Cluster struct {
	Kind                     string                 `json:"kind"`
	ID                       string                 `json:"id"`
	Href                     string                 `json:"href"`
	Name                     string                 `json:"name"`
	ExternalID               string                 `json:"external_id"`
	DisplayName              string                 `json:"display_name"`
	CreationTimestamp        *metav1.Time           `json:"creation_timestamp,omitempty"`
	ActivityTimestamp        *metav1.Time           `json:"activity_timestamp,omitempty"`
	CloudProvider            StandardKind           `json:"cloud_provider,omitempty"`
	OpenShiftVersion         string                 `json:"openshift_version"`
	Subscription             StandardKind           `json:"subscription,omitempty"`
	Region                   StandardKind           `json:"region,omitempty"`
	Console                  Console                `json:"console,omitempty"`
	Nodes                    map[string]interface{} `json:"nodes,omitempty"`
	State                    string                 `json:"state"`
	Groups                   StandardKind           `json:"groups,omitempty"`
	Network                  interface{}            `json:"network,omitempty"`
	ExternalConfig           map[string]interface{} `json:"external_configuration,omitempty"`
	MultiAZ                  bool                   `json:"multi_az,omitempty"`
	Managed                  bool                   `json:"managed,omitempty"`
	BYOC                     bool                   `json:"byoc,omitempty"`
	CCS                      map[string]interface{} `json:"ccs,omitempty"`
	IdentityProviders        StandardKind           `json:"identity_providers,omitempty"`
	AWSInfraAccessRoleGrants map[string]interface{} `json:"aws_infrastructure_access_role_grants,omitempty"`
	Metrics                  map[string]interface{} `json:"metrics,omitempty"`
	Addons                   StandardKind           `json:"addons,omitempty"`
	Ingresses                StandardKind           `json:"ingresses,omitempty"`
	HealthState              string                 `json:"health_state,omitempty"`
	Product                  StandardKind           `json:"product,omitempty"`
	DNSReady                 bool                   `json:"dns_ready,omitempty"`
}

// ClusterRequest ...
func ClusterRequest(config *discoveryv1.DiscoveryConfig) *OCMRequest {
	ocmRequest := &OCMRequest{path: OCMClusterPath}
	if ocmURL, ok := config.Annotations["OCM_URL"]; ok {
		ocmRequest.path = fmt.Sprintf("%s/api/clusters_mgmt/v1/clusters", ocmURL)
	}
	return ocmRequest
}

// DiscoveredCluster ...
func DiscoveredCluster(cluster Cluster) discoveryv1.DiscoveredCluster {
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

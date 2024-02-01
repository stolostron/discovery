// Copyright Contributors to the Open Cluster Management project

package cluster

import (
	discovery "github.com/stolostron/discovery/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StandardKind ...
type StandardKind struct {
	Kind string `yaml:"kind,omitempty"`
	ID   string `yaml:"id,omitempty"`
	Href string `href:"kind,omitempty"`
}

// Console ...
type Console struct {
	URL string `yaml:"url,omitempty"`
}

// Cluster represents a single cluster format returned by OCM
type Cluster struct {
	Kind                     string                 `json:"kind"`
	ID                       string                 `json:"id"`
	Href                     string                 `json:"href"`
	Name                     string                 `json:"name"`
	API                      APISettings            `json:"api,omitempty"`
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

type APISettings struct {
	URL       string `json:"url,omitempty"`
	Listening string `json:"listening,omitempty"`
}

// ClusterResponse represents the successful response format by OCM on a cluster request
type ClusterResponse struct {
	Kind   string    `json:"kind"`
	Page   int       `json:"page"`
	Size   int       `json:"size"`
	Total  int       `json:"total"`
	Items  []Cluster `json:"items"`
	Reason string    `json:"reason"`
}

// ClusterRequest contains the data used to customize a cluster get request
type ClusterRequest struct {
	BaseURL string
	Token   string
	Page    int
	Size    int
	Filter  discovery.Filter
}

// ClusterError represents the error format response by OCM on a cluster request.
// Full list of responses available at https://api.openshift.com/api/clusters_mgmt/v1/errors/
type ClusterError struct {
	Kind   string `json:"kind"`
	ID     string `json:"id"`
	Href   string `json:"href"`
	Code   string `json:"code"`
	Reason string `json:"reason"`
	// Error is for setting an internal error for tracking
	Error error `json:"-"`
	// Response is for storing the raw response on an error
	Response []byte `json:"-"`
}

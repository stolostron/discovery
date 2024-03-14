// Copyright Contributors to the Open Cluster Management project

package cluster

import (
	cmn "github.com/stolostron/discovery/pkg/ocm/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Cluster represents a single cluster format returned by OCM
type Cluster struct {
	Kind                     string                 `json:"kind"`
	ID                       string                 `json:"id"`
	Href                     string                 `json:"href"`
	Name                     string                 `json:"name"`
	API                      cmn.APISettings        `json:"api,omitempty"`
	ExternalID               string                 `json:"external_id"`
	DisplayName              string                 `json:"display_name"`
	CreationTimestamp        *metav1.Time           `json:"creation_timestamp,omitempty"`
	ActivityTimestamp        *metav1.Time           `json:"activity_timestamp,omitempty"`
	CloudProvider            cmn.StandardKind       `json:"cloud_provider,omitempty"`
	OpenShiftVersion         string                 `json:"openshift_version"`
	Subscription             cmn.StandardKind       `json:"subscription,omitempty"`
	Region                   cmn.StandardKind       `json:"region,omitempty"`
	Console                  cmn.Console            `json:"console,omitempty"`
	Nodes                    map[string]interface{} `json:"nodes,omitempty"`
	State                    string                 `json:"state"`
	Groups                   cmn.StandardKind       `json:"groups,omitempty"`
	Network                  interface{}            `json:"network,omitempty"`
	ExternalConfig           map[string]interface{} `json:"external_configuration,omitempty"`
	MultiAZ                  bool                   `json:"multi_az,omitempty"`
	Managed                  bool                   `json:"managed,omitempty"`
	BYOC                     bool                   `json:"byoc,omitempty"`
	CCS                      map[string]interface{} `json:"ccs,omitempty"`
	IdentityProviders        cmn.StandardKind       `json:"identity_providers,omitempty"`
	AWSInfraAccessRoleGrants map[string]interface{} `json:"aws_infrastructure_access_role_grants,omitempty"`
	Metrics                  map[string]interface{} `json:"metrics,omitempty"`
	Addons                   cmn.StandardKind       `json:"addons,omitempty"`
	Ingresses                cmn.StandardKind       `json:"ingresses,omitempty"`
	HealthState              string                 `json:"health_state,omitempty"`
	Product                  cmn.StandardKind       `json:"product,omitempty"`
	DNSReady                 bool                   `json:"dns_ready,omitempty"`
}

// ClusterResponse represents the successful response format by OCM on a cluster request
type ClusterResponse struct {
	cmn.Response
	Items []Cluster `json:"items"`
}

// ClusterRequest contains the data used to customize a cluster get request
type ClusterRequest struct {
	cmn.Request
}

// ClusterError represents the error format response by OCM on a cluster request.
// Full list of responses available at https://api.openshift.com/api/clusters_mgmt/v1/errors/
type ClusterError struct {
	cmn.ErrorResponse
}

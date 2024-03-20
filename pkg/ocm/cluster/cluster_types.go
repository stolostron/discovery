// Copyright Contributors to the Open Cluster Management project

package cluster

import (
	"github.com/stolostron/discovery/pkg/ocm/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// clusterURL is the URL format for accessing clusters in OCM.
const (
	clusterURL = "%s/api/clusters_mgmt/v1/clusters"
)

// Cluster represents a single cluster format returned by OCM.
type Cluster struct {
	Kind                     string                 `json:"kind" yaml:"kind"`
	ID                       string                 `json:"id" yaml:"id"`
	Href                     string                 `json:"href" yaml:"href"`
	Name                     string                 `json:"name" yaml:"name"`
	API                      utils.APISettings      `json:"api,omitempty" yaml:"api,omitempty"`
	ExternalID               string                 `json:"external_id" yaml:"external_id"`
	DisplayName              string                 `json:"display_name" yaml:"display_name"`
	CreationTimestamp        *metav1.Time           `json:"creation_timestamp,omitempty" yaml:"creation_timestamp,omitempty"`
	ActivityTimestamp        *metav1.Time           `json:"activity_timestamp,omitempty" yaml:"activity_timestamp,omitempty"`
	CloudProvider            utils.StandardKind     `json:"cloud_provider,omitempty" yaml:"cloud_provider,omitempty"`
	OpenShiftVersion         string                 `json:"openshift_version" yaml:"openshift_version"`
	Subscription             utils.StandardKind     `json:"subscription,omitempty" yaml:"subscription,omitempty"`
	Region                   utils.StandardKind     `json:"region,omitempty" yaml:"region,omitempty"`
	Console                  utils.Console          `json:"console,omitempty" yaml:"console,omitempty"`
	Nodes                    map[string]interface{} `json:"nodes,omitempty" yaml:"nodes,omitempty"`
	State                    string                 `json:"state" yaml:"state"`
	Groups                   utils.StandardKind     `json:"groups,omitempty" yaml:"groups,omitempty"`
	Network                  interface{}            `json:"network,omitempty" yaml:"network,omitempty"`
	ExternalConfig           map[string]interface{} `json:"external_configuration,omitempty" yaml:"external_configuration,omitempty"`
	MultiAZ                  bool                   `json:"multi_az,omitempty" yaml:"multi_az,omitempty"`
	Managed                  bool                   `json:"managed,omitempty" yaml:"managed,omitempty"`
	BYOC                     bool                   `json:"byoc,omitempty" yaml:"byoc,omitempty"`
	CCS                      map[string]interface{} `json:"ccs,omitempty" yaml:"ccs,omitempty"`
	IdentityProviders        utils.StandardKind     `json:"identity_providers,omitempty" yaml:"identity_providers,omitempty"`
	AWSInfraAccessRoleGrants map[string]interface{} `json:"aws_infrastructure_access_role_grants,omitempty" yaml:"aws_infrastructure_access_role_grants,omitempty"`
	Metrics                  map[string]interface{} `json:"metrics,omitempty" yaml:"metrics,omitempty"`
	Addons                   utils.StandardKind     `json:"addons,omitempty" yaml:"addons,omitempty"`
	Ingresses                utils.StandardKind     `json:"ingresses,omitempty" yaml:"ingresses,omitempty"`
	HealthState              string                 `json:"health_state,omitempty" yaml:"health_state,omitempty"`
	Product                  utils.StandardKind     `json:"product,omitempty" yaml:"product,omitempty"`
	DNSReady                 bool                   `json:"dns_ready,omitempty" yaml:"dns_ready,omitempty"`
}

// GetClusterURL returns the URL format for accessing clusters in OCM.
func GetClusterURL() string {
	return clusterURL
}

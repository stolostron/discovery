// Copyright Contributors to the Open Cluster Management project

package cluster

import (
	"github.com/stolostron/discovery/pkg/ocm/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	clusterURL = "%s/api/clusters_mgmt/v1/clusters"
)

// Cluster represents a single cluster format returned by OCM
type Cluster struct {
	Kind                     string                 `json:"kind"`
	ID                       string                 `json:"id"`
	Href                     string                 `json:"href"`
	Name                     string                 `json:"name"`
	API                      utils.APISettings      `json:"api,omitempty"`
	ExternalID               string                 `json:"external_id"`
	DisplayName              string                 `json:"display_name"`
	CreationTimestamp        *metav1.Time           `json:"creation_timestamp,omitempty"`
	ActivityTimestamp        *metav1.Time           `json:"activity_timestamp,omitempty"`
	CloudProvider            utils.StandardKind     `json:"cloud_provider,omitempty"`
	OpenShiftVersion         string                 `json:"openshift_version"`
	Subscription             utils.StandardKind     `json:"subscription,omitempty"`
	Region                   utils.StandardKind     `json:"region,omitempty"`
	Console                  utils.Console          `json:"console,omitempty"`
	Nodes                    map[string]interface{} `json:"nodes,omitempty"`
	State                    string                 `json:"state"`
	Groups                   utils.StandardKind     `json:"groups,omitempty"`
	Network                  interface{}            `json:"network,omitempty"`
	ExternalConfig           map[string]interface{} `json:"external_configuration,omitempty"`
	MultiAZ                  bool                   `json:"multi_az,omitempty"`
	Managed                  bool                   `json:"managed,omitempty"`
	BYOC                     bool                   `json:"byoc,omitempty"`
	CCS                      map[string]interface{} `json:"ccs,omitempty"`
	IdentityProviders        utils.StandardKind     `json:"identity_providers,omitempty"`
	AWSInfraAccessRoleGrants map[string]interface{} `json:"aws_infrastructure_access_role_grants,omitempty"`
	Metrics                  map[string]interface{} `json:"metrics,omitempty"`
	Addons                   utils.StandardKind     `json:"addons,omitempty"`
	Ingresses                utils.StandardKind     `json:"ingresses,omitempty"`
	HealthState              string                 `json:"health_state,omitempty"`
	Product                  utils.StandardKind     `json:"product,omitempty"`
	DNSReady                 bool                   `json:"dns_ready,omitempty"`
}

func GetClusterURL() string {
	return clusterURL
}

func NewCluster(name string) *Cluster {
	return &Cluster{
		Name:        name,
		DisplayName: name,
	}
}

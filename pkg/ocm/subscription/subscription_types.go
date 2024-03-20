// Copyright Contributors to the Open Cluster Management project

package subscription

import (
	"github.com/stolostron/discovery/pkg/ocm/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// subscriptionURL is the URL format for accessing subscriptions in OCM.
const (
	subscriptionURL = "%s/api/accounts_mgmt/v1/subscriptions"
)

// Subscription represents a single cluster's subscription format returned by OCM.
type Subscription struct {
	ID                string             `json:"id" yaml:"id"`
	Kind              string             `json:"kind" yaml:"kind"`
	Href              string             `json:"href" yaml:"href"`
	Plan              utils.StandardKind `json:"plan,omitempty" yaml:"plan,omitempty"`
	ClusterID         string             `json:"cluster_id,omitempty" yaml:"cluster_id,omitempty"`
	ConsoleURL        string             `json:"console_url,omitempty" yaml:"console_url,omitempty"`
	ExternalClusterID string             `json:"external_cluster_id,omitempty" yaml:"external_cluster_id,omitempty"`
	OrganizationID    string             `json:"organization_id,omitempty" yaml:"organization_id,omitempty"`
	LastTelemetryDate *metav1.Time       `json:"last_telemetry_date,omitempty" yaml:"last_telemetry_date,omitempty"`
	CreatedAt         *metav1.Time       `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	UpdatedAt         *metav1.Time       `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
	Metrics           []utils.Metrics    `json:"metrics,omitempty" yaml:"metrics,omitempty"`
	CloudProviderID   string             `json:"cloud_provider_id,omitempty" yaml:"cloud_provider_id,omitempty"`
	SupportLevel      string             `json:"support_level,omitempty" yaml:"support_level,omitempty"`
	DisplayName       string             `json:"display_name,omitempty" yaml:"display_name,omitempty"`
	Creator           utils.StandardKind `json:"creator" yaml:"creator"`
	Managed           bool               `json:"managed,omitempty" yaml:"managed,omitempty"`
	Status            string             `json:"status" yaml:"status"`
	Provenance        string             `json:"provenance,omitempty" yaml:"provenance,omitempty"`
	LastReconcileDate *metav1.Time       `json:"last_reconcile_date,omitempty" yaml:"last_reconcile_date,omitempty"`
	LastReleasedAt    *metav1.Time       `json:"last_released_at,omitempty" yaml:"last_released_at,omitempty"`
}

// GetSubscriptionURL returns the URL format for accessing cluster's subscriptions in OCM.
func GetSubscriptionURL() string {
	return subscriptionURL
}

// Copyright Contributors to the Open Cluster Management project

package subscription

import (
	"github.com/stolostron/discovery/pkg/ocm/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	subscriptionURL = "%s/api/accounts_mgmt/v1/subscriptions"
)

// Subscription ...
type Subscription struct {
	ID                string             `json:"id"`
	Kind              string             `json:"kind"`
	Href              string             `json:"href"`
	Plan              utils.StandardKind `json:"plan,omitempty"`
	ClusterID         string             `json:"cluster_id,omitempty"`
	ConsoleURL        string             `json:"console_url,omitempty"`
	ExternalClusterID string             `json:"external_cluster_id,omitempty"`
	OrganizationID    string             `json:"organization_id,omitempty"`
	LastTelemetryDate *metav1.Time       `json:"last_telemetry_date,omitempty"`
	CreatedAt         *metav1.Time       `json:"created_at,omitempty"`
	UpdatedAt         *metav1.Time       `json:"updated_at,omitempty"`
	Metrics           []utils.Metrics    `json:"metrics,omitempty"`
	CloudProviderID   string             `json:"cloud_provider_id,omitempty"`
	SupportLevel      string             `json:"support_level,omitempty"`
	DisplayName       string             `json:"display_name,omitempty"`
	Creator           utils.StandardKind `json:"creator"`
	Managed           bool               `json:"managed,omitempty"`
	Status            string             `json:"status"`
	Provenance        string             `json:"provenance,omitempty"`
	LastReconcileDate *metav1.Time       `json:"last_reconcile_date,omitempty"`
	LastReleasedAt    string             `json:"last_released_at,omitempty"`
}

func GetSubscriptionURL() string {
	return subscriptionURL
}

// NewSubscription creates a new Subscription object with the provided ID and kind.
func NewSubscription(id, kind, name string) *Subscription {
	return &Subscription{
		ID:          id,
		Kind:        kind,
		DisplayName: name,
	}
}

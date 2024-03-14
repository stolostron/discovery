// Copyright Contributors to the Open Cluster Management project

package subscription

import (
	cmn "github.com/stolostron/discovery/pkg/ocm/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Subscription ...
type Subscription struct {
	ID                string           `json:"id"`
	Kind              string           `json:"kind"`
	Href              string           `json:"href"`
	Plan              cmn.StandardKind `json:"plan,omitempty"`
	ClusterID         string           `json:"cluster_id,omitempty"`
	ConsoleURL        string           `json:"console_url,omitempty"`
	ExternalClusterID string           `json:"external_cluster_id,omitempty"`
	OrganizationID    string           `json:"organization_id,omitempty"`
	LastTelemetryDate *metav1.Time     `json:"last_telemetry_date,omitempty"`
	CreatedAt         *metav1.Time     `json:"created_at,omitempty"`
	UpdatedAt         *metav1.Time     `json:"updated_at,omitempty"`
	Metrics           []cmn.Metrics    `json:"metrics,omitempty"`
	CloudProviderID   string           `json:"cloud_provider_id,omitempty"`
	SupportLevel      string           `json:"support_level,omitempty"`
	DisplayName       string           `json:"display_name,omitempty"`
	Creator           cmn.StandardKind `json:"creator"`
	Managed           bool             `json:"managed,omitempty"`
	Status            string           `json:"status"`
	Provenance        string           `json:"provenance,omitempty"`
	LastReconcileDate *metav1.Time     `json:"last_reconcile_date,omitempty"`
	LastReleasedAt    string           `json:"last_released_at,omitempty"`
}

// SubscriptionResponse ...
type SubscriptionResponse struct {
	cmn.Response
	Items []Subscription `json:"items"`
}

// SubscriptionRequest contains the data used to customize a subscription get request
type SubscriptionRequest struct {
	cmn.Request
}

// SubscriptionError represents the error format response by OCM on a subscription request.
// Full list of responses available at https://api.openshift.com/api/accounts_mgmt/v1/errors/
type SubscriptionError struct {
	cmn.ErrorResponse
}

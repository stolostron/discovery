// Copyright Contributors to the Open Cluster Management project

package subscription

import (
	discovery "github.com/stolostron/discovery/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StandardKind ...
type StandardKind struct {
	Kind string `json:"kind,omitempty" yaml:"kind,omitempty"`
	ID   string `json:"id,omitempty" yaml:"id,omitempty"`
	Href string `json:"href,omitempty" yaml:"href,omitempty"`
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
}

// Creator ...
type Creator struct {
	Email     string `json:"email,omitempty" yaml:"email,omitempty"`
	FirstName string `json:"first_name,omitempty" yaml:"first_name,omitempty"`
	Href      string `json:"href,omitempty" yaml:"href,omitempty"`
	ID        string `json:"id,omitempty" yaml:"id,omitempty"`
	Kind      string `json:"kind,omitempty" yaml:"kind,omitempty"`
	LastName  string `json:"last_name,omitempty" yaml:"last_name,omitempty"`
	Name      string `json:"name,omitempty" yaml:"name,omitempty"`
	UserName  string `json:"username,omitempty" yaml:"username,omitempty"`
}

// Metrics ...
type Metrics struct {
	OpenShiftVersion string `json:"openshift_version,omitempty"`
}

// Subscription ...
type Subscription struct {
	ClusterID         string       `json:"cluster_id,omitempty" yaml:"cluster_id,omitempty"`
	CloudProviderID   string       `json:"cloud_provider_id,omitempty" yaml:"cloud_provider_id,omitempty"`
	ConsoleURL        string       `json:"console_url,omitempty" yaml:"console_url,omitempty"`
	CreatedAt         *metav1.Time `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	Creator           Creator      `json:"creator" yaml:"creator"`
	DisplayName       string       `json:"display_name,omitempty" yaml:"display_name,omitempty"`
	ExternalClusterID string       `json:"external_cluster_id,omitempty" yaml:"external_cluster_id,omitempty"`
	Href              string       `json:"href" yaml:"href"`
	ID                string       `json:"id" yaml:"id"`
	Kind              string       `json:"kind" yaml:"kind"`
	LastReconcileDate *metav1.Time `json:"last_reconcile_date,omitempty" yaml:"last_reconcile_date,omitempty"`
	LastReleasedAt    string       `json:"last_released_at,omitempty" yaml:"last_released_at,omitempty"`
	LastTelemetryDate *metav1.Time `json:"last_telemetry_date,omitempty" yaml:"last_telemetry_date,omitempty"`
	Managed           bool         `json:"managed,omitempty" yaml:"managed,omitempty"`
	Metrics           []Metrics    `json:"metrics,omitempty" yaml:"metrics,omitempty"`
	OrganizationID    string       `json:"organization_id,omitempty" yaml:"organization_id,omitempty"`
	Plan              StandardKind `json:"plan,omitempty" yaml:"plan,omitempty"`
	Provenance        string       `json:"provenance,omitempty" yaml:"provenance,omitempty"`
	RegionID          string       `json:"region_id,omitempty" yaml:"region_id,omitempty"`
	Status            string       `json:"status" yaml:"status"`
	SupportLevel      string       `json:"support_level,omitempty" yaml:"support_level,omitempty"`
	UpdatedAt         *metav1.Time `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
}

// SubscriptionResponse ...
type SubscriptionResponse struct {
	Kind  string         `json:"kind"`
	Page  int            `json:"page"`
	Size  int            `json:"size"`
	Total int            `json:"total"`
	Items []Subscription `json:"items"`
}

// SubscriptionRequest contains the data used to customize a subscription get request
type SubscriptionRequest struct {
	BaseURL string
	Token   string
	Page    int
	Size    int
	Filter  discovery.Filter
}

// SubscriptionError represents the error format response by OCM on a subscription request.
// Full list of responses available at https://api.openshift.com/api/accounts_mgmt/v1/errors/
type SubscriptionError struct {
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

package subscription_domain

import (
	discoveryv1 "github.com/open-cluster-management/discovery/api/v1"
)

// StandardKind ...
type StandardKind struct {
	Kind string `yaml:"kind,omitempty"`
	ID   string `yaml:"id,omitempty"`
	Href string `href:"kind,omitempty"`
}

// Subscription ...
type Subscription struct {
	ID                string       `yaml:"id" json:"id"`
	Kind              string       `yaml:"kind" json:"kind"`
	Href              string       `yaml:"href" json:"href"`
	Plan              StandardKind `yaml:"plan,omitempty" json:"plan,omitempty"`
	ClusterID         string       `yaml:"cluster_id,omitempty" json:"cluster_id,omitempty"`
	ExternalClusterID string       `yaml:"external_cluster_id,omitempty" json:"external_cluster_id,omitempty"`
	OrganizationID    string       `yaml:"organization_id,omitempty" json:"organization_id,omitempty"`
	LastTelemetryDate string       `yaml:"last_telemetry_date,omitempty" json:"last_telemetry_date,omitempty"`
	CreatedAt         string       `yaml:"created_at,omitempty" json:"created_at,omitempty"`
	UpdatedAt         string       `yaml:"updated_at,omitempty" json:"updated_at,omitempty"`
	SupportLevel      string       `yaml:"support_level,omitempty" json:"support_level,omitempty"`
	DisplayName       string       `yaml:"display_name,omitempty" json:"display_name,omitempty"`
	Creator           StandardKind `yaml:"creator" json:"creator"`
	Managed           bool         `yaml:"managed,omitempty" json:"managed,omitempty"`
	Status            string       `yaml:"status" json:"status"`
	Provenance        string       `yaml:"provenance,omitempty" json:"provenance,omitempty"`
	LastReconcileDate string       `yaml:"last_reconcile_date,omitempty" json:"last_reconcile_date,omitempty"`
	LastReleasedAt    string       `yaml:"last_released_at,omitempty" json:"last_released_at,omitempty"`
	Reason            string       `yaml:"reason,omitempty" json:"reason,omitempty"`
}

// SubscriptionResponse ...
type SubscriptionResponse struct {
	Kind   string         `yaml:"kind" json:"kind"`
	Page   string         `yaml:"page" json:"page"`
	Size   int            `yaml:"size" json:"size"`
	Total  int            `yaml:"total" json:"total"`
	Items  []Subscription `yaml:"items" json:"items"`
	Reason string         `yaml:"reason" json:"reason"`
}

// SubscriptionRequest contains the data used to customize a subscription get request
type SubscriptionRequest struct {
	BaseURL string
	Token   string
	Page    int
	Size    int
	Filter  discoveryv1.Filter
}

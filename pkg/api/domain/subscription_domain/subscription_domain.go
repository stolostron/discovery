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
	ID                string       `json:"id"`
	Kind              string       `json:"kind"`
	Href              string       `json:"href"`
	Plan              StandardKind `json:"plan,omitempty"`
	ClusterID         string       `json:"cluster_id,omitempty"`
	ExternalClusterID string       `json:"external_cluster_id,omitempty"`
	OrganizationID    string       `json:"organization_id,omitempty"`
	LastTelemetryDate string       `json:"last_telemetry_date,omitempty"`
	CreatedAt         string       `json:"created_at,omitempty"`
	UpdatedAt         string       `json:"updated_at,omitempty"`
	SupportLevel      string       `json:"support_level,omitempty"`
	DisplayName       string       `json:"display_name,omitempty"`
	Creator           StandardKind `json:"creator"`
	Managed           bool         `json:"managed,omitempty"`
	Status            string       `json:"status"`
	Provenance        string       `json:"provenance,omitempty"`
	LastReconcileDate string       `json:"last_reconcile_date,omitempty"`
	LastReleasedAt    string       `json:"last_released_at,omitempty"`
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
	Filter  discoveryv1.Filter
}

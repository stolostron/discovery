// Copyright Contributors to the Open Cluster Management project

package common

import (
	"net/http"

	discovery "github.com/stolostron/discovery/api/v1"
)

// APISettings represents settings related to an API.
type APISettings struct {
	URL       string `json:"url,omitempty"`
	Listening string `json:"listening,omitempty"`
}

// Console represents settings related to a console.
type Console struct {
	URL string `yaml:"url,omitempty"`
}

// ErrorResponse represents the common structure for error responses.
type ErrorResponse struct {
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

// Metrics represents metrics related to a system or application.
type Metrics struct {
	OpenShiftVersion string `json:"openshift_version,omitempty"`
}

type Provider struct {
	GetInterface
}

// Response represents the common structure for API responses.
type Response struct {
	Kind   string `json:"kind"`
	Page   int    `json:"page"`
	Size   int    `json:"size"`
	Total  int    `json:"total"`
	Reason string `json:"reason,omitempty"` // Reason is specific to ClusterResponse.
	// Add any other common fields here.
}

type RestClient struct{}

// Request represents the common structure for API requests.
type Request struct {
	BaseURL string
	Token   string
	Page    int
	Size    int
	Filter  discovery.Filter
	// Add any other common fields here.
}

// StandardKind represents a standard kind with optional ID and Href.
type StandardKind struct {
	Kind string `yaml:"kind,omitempty"`
	ID   string `yaml:"id,omitempty"`
	Href string `href:"kind,omitempty"`
}

type GetInterface interface {
	Get(*http.Request) (*http.Response, error)
}

type ProviderInterface interface {
	GetData(request Request) (interface{}, error)
}

// Copyright Contributors to the Open Cluster Management project

package common

import (
	"net/http"

	discovery "github.com/stolostron/discovery/api/v1"
)

// Error represents the common structure for error responses.
type Error struct {
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

// Response represents the common structure for API responses.
type Response struct {
	Kind   string        `json:"kind"`
	Page   int           `json:"page"`
	Size   int           `json:"size"`
	Total  int           `json:"total"`
	Reason string        `json:"reason,omitempty"` // Reason is specific to ClusterResponse.
	Items  []interface{} `json:"items"`
	// Add any other common fields here.
}

// RestClient represents an HTTP client for making requests
type RestClient struct{}

// RestProvider implements ResourceLoader interface
type RestProvider struct {
	ResourceLoader
}

// Request represents the common structure for API requests.
type Request struct {
	BaseURL string
	Token   string
	Page    int
	Size    int
	Filter  discovery.Filter
	// Add any other common fields here.
}

// HTTPRequester represents an interface for making HTTP requests
type HTTPRequester interface {
	Get(*http.Request) (*http.Response, error)
}

type ResourceLoader interface {
	GetResources(request Request, endpointURL string) (*Response, *Error)
}

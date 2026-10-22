// Copyright Contributors to the Open Cluster Management project

package cluster

// APISettings contains API server information
type APISettings struct {
	URL string `json:"url,omitempty"`
}

// Cluster represents a minimal cluster format returned by OCM cluster_mgmt API
// We only include fields we actually need to minimize parsing overhead
type Cluster struct {
	Kind string      `json:"kind"`
	ID   string      `json:"id"`
	Href string      `json:"href"`
	API  APISettings `json:"api"`
}

// ClusterError represents an error response from the cluster_mgmt API
type ClusterError struct {
	Kind   string `json:"kind"`
	ID     string `json:"id"`
	Href   string `json:"href"`
	Code   string `json:"code"`
	Reason string `json:"reason"`
	// Error is for setting an internal error for tracking
	Error error `json:"-"`
}

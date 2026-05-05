// Copyright Contributors to the Open Cluster Management project

package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	clusterByIDURL = "%s/api/clusters_mgmt/v1/clusters/%s"
)

var logf = log.Log.WithName("cluster-client")

// Client interface for getting cluster information
type Client interface {
	GetClusterByID(clusterID string) (*Cluster, error)
}

// HTTPClient interface for making HTTP requests
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

// clusterClient implements the Client interface
type clusterClient struct {
	baseURL    string
	token      string
	httpClient HTTPClient
}

// NewClient creates a new cluster client
func NewClient(baseURL, token string) Client {
	return &clusterClient{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second, // Prevent indefinite blocking
		},
	}
}

// GetClusterByID retrieves a single cluster by its ID from the cluster_mgmt API
func (c *clusterClient) GetClusterByID(clusterID string) (*Cluster, error) {
	if clusterID == "" {
		return nil, fmt.Errorf("cluster ID cannot be empty")
	}

	url := fmt.Sprintf(clusterByIDURL, c.baseURL, clusterID)
	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.token))
	req.Header.Add("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle error responses
	if resp.StatusCode >= 400 {
		var clusterErr ClusterError
		if err := json.Unmarshal(body, &clusterErr); err != nil {
			logf.V(1).Info("API error response received", "status", resp.StatusCode, "body", string(body))
			return nil, fmt.Errorf("failed to retrieve cluster information")
		}
		logf.V(1).Info("Cluster API error", "reason", clusterErr.Reason, "code", clusterErr.Code)
		clusterErr.Error = fmt.Errorf("failed to retrieve cluster information")
		return nil, clusterErr.Error
	}

	// Parse successful response
	var cluster Cluster
	if err := json.Unmarshal(body, &cluster); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cluster response: %w", err)
	}

	return &cluster, nil
}

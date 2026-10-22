// Copyright Contributors to the Open Cluster Management project

package cluster

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"
)

// mockHTTPClient is a mock implementation of HTTPClient for testing
type mockHTTPClient struct {
	response *http.Response
	err      error
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.response, m.err
}

func TestGetClusterByID_Success(t *testing.T) {
	responseBody := `{
		"kind": "Cluster",
		"id": "test-cluster-id",
		"href": "/api/clusters_mgmt/v1/clusters/test-cluster-id",
		"api": {
			"url": "https://api.test-cluster.example.com:443"
		}
	}`

	mockHTTP := &mockHTTPClient{
		response: &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString(responseBody)),
		},
		err: nil,
	}

	client := &clusterClient{
		baseURL:    "https://api.openshift.com",
		token:      "test-token",
		httpClient: mockHTTP,
	}

	cluster, err := client.GetClusterByID("test-cluster-id")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if cluster.ID != "test-cluster-id" {
		t.Errorf("Expected cluster ID 'test-cluster-id', got '%s'", cluster.ID)
	}

	if cluster.API.URL != "https://api.test-cluster.example.com:443" {
		t.Errorf("Expected API URL 'https://api.test-cluster.example.com:443', got '%s'", cluster.API.URL)
	}
}

func TestGetClusterByID_EmptyID(t *testing.T) {
	client := &clusterClient{
		baseURL:    "https://api.openshift.com",
		token:      "test-token",
		httpClient: &mockHTTPClient{},
	}

	_, err := client.GetClusterByID("")
	if err == nil {
		t.Fatal("Expected error for empty cluster ID, got nil")
	}
}

func TestGetClusterByID_HTTPError(t *testing.T) {
	mockHTTP := &mockHTTPClient{
		response: nil,
		err:      fmt.Errorf("network error"),
	}

	client := &clusterClient{
		baseURL:    "https://api.openshift.com",
		token:      "test-token",
		httpClient: mockHTTP,
	}

	_, err := client.GetClusterByID("test-cluster-id")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestGetClusterByID_APIError(t *testing.T) {
	responseBody := `{
		"kind": "Error",
		"id": "404",
		"href": "/api/clusters_mgmt/v1/errors/404",
		"code": "CLUSTERS-MGMT-404",
		"reason": "Cluster not found"
	}`

	mockHTTP := &mockHTTPClient{
		response: &http.Response{
			StatusCode: 404,
			Body:       io.NopCloser(bytes.NewBufferString(responseBody)),
		},
		err: nil,
	}

	client := &clusterClient{
		baseURL:    "https://api.openshift.com",
		token:      "test-token",
		httpClient: mockHTTP,
	}

	_, err := client.GetClusterByID("test-cluster-id")
	if err == nil {
		t.Fatal("Expected error for 404 response, got nil")
	}
}

func TestGetClusterByID_InvalidJSON(t *testing.T) {
	responseBody := `{invalid json}`

	mockHTTP := &mockHTTPClient{
		response: &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString(responseBody)),
		},
		err: nil,
	}

	client := &clusterClient{
		baseURL:    "https://api.openshift.com",
		token:      "test-token",
		httpClient: mockHTTP,
	}

	_, err := client.GetClusterByID("test-cluster-id")
	if err == nil {
		t.Fatal("Expected error for invalid JSON, got nil")
	}
}

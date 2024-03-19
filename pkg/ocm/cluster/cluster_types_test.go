// Copyright Contributors to the Open Cluster Management project

package cluster

import (
	"testing"
)

func TestGetClusterURL(t *testing.T) {
	tests := []struct {
		name        string
		endpointURL string
		want        bool
	}{
		{
			name:        "Should get OCM API URL for cluster",
			endpointURL: "%s/api/clusters_mgmt/v1/clusters",
			want:        true,
		},
		{
			name:        "Should get OCM API URL for cluster",
			endpointURL: "path/to/api/clusters_mgmt/v1/clusters",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.endpointURL == GetClusterURL(); got != tt.want {
				t.Errorf("GetSubscriptionURL() = %v, got %v, want %v", GetClusterURL(), got, tt.want)
			}
		})
	}
}

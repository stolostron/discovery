// Copyright Contributors to the Open Cluster Management project

package subscription

import (
	"testing"
)

func TestGetSubscriptionURL(t *testing.T) {
	tests := []struct {
		name        string
		endpointURL string
		want        bool
	}{
		{
			name:        "Should get OCM API URL for subscription",
			endpointURL: "%s/api/accounts_mgmt/v1/subscriptions",
			want:        true,
		},
		{
			name:        "Should get OCM API URL for subscription",
			endpointURL: "path/to/api/accounts_mgmt/v1/subscriptions",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.endpointURL == GetSubscriptionURL(); got != tt.want {
				t.Errorf("GetSubscriptionURL() = %v, got %v, want %v", GetSubscriptionURL(), got, tt.want)
			}
		})
	}
}

// Copyright Contributors to the Open Cluster Management project

package common

import (
	"testing"
	"time"

	discovery "github.com/stolostron/discovery/api/v1"
	"github.com/stolostron/discovery/pkg/ocm/cluster"
	sub "github.com/stolostron/discovery/pkg/ocm/subscription"
	"github.com/stolostron/discovery/pkg/ocm/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFilterResourcesClusters(t *testing.T) {
	day := metav1.NewTime(time.Date(2020, 5, 29, 6, 0, 0, 0, time.UTC))
	tests := []struct {
		name string
		f    discovery.Filter
		subs []interface{}
		want []cluster.Cluster
	}{
		{
			name: "hi",
			f:    discovery.Filter{LastActive: 1000000000, OpenShiftVersions: []discovery.Semver{"4.8", "4.9"}},
			subs: []interface{}{
				cluster.Cluster{
					DisplayName:       "valid-cluster",
					ExternalID:        "9cf50ab1-1f8a-4205-8a84-6958d49b469b",
					OpenShiftVersion:  "4.8.5",
					State:             "Ready",
					ActivityTimestamp: &day,
				},
				cluster.Cluster{
					DisplayName:       "filtered-by-status",
					OpenShiftVersion:  "4.8.5",
					State:             "Archived",
					ActivityTimestamp: &day,
				},
				cluster.Cluster{
					DisplayName:       "filtered-by-version",
					OpenShiftVersion:  "4.6.5",
					State:             "Ready",
					ActivityTimestamp: &day,
				},
				cluster.Cluster{
					DisplayName:      "filtered-by-date",
					OpenShiftVersion: "4.6.5",
					State:            "Archived",
				},
			},
			want: []cluster.Cluster{
				{
					DisplayName:       "valid-cluster",
					ExternalID:        "9cf50ab1-1f8a-4205-8a84-6958d49b469b",
					OpenShiftVersion:  "4.8.5",
					State:             "Active",
					ActivityTimestamp: &day,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterResources(tt.subs, "cluster", tt.f).([]cluster.Cluster)
			if len(got) != len(tt.want) {
				t.Fatalf("Filter() did not return the desired number of clusters. got = %+v, want %+v", got, tt.want)
			}
			for i := range got {
				if got[i].DisplayName != tt.want[i].DisplayName {
					t.Errorf("Filter() = %+v, want %+v", got, tt.want)
				}
			}
		})
	}
}

func TestFilterResourcesSubscription(t *testing.T) {
	day := metav1.NewTime(time.Date(2020, 5, 29, 6, 0, 0, 0, time.UTC))
	tests := []struct {
		name string
		f    discovery.Filter
		subs []interface{}
		want []sub.Subscription
	}{
		{
			name: "hi",
			f:    discovery.Filter{LastActive: 1000000000, OpenShiftVersions: []discovery.Semver{"4.8", "4.9"}},
			subs: []interface{}{
				sub.Subscription{
					DisplayName:       "valid-subscription",
					ConsoleURL:        "https://console-openshift-console.apps.installer-pool-j88kj.dev01.red-chesterfield.com",
					ExternalClusterID: "9cf50ab1-1f8a-4205-8a84-6958d49b469b",
					Metrics:           []utils.Metrics{{OpenShiftVersion: "4.8.5"}},
					Status:            "Active",
					LastTelemetryDate: &day,
				},
				sub.Subscription{
					DisplayName:       "filtered-by-status",
					Metrics:           []utils.Metrics{{OpenShiftVersion: "4.8.5"}},
					Status:            "Archived",
					LastTelemetryDate: &day,
				},
				sub.Subscription{
					DisplayName:       "filtered-by-version",
					Metrics:           []utils.Metrics{{OpenShiftVersion: "4.6.5"}},
					Status:            "Active",
					LastTelemetryDate: &day,
				},
				sub.Subscription{
					DisplayName: "filtered-by-date",
					Metrics:     []utils.Metrics{{OpenShiftVersion: "4.6.5"}},
					Status:      "Archived",
				},
			},
			want: []sub.Subscription{
				{
					DisplayName:       "valid-subscription",
					ConsoleURL:        "https://console-openshift-console.apps.installer-pool-j88kj.dev01.red-chesterfield.com",
					ExternalClusterID: "9cf50ab1-1f8a-4205-8a84-6958d49b469b",
					Metrics:           []utils.Metrics{{OpenShiftVersion: "4.8.5"}},
					Status:            "Active",
					LastTelemetryDate: &day,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterResources(tt.subs, "subscription", tt.f).([]sub.Subscription)
			if len(got) != len(tt.want) {
				t.Fatalf("Filter() did not return the desired number of subscriptions. got = %+v, want %+v", got, tt.want)
			}
			for i := range got {
				if got[i].DisplayName != tt.want[i].DisplayName {
					t.Errorf("Filter() = %+v, want %+v", got, tt.want)
				}
			}
		})
	}
}

func Test_statusFilter(t *testing.T) {
	tests := []struct {
		name string
		sub  sub.Subscription
		want bool
	}{
		{
			name: "Archived sub",
			sub:  sub.Subscription{Status: "Archived"},
			want: false,
		},
		{
			name: "Deprovisioned sub",
			sub:  sub.Subscription{Status: "Deprovisioned"},
			want: false,
		},
		{
			name: "Non-archived sub",
			sub:  sub.Subscription{Status: "Active"},
			want: true,
		},
		{
			name: "No status",
			sub:  sub.Subscription{Status: ""},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := statusFilter()
			if got := filter(tt.sub); got != tt.want {
				t.Errorf("archiveFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_openshiftVersionFilter(t *testing.T) {
	tests := []struct {
		name     string
		sub      sub.Subscription
		versions []discovery.Semver
		want     bool
	}{
		{
			name:     "Matching version",
			sub:      sub.Subscription{Metrics: []utils.Metrics{{OpenShiftVersion: "4.6.1"}}},
			versions: []discovery.Semver{"4.5", "4.6"},
			want:     true,
		},
		{
			name:     "Old version",
			sub:      sub.Subscription{Metrics: []utils.Metrics{{OpenShiftVersion: "4.6.1"}}},
			versions: []discovery.Semver{"4.8", "4.9"},
			want:     false,
		},
		{
			name:     "Missing version",
			sub:      sub.Subscription{Metrics: []utils.Metrics{{OpenShiftVersion: ""}}},
			versions: []discovery.Semver{"4.8", "4.9"},
			want:     false,
		},
		{
			name:     "Missing metrics",
			sub:      sub.Subscription{},
			versions: []discovery.Semver{"4.8", "4.9"},
			want:     false,
		},
		{
			name:     "No version filter",
			sub:      sub.Subscription{Metrics: []utils.Metrics{{OpenShiftVersion: ""}}},
			versions: []discovery.Semver{},
			want:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := openshiftVersionFilter(tt.versions)
			if got := filter(tt.sub); got != tt.want {
				t.Errorf("openshiftVersionFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_lastActiveFilter(t *testing.T) {
	today := metav1.NewTime(time.Date(2020, 5, 29, 6, 0, 0, 0, time.UTC))
	earlier := metav1.NewTime(time.Date(2020, 5, 29, 5, 0, 0, 0, time.UTC))
	later := metav1.NewTime(time.Date(2020, 5, 29, 7, 0, 0, 0, time.UTC))
	tomorrow := metav1.NewTime(time.Date(2020, 5, 30, 6, 0, 0, 0, time.UTC))
	yesterday := metav1.NewTime(time.Date(2020, 5, 28, 7, 0, 0, 0, time.UTC))
	earlyYesterday := metav1.NewTime(time.Date(2020, 5, 28, 5, 0, 0, 0, time.UTC))

	tests := []struct {
		name    string
		sub     sub.Subscription
		current time.Time
		daysAgo int
		want    bool
	}{
		{
			name:    "Same day earlier",
			sub:     sub.Subscription{LastTelemetryDate: &earlier},
			current: today.Time,
			daysAgo: 0,
			want:    false,
		},
		{
			name:    "Same day later",
			sub:     sub.Subscription{LastTelemetryDate: &later},
			current: today.Time,
			daysAgo: 0,
			want:    true,
		},
		{
			name:    "Day before",
			sub:     sub.Subscription{LastTelemetryDate: &yesterday},
			current: today.Time,
			daysAgo: 0,
			want:    false,
		},
		{
			name:    "Later yesterday",
			sub:     sub.Subscription{LastTelemetryDate: &yesterday},
			current: today.Time,
			daysAgo: 1,
			want:    true,
		},
		{
			name:    "Earlier yesterday",
			sub:     sub.Subscription{LastTelemetryDate: &earlyYesterday},
			current: today.Time,
			daysAgo: 1,
			want:    false,
		},
		{
			name:    "Two days apart",
			sub:     sub.Subscription{LastTelemetryDate: &yesterday},
			current: tomorrow.Time,
			daysAgo: 1,
			want:    false,
		},
		{
			name:    "Negative days ago",
			sub:     sub.Subscription{LastTelemetryDate: &later},
			current: today.Time,
			daysAgo: -1,
			want:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := lastActiveFilter(tt.current, tt.daysAgo)
			if got := filter(tt.sub); got != tt.want {
				t.Errorf("lastActiveFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}

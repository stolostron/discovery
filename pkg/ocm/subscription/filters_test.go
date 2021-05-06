package subscription

import (
	"testing"
	"time"

	discovery "github.com/open-cluster-management/discovery/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFilter(t *testing.T) {
	day := metav1.NewTime(time.Date(2020, 5, 29, 6, 0, 0, 0, time.UTC))
	tests := []struct {
		name string
		f    discovery.Filter
		subs []Subscription
		want []Subscription
	}{
		{
			name: "hi",
			f:    discovery.Filter{LastActive: 1000000000, OpenShiftVersions: []discovery.Semver{"4.8", "4.9"}},
			subs: []Subscription{
				{
					DisplayName:       "valid-subscription",
					ConsoleURL:        "https://console-openshift-console.apps.installer-pool-j88kj.dev01.red-chesterfield.com",
					ExternalClusterID: "9cf50ab1-1f8a-4205-8a84-6958d49b469b",
					Metrics:           []Metrics{{OpenShiftVersion: "4.8.5"}},
					Status:            "Active",
					UpdatedAt:         &day,
				},
				{
					DisplayName: "filtered-by-status",
					Metrics:     []Metrics{{OpenShiftVersion: "4.8.5"}},
					Status:      "Archived",
					UpdatedAt:   &day,
				},
				{
					DisplayName: "filtered-by-version",
					Metrics:     []Metrics{{OpenShiftVersion: "4.6.5"}},
					Status:      "Active",
					UpdatedAt:   &day,
				},
				{
					DisplayName: "filtered-by-date",
					Metrics:     []Metrics{{OpenShiftVersion: "4.6.5"}},
					Status:      "Active",
				},
			},
			want: []Subscription{
				{
					DisplayName:       "valid-subscription",
					ConsoleURL:        "https://console-openshift-console.apps.installer-pool-j88kj.dev01.red-chesterfield.com",
					ExternalClusterID: "9cf50ab1-1f8a-4205-8a84-6958d49b469b",
					Metrics:           []Metrics{{OpenShiftVersion: "4.8.5"}},
					Status:            "Active",
					UpdatedAt:         &day,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Filter(tt.subs, tt.f)
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

func Test_archiveFilter(t *testing.T) {
	tests := []struct {
		name string
		sub  Subscription
		want bool
	}{
		{
			name: "Archived sub",
			sub:  Subscription{Status: "Archived"},
			want: false,
		},
		{
			name: "Non-archived sub",
			sub:  Subscription{Status: "Active"},
			want: true,
		},
		{
			name: "No status",
			sub:  Subscription{Status: ""},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := archiveFilter()
			if got := filter(tt.sub); got != tt.want {
				t.Errorf("archiveFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_openshiftVersionFilter(t *testing.T) {
	tests := []struct {
		name     string
		sub      Subscription
		versions []discovery.Semver
		want     bool
	}{
		{
			name:     "Matching version",
			sub:      Subscription{Metrics: []Metrics{{OpenShiftVersion: "4.6.1"}}},
			versions: []discovery.Semver{"4.5", "4.6"},
			want:     true,
		},
		{
			name:     "Old version",
			sub:      Subscription{Metrics: []Metrics{{OpenShiftVersion: "4.6.1"}}},
			versions: []discovery.Semver{"4.8", "4.9"},
			want:     false,
		},
		{
			name:     "Missing version",
			sub:      Subscription{Metrics: []Metrics{{OpenShiftVersion: ""}}},
			versions: []discovery.Semver{"4.8", "4.9"},
			want:     false,
		},
		{
			name:     "Missing metrics",
			sub:      Subscription{},
			versions: []discovery.Semver{"4.8", "4.9"},
			want:     false,
		},
		{
			name:     "No version filter",
			sub:      Subscription{Metrics: []Metrics{{OpenShiftVersion: ""}}},
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
	theDay := metav1.NewTime(time.Date(2020, 5, 29, 6, 0, 0, 0, time.UTC))
	dayOfEarlier := metav1.NewTime(time.Date(2020, 5, 29, 5, 0, 0, 0, time.UTC))
	dayOfLater := metav1.NewTime(time.Date(2020, 5, 29, 7, 0, 0, 0, time.UTC))
	dayAfter := metav1.NewTime(time.Date(2020, 5, 30, 0, 0, 0, 0, time.UTC))
	dayBefore := metav1.NewTime(time.Date(2020, 5, 28, 0, 0, 0, 0, time.UTC))

	tests := []struct {
		name    string
		sub     Subscription
		current time.Time
		daysAgo int
		want    bool
	}{
		{
			name:    "Same day earlier",
			sub:     Subscription{UpdatedAt: &dayOfEarlier},
			current: theDay.Time,
			daysAgo: 0,
			want:    true,
		},
		{
			name:    "Same day later",
			sub:     Subscription{UpdatedAt: &dayOfLater},
			current: theDay.Time,
			daysAgo: 1,
			want:    true,
		},
		{
			name:    "Day before",
			sub:     Subscription{UpdatedAt: &dayBefore},
			current: theDay.Time,
			daysAgo: 0,
			want:    false,
		},
		{
			name:    "Two days apart",
			sub:     Subscription{UpdatedAt: &dayBefore},
			current: dayAfter.Time,
			daysAgo: 1,
			want:    false,
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

// Copyright Contributors to the Open Cluster Management project

package subscription

import (
	"strings"
	"time"

	discovery "github.com/stolostron/discovery/api/v1"
)

// filterFunc returns true if the Subscription passes the filter
type filterFunc func(sub Subscription) bool

// Filter creates filter functions based on the provided filter spec and returns
// only the list of subscriptions that pass all filters
func Filter(subs []Subscription, f discovery.Filter) []Subscription {
	vsf := make([]Subscription, 0)
	filters := createFilters(f)
	for _, s := range subs {
		if all(s, filters) {
			vsf = append(vsf, s)
		}
	}
	return vsf
}

// all returns true if the Subscription passes all filters
func all(s Subscription, fs []filterFunc) bool {
	for _, f := range fs {
		if !f(s) {
			return false
		}
	}
	return true
}

// createFilters returns a list of filter functions generated from the Filter spec
func createFilters(f discovery.Filter) []filterFunc {
	return []filterFunc{
		statusFilter(),
		clusterTypeFilter(f.ClusterTypes),
		infrastructureProviderFilter(f.InfrastructureProviders),
		openshiftVersionFilter(f.OpenShiftVersions),
		regionFilter(f.Regions),
		lastActiveFilter(time.Now(), f.LastActive),
	}
}

// BOOKMARK: This is where clusters are filtered
// statusFilter filters out clusters with non-functioning status
func statusFilter() filterFunc {
	return func(sub Subscription) bool {
		return sub.Status != "Archived" &&
			sub.Status != "Deprovisioned"
	}
}

// deprovisionedFilter filters out clusters with a 'Deprovisioned' status
func deprovisionedFilter() filterFunc {
	return func(sub Subscription) bool {
		return sub.Status != "Deprovisioned"
	}
}

// openshiftVersionFilter filters out clusters with versions not in the
// list of Major/Minor semver versions
func openshiftVersionFilter(versions []discovery.Semver) filterFunc {
	if len(versions) == 0 {
		// noop filter
		return func(sub Subscription) bool { return true }
	}

	sv := make([]string, len(versions))
	for i, v := range versions {
		sv[i] = string(v)
	}
	return func(sub Subscription) bool {
		if len(sub.Metrics) == 0 {
			return false
		}
		for _, v := range sv {
			if strings.HasPrefix(sub.Metrics[0].OpenShiftVersion, v) {
				return true
			}
		}
		return false
	}
}

// lastActiveFilter filters out clusters that haven't been updated in the last n days
func lastActiveFilter(currentDate time.Time, n int) filterFunc {
	t := lastActiveDateTime(currentDate, n)
	return func(sub Subscription) bool {
		if sub.LastTelemetryDate == nil {
			return false
		}
		return sub.LastTelemetryDate.Time.After(t)
	}
}

// return the time that is `daysAgo` days before `currentDate`
func lastActiveDateTime(currentDate time.Time, daysAgo int) time.Time {
	if daysAgo < 0 {
		daysAgo = 0
	}
	return currentDate.AddDate(0, 0, -daysAgo)
}

func commonFilter[T comparable](list []T, matchFunc func(sub Subscription) T) filterFunc {
	if len(list) == 0 {
		// noop filter
		return func(sub Subscription) bool { return true }
	}

	return func(sub Subscription) bool {
		if len(sub.Metrics) == 0 {
			return false
		}

		value := matchFunc(sub)
		for _, item := range list {
			if value == item {
				return true
			}
		}
		return false
	}
}

// clusterTypeFilter filters out subscriptions with cluster types not in the given list
func clusterTypeFilter(clusterTypes []string) filterFunc {
	return commonFilter(clusterTypes, func(sub Subscription) string {
		return sub.Plan.ID
	})
}

// infrastructureProviderFilter filters out subscriptions with cloud providers not in the given list
func infrastructureProviderFilter(infrastructures []string) filterFunc {
	return commonFilter(infrastructures, func(sub Subscription) string {
		return sub.CloudProviderID
	})
}

// regionFilter filters out subscriptions with regions not in the given list
func regionFilter(regions []string) filterFunc {
	return commonFilter(regions, func(sub Subscription) string {
		return sub.RegionID
	})
}

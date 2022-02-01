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
		archiveFilter(),
		openshiftVersionFilter(f.OpenShiftVersions),
		lastActiveFilter(time.Now(), f.LastActive),
	}
}

// archiveFilter filters out clusters with an 'Archived' status
func archiveFilter() filterFunc {
	return func(sub Subscription) bool {
		return sub.Status != "Archived"
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
		if sub.UpdatedAt == nil {
			return false
		}
		return sub.UpdatedAt.Time.After(t)
	}
}

// return the time that is `daysAgo` days before `currentDate`
func lastActiveDateTime(currentDate time.Time, daysAgo int) time.Time {
	if daysAgo < 0 {
		daysAgo = 0
	}
	return currentDate.AddDate(0, 0, -daysAgo)
}

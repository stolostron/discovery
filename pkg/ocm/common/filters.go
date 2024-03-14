// Copyright Contributors to the Open Cluster Management project

package common

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	discovery "github.com/stolostron/discovery/api/v1"
	"github.com/stolostron/discovery/pkg/ocm/subscription"
)

// filterFunc returns true if the Subscription passes the filter
// type filterFunc func(s subscription.Subscription) bool

// Filter creates filter functions based on the provided filter spec and returns
// only the list of subscriptions that pass all filters
func Filter(subs []subscription.Subscription, f discovery.Filter) []subscription.Subscription {
	vsf := make([]subscription.Subscription, 0)
	filters := createFilters(f)
	for _, s := range subs {
		if all(s, filters) {
			vsf = append(vsf, s)
		}
	}
	return vsf
}

// all returns true if the Subscription passes all filters
func all(s subscription.Subscription, fs []filterFunc) bool {
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
		openshiftVersionFilter(f.OpenShiftVersions),
		lastActiveFilter(time.Now(), f.LastActive),
	}
}

// BOOKMARK: This is where clusters are filtered
// statusFilter filters out clusters with non-functioning status
func statusFilter() filterFunc {
	return func(sub subscription.Subscription) bool {
		return sub.Status != "Archived" &&
			sub.Status != "Deprovisioned"
	}
}

// deprovisionedFilter filters out clusters with a 'Deprovisioned' status
func deprovisionedFilter() filterFunc {
	return func(sub subscription.Subscription) bool {
		return sub.Status != "Deprovisioned"
	}
}

// openshiftVersionFilter filters out clusters with versions not in the
// list of Major/Minor semver versions
func openshiftVersionFilter(versions []discovery.Semver) filterFunc {
	if len(versions) == 0 {
		// noop filter
		return func(sub subscription.Subscription) bool { return true }
	}

	sv := make([]string, len(versions))
	for i, v := range versions {
		sv[i] = string(v)
	}
	return func(sub subscription.Subscription) bool {
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
	return func(sub subscription.Subscription) bool {
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

// applyPreFilters adds fields to the http query to limit the number of items returned
// TODO: Determine if activity_timestamp is the best field to compare against vs updated_at.
func applyPreFilters(query *url.Values, filters discovery.Filter) {
	if filters.LastActive != 0 {
		layoutISO := "2006-01-02T15:04:05"
		query.Add("search", fmt.Sprintf("updated_at >= '%s'", lastActiveDateTime(time.Now(), filters.LastActive).Format(layoutISO)))
	}
}

// Copyright Contributors to the Open Cluster Management project

package common

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	discovery "github.com/stolostron/discovery/api/v1"
	"github.com/stolostron/discovery/pkg/ocm/cluster"
	"github.com/stolostron/discovery/pkg/ocm/subscription"
)

// filterFunc returns true if the Subscription passes the filter
type filterFunc func(resource interface{}) bool

// FilterResources creates filter functions based on the provided filter spec and returns
// only the list of cluster or subscriptions that pass all filters
func FilterResources(resource interface{}, f discovery.Filter) interface{} {
	switch rs := resource.(type) {
	case []cluster.Cluster:
		return filterClusters(rs, f)

	case []subscription.Subscription:
		return filterSubscriptions(rs, f)

	default:
		return nil
	}
}

func filterClusters(clusters []cluster.Cluster, f discovery.Filter) []cluster.Cluster {
	vcf := make([]cluster.Cluster, 0)
	filters := createFilters(f)

	for _, c := range clusters {
		if all(c, filters) {
			vcf = append(vcf, c)
		}
	}

	return vcf
}

// filterSubscriptions
func filterSubscriptions(subs []subscription.Subscription, f discovery.Filter) []subscription.Subscription {
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
func all(resources interface{}, fs []filterFunc) bool {
	for _, f := range fs {
		if !f(resources) {
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

// statusFilter filters out clusters with non-functioning status
func statusFilter() filterFunc {
	return func(resource interface{}) bool {
		switch obj := resource.(type) {
		case cluster.Cluster:
			return obj.State != "Archived" && obj.State != "Deprovisioned"

		case subscription.Subscription:
			return obj.Status != "Archived" && obj.Status != "Deprovisioned"

		default:
			return false
		}
	}
}

// openshiftVersionFilter filters out clusters with versions not in the list of Major/Minor semver versions
func openshiftVersionFilter(versions []discovery.Semver) filterFunc {
	if len(versions) == 0 {
		// noop filter
		return func(resource interface{}) bool { return true }
	}

	sv := make([]string, len(versions))
	for i, v := range versions {
		sv[i] = string(v)
	}

	return func(resource interface{}) bool {
		switch obj := resource.(type) {
		case cluster.Cluster:
			for _, v := range sv {
				if strings.HasPrefix(obj.OpenShiftVersion, v) {
					return true
				}
			}

		case subscription.Subscription:
			if len(obj.Metrics) == 0 {
				return false
			}
			for _, v := range sv {
				if strings.HasPrefix(obj.Metrics[0].OpenShiftVersion, v) {
					return true
				}
			}
		}
		return false
	}
}

// lastActiveFilter filters out clusters that haven't been updated in the last n days
func lastActiveFilter(currentDate time.Time, n int) filterFunc {
	t := lastActiveDateTime(currentDate, n)

	return func(resource interface{}) bool {
		switch obj := resource.(type) {
		case cluster.Cluster:
			return obj.ActivityTimestamp != nil && obj.ActivityTimestamp.Time.After(t)

		case subscription.Subscription:
			return obj.LastTelemetryDate != nil && obj.LastTelemetryDate.Time.After(t)
		}

		return false
	}
}

// lastActiveDateTime return the time that is `daysAgo` days before `currentDate`
func lastActiveDateTime(currentDate time.Time, daysAgo int) time.Time {
	if daysAgo < 0 {
		daysAgo = 0
	}
	return currentDate.AddDate(0, 0, -daysAgo)
}

// applyPreFilters adds fields to the http query to limit the number of items returned
func applyPreFilters(query *url.Values, filters discovery.Filter, objectType string) {
	if filters.LastActive != 0 {
		layoutISO := "2006-01-02T15:04:05"
		var objectFilter string

		switch objectType {
		case "cluster":
			objectFilter = "activity_timestamp"

		default:
			objectFilter = "updated_at"
		}

		query.Add("search", fmt.Sprintf("%s >= '%s'", objectFilter, lastActiveDateTime(
			time.Now(), filters.LastActive).Format(layoutISO)))
	}
}

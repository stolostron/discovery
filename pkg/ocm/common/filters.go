// Copyright Contributors to the Open Cluster Management project

package common

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	discovery "github.com/stolostron/discovery/api/v1"
	"github.com/stolostron/discovery/pkg/ocm/cluster"
	"github.com/stolostron/discovery/pkg/ocm/subscription"
)

// filterFunc represents a function that determines whether a resource passes a filter.
type filterFunc func(resource interface{}) bool

/*
FilterResources filters resources based on the provided filter specifications and object type.
It returns the filtered list of clusters or subscriptions.
*/
func FilterResources(resource []interface{}, objectType string, f discovery.Filter) interface{} {
	var clusters []cluster.Cluster
	var subscriptions []subscription.Subscription

	for _, item := range resource {
		switch objectType {
		case "cluster":
			clusterBytes, err := json.Marshal(item)
			if err != nil {
				logr.Error(err, "Error marshalling object type into json")
				continue
			}

			var c cluster.Cluster
			if err := json.Unmarshal(clusterBytes, &c); err != nil {
				logr.Error(err, "Error unmarshalling object type into json")
				continue
			}
			clusters = append(clusters, c)

		case "subscription":
			subscriptionBytes, err := json.Marshal(item)
			if err != nil {
				logr.Error(err, "Error marshalling object type into json")
				continue
			}

			var s subscription.Subscription
			if err := json.Unmarshal(subscriptionBytes, &s); err != nil {
				logr.Error(err, "Error unmarshalling object type into json")
				continue
			}
			subscriptions = append(subscriptions, s)
		}
	}

	if objectType == "cluster" {
		return filterClusters(clusters, f)

	} else if objectType == "subscription" {
		return filterSubscriptions(subscriptions, f)
	}

	return nil
}

// filterClusters filters clusters based on the provided filter specifications.
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

// filterSubscriptions filters subscriptions based on the provided filter specifications.
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

// all checks if a resource passes all provided filter functions.
func all(resources interface{}, fs []filterFunc) bool {
	for _, f := range fs {
		if !f(resources) {
			return false
		}
	}
	return true
}

// createFilters creates a list of filter functions based on the given Filter specification.
func createFilters(f discovery.Filter) []filterFunc {
	return []filterFunc{
		statusFilter(),
		openshiftVersionFilter(f.OpenShiftVersions),
		lastActiveFilter(time.Now(), f.LastActive),
	}
}

// statusFilter filters out clusters or subscriptions with non-operational status.
func statusFilter() filterFunc {
	return func(resource interface{}) bool {
		switch obj := resource.(type) {
		case cluster.Cluster:
			return obj.State != "Archived" && obj.State != "Deprovisioned"

		case subscription.Subscription:
			return obj.Status != "Archived" && obj.Status != "Deprovisioned"

		default:
			logr.Info(fmt.Sprintf("unknown object type (%T) detected: %v", obj, obj))
			return false
		}
	}
}

/*
openshiftVersionFilter filters out clusters or subscriptions with versions not in the list of Major/Minor
semver versions.
*/
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

// lastActiveFilter filters out clusters or subscriptions that haven't been updated in the last n days.
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

// lastActiveDateTime returns the time that is 'daysAgo' days before 'currentDate'.
func lastActiveDateTime(currentDate time.Time, daysAgo int) time.Time {
	if daysAgo < 0 {
		daysAgo = 0
	}
	return currentDate.AddDate(0, 0, -daysAgo)
}

// applyPreFilters adds fields to the HTTP query to limit the number of items returned.
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

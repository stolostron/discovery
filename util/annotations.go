// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package utils

import (
	"strings"

	discovery "github.com/stolostron/discovery/api/v1"
)

var (
	/*
		AnnotationPreviouslyAutoImported is an annotation used to indicate that a discovered cluster was previously
		imported automatically.
	*/
	AnnotationPreviouslyAutoImported = "discovery.open-cluster-management.io/previously-auto-imported"
)

/*
IsAnnotationTrue checks if a specific annotation key in the given instance is set to "true".
*/
func IsAnnotationTrue(instance *discovery.DiscoveredCluster, annotationKey string) bool {
	a := instance.GetAnnotations()
	if a == nil {
		return false
	}

	value := strings.EqualFold(a[annotationKey], "true")
	return value
}

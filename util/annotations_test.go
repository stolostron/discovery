// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package utils

import (
	"testing"

	discovery "github.com/stolostron/discovery/api/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_IsAnnotationTrue(t *testing.T) {
	tests := []struct {
		name string
		dc   *discovery.DiscoveredCluster
		want bool
	}{
		{
			name: "should return annotation being true",
			dc: &discovery.DiscoveredCluster{
				ObjectMeta: v1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
					Annotations: map[string]string{
						AnnotationPreviouslyAutoImported: "true",
					},
				},
			},
			want: true,
		},
		{
			name: "should return annotation being false",
			dc: &discovery.DiscoveredCluster{
				ObjectMeta: v1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
					Annotations: map[string]string{
						AnnotationPreviouslyAutoImported: "false",
					},
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAnnotationTrue(tt.dc, AnnotationPreviouslyAutoImported); got != tt.want {
				t.Errorf("IsAnnotationTrue(tt.dc, AnnotationPreviouslyAutoImported) = %v, want %v", got, tt.want)
			}
		})
	}
}

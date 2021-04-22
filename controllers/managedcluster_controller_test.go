// Copyright Contributors to the Open Cluster Management project

package controllers

import (
	"testing"

	discoveryv1 "github.com/open-cluster-management/discovery/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func Test_getClusterID(t *testing.T) {
	tests := []struct {
		name           string
		managedCluster unstructured.Unstructured
		want           string
	}{
		{
			name: "Managed cluster without labels populated",
			managedCluster: unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "cluster.open-cluster-management.io/v1",
					"kind":       "ManagedCluster",
					"metadata": map[string]interface{}{
						"name":      "managedcluster1",
						"namespace": "test",
					},
				},
			},
			want: "",
		},
		{
			name: "Managed cluster with labels populated",
			managedCluster: unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "cluster.open-cluster-management.io/v1",
					"kind":       "ManagedCluster",
					"metadata": map[string]interface{}{
						"name":      "managedcluster1",
						"namespace": "test",
						"labels":    map[string]interface{}{"clusterID": "cluster-id-1"},
					},
				},
			},
			want: "cluster-id-1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getClusterID(tt.managedCluster); got != tt.want {
				t.Errorf("getClusterID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_setManagedStatus(t *testing.T) {
	tests := []struct {
		name string
		dc   *discoveryv1.DiscoveredCluster
		want bool
	}{
		{
			name: "Managed labels set",
			dc: &discoveryv1.DiscoveredCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "a",
					Namespace: "b",
					Labels:    map[string]string{"isManagedCluster": "true"},
				},
				Spec: discoveryv1.DiscoveredClusterSpec{
					IsManagedCluster: true,
				},
			},
			want: false,
		},
		{
			name: "Managed label missing",
			dc: &discoveryv1.DiscoveredCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "a",
					Namespace: "b",
					Labels:    map[string]string{"isManagedCluster": "false"},
				},
				Spec: discoveryv1.DiscoveredClusterSpec{
					IsManagedCluster: false,
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := setManagedStatus(tt.dc); got != tt.want {
				t.Errorf("setManagedStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_unsetManagedStatus(t *testing.T) {
	type args struct {
		dc *discoveryv1.DiscoveredCluster
	}
	tests := []struct {
		name string
		dc   *discoveryv1.DiscoveredCluster
		want bool
	}{
		{
			name: "Managed labels set",
			dc: &discoveryv1.DiscoveredCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "a",
					Namespace: "b",
					Labels:    map[string]string{"isManagedCluster": "true"},
				},
				Spec: discoveryv1.DiscoveredClusterSpec{
					IsManagedCluster: true,
				},
			},
			want: true,
		},
		{
			name: "Managed label missing",
			dc: &discoveryv1.DiscoveredCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "a",
					Namespace: "b",
					Labels:    map[string]string{"isManagedCluster": "false"},
				},
				Spec: discoveryv1.DiscoveredClusterSpec{
					IsManagedCluster: false,
				},
			},
			want: false,
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := unsetManagedStatus(tt.dc); got != tt.want {
				t.Errorf("unsetManagedStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

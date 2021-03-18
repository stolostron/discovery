// Copyright Contributors to the Open Cluster Management project

package controllers

import (
	"reflect"
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

func Test_matchingDiscoveredCluster(t *testing.T) {
	a := discoveryv1.DiscoveredCluster{
		Spec: discoveryv1.DiscoveredClusterSpec{
			Name: "dc1",
		},
	}
	b := discoveryv1.DiscoveredCluster{
		Spec: discoveryv1.DiscoveredClusterSpec{
			Name: "dc2",
		},
	}

	type args struct {
		discoveredList *discoveryv1.DiscoveredClusterList
		id             string
	}
	tests := []struct {
		name string
		args args
		want *discoveryv1.DiscoveredCluster
	}{
		{
			name: "Matching cluster",
			args: args{
				discoveredList: &discoveryv1.DiscoveredClusterList{
					Items: []discoveryv1.DiscoveredCluster{a, b},
				},
				id: "dc1",
			},
			want: &a,
		},
		{
			name: "No matching cluster",
			args: args{
				discoveredList: &discoveryv1.DiscoveredClusterList{
					Items: []discoveryv1.DiscoveredCluster{a, b},
				},
				id: "dc0",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchingDiscoveredCluster(tt.args.discoveredList, tt.args.id); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("matchingDiscoveredCluster() = %v, want %v", got, tt.want)
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

func Test_isManagedCluster(t *testing.T) {
	managedClusters := unstructured.UnstructuredList{
		Items: []unstructured.Unstructured{
			{
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
			{
				Object: map[string]interface{}{
					"apiVersion": "cluster.open-cluster-management.io/v1",
					"kind":       "ManagedCluster",
					"metadata": map[string]interface{}{
						"name":      "managedcluster2",
						"namespace": "test",
						"labels":    map[string]interface{}{"clusterID": "cluster-id-2"},
					},
				},
			},
		},
	}

	type args struct {
		dc              discoveryv1.DiscoveredCluster
		managedClusters *unstructured.UnstructuredList
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Cluster is managed",
			args: args{
				dc: discoveryv1.DiscoveredCluster{
					Spec: discoveryv1.DiscoveredClusterSpec{
						Name: "cluster-id-1",
					},
				},
				managedClusters: &managedClusters,
			},
			want: true,
		},
		{
			name: "Cluster is not managed",
			args: args{
				dc: discoveryv1.DiscoveredCluster{
					Spec: discoveryv1.DiscoveredClusterSpec{
						Name: "cluster-id-0",
					},
				},
				managedClusters: &managedClusters,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isManagedCluster(tt.args.dc, tt.args.managedClusters); got != tt.want {
				t.Errorf("isManagedCluster() = %v, want %v", got, tt.want)
			}
		})
	}
}

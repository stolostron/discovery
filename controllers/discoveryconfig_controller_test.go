// Copyright Contributors to the Open Cluster Management project

/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"testing"

	discoveryv1 "github.com/open-cluster-management/discovery/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func Test_parseUserToken(t *testing.T) {
	tests := []struct {
		name    string
		secret  *corev1.Secret
		want    string
		wantErr bool
	}{
		{
			name: "Dummy token set",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				Data: map[string][]byte{
					"metadata": []byte("ocmAPIToken: dummytoken"),
				},
			},
			want:    "dummytoken",
			wantErr: false,
		},
		{
			name: "Missing token",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseUserToken(tt.secret)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseUserToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseUserToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_assignManagedStatus(t *testing.T) {
	discovered := map[string]discoveryv1.DiscoveredCluster{
		"a": {
			ObjectMeta: metav1.ObjectMeta{
				Name:      "a",
				Namespace: "test",
				Labels:    map[string]string{"isManagedCluster": "false"},
			},
			Spec: discoveryv1.DiscoveredClusterSpec{
				IsManagedCluster: false,
			},
		},
		"b": {
			ObjectMeta: metav1.ObjectMeta{
				Name:      "b",
				Namespace: "test",
			},
		},
		"c": {
			ObjectMeta: metav1.ObjectMeta{
				Name:      "c",
				Namespace: "test",
			},
		},
	}

	managed := []unstructured.Unstructured{
		{
			Object: map[string]interface{}{
				"apiVersion": "cluster.open-cluster-management.io/v1",
				"kind":       "ManagedCluster",
				"metadata": map[string]interface{}{
					"name":   "a",
					"labels": map[string]interface{}{"clusterID": "a"},
				},
			},
		},
		{
			Object: map[string]interface{}{
				"apiVersion": "cluster.open-cluster-management.io/v1",
				"kind":       "ManagedCluster",
				"metadata": map[string]interface{}{
					"name":   "b",
					"labels": map[string]interface{}{"clusterID": "b"},
				},
			},
		},
	}

	assignManagedStatus(discovered, managed)

	t.Run("Cluster managed status changed", func(t *testing.T) {
		dc := discovered["a"]
		managedLabel := dc.GetLabels()["isManagedCluster"]
		if !dc.Spec.IsManagedCluster || managedLabel != "true" {
			t.Errorf("Expected cluster %s to be labeled as managed: %+v", dc.Name, dc)
		}
	})
	t.Run("Cluster managed status added", func(t *testing.T) {
		dc := discovered["b"]
		managedLabel := dc.GetLabels()["isManagedCluster"]
		if !dc.Spec.IsManagedCluster || managedLabel != "true" {
			t.Errorf("Expected cluster %s to be labeled as managed: %+v", dc.Name, dc)
		}
	})
	t.Run("Cluster managed status not added", func(t *testing.T) {
		dc := discovered["c"]
		managedLabel := dc.GetLabels()["isManagedCluster"]
		if dc.Spec.IsManagedCluster || managedLabel == "true" {
			t.Errorf("Expected cluster %s to be labeled as managed: %+v", dc.Name, dc)
		}
	})
}

func Test_merge(t *testing.T) {
	type args struct {
		clusters map[string]discoveryv1.DiscoveredCluster
		dc       discoveryv1.DiscoveredCluster
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Combine provider connections",
			args: args{
				clusters: map[string]discoveryv1.DiscoveredCluster{
					"a": {
						ObjectMeta: metav1.ObjectMeta{
							Name:      "a",
							Namespace: "test",
							Labels:    map[string]string{"isManagedCluster": "false"},
						},
						Spec: discoveryv1.DiscoveredClusterSpec{
							Name: "a",
							ProviderConnections: []corev1.ObjectReference{
								{
									APIVersion: "v1",
									Kind:       "Secret",
									Name:       "testsecret1",
									Namespace:  "test",
								},
							},
						},
					},
				},
				dc: discoveryv1.DiscoveredCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "a",
						Namespace: "test",
					},
					Spec: discoveryv1.DiscoveredClusterSpec{
						Name: "a",
						ProviderConnections: []corev1.ObjectReference{
							{
								APIVersion: "v1",
								Kind:       "Secret",
								Name:       "testsecret2",
								Namespace:  "test",
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merge(tt.args.clusters, tt.args.dc)
			if len(tt.args.clusters["a"].Spec.ProviderConnections) != 2 {
				t.Errorf("Expected merged providerConnections: %+v", tt.args.clusters["a"])
			}
		})
	}
}

func Test_same(t *testing.T) {
	cluster1 := discoveryv1.DiscoveredCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "c1",
			Namespace: "test",
			Labels:    map[string]string{"isManagedCluster": "false"},
		},
		Spec: discoveryv1.DiscoveredClusterSpec{
			Name:          "c1",
			CloudProvider: "aws",
			Subscription: discoveryv1.SubscriptionSpec{
				Status: "Active",
			},
		},
	}
	cluster2 := discoveryv1.DiscoveredCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "c2",
			Namespace: "test",
			Labels:    map[string]string{"isManagedCluster": "false"},
		},
		Spec: discoveryv1.DiscoveredClusterSpec{
			Name:          "c2",
			CloudProvider: "aws",
			Subscription: discoveryv1.SubscriptionSpec{
				Status: "Active",
			},
		},
	}

	tests := []struct {
		name string
		c1   discoveryv1.DiscoveredCluster
		c2   discoveryv1.DiscoveredCluster
		want bool
	}{
		{
			name: "Equivalent clusters",
			c1:   cluster1,
			c2:   cluster1,
			want: true,
		},
		{
			name: "Different clusters",
			c1:   cluster1,
			c2:   cluster2,
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := same(tt.c1, tt.c2); got != tt.want {
				t.Errorf("same() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getURLOverride(t *testing.T) {
	tests := []struct {
		name   string
		config *discoveryv1.DiscoveryConfig
		want   string
	}{
		{
			name: "Override annotated",
			config: &discoveryv1.DiscoveryConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "testconfig",
					Namespace:   "test",
					Annotations: map[string]string{baseURLAnnotation: "http://mock-ocm-server.test.svc.cluster.local:3000"},
				},
			},
			want: "http://mock-ocm-server.test.svc.cluster.local:3000",
		},
		{
			name: "No override specified",
			config: &discoveryv1.DiscoveryConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testconfig",
					Namespace: "test",
				},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getURLOverride(tt.config); got != tt.want {
				t.Errorf("getURLOverride() = %v, want %v", got, tt.want)
			}
		})
	}
}

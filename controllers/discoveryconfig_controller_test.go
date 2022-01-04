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

	discovery "github.com/stolostron/discovery/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
					"ocmAPIToken": []byte("dummytoken"),
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
	discovered := map[string]discovery.DiscoveredCluster{
		"a": {
			ObjectMeta: metav1.ObjectMeta{
				Name:      "a",
				Namespace: "test",
				Labels:    map[string]string{"isManagedCluster": "false"},
			},
			Spec: discovery.DiscoveredClusterSpec{
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

	managed := []metav1.PartialObjectMetadata{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "a",
				Labels: map[string]string{"clusterID": "a"},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "b",
				Labels: map[string]string{"clusterID": "b"},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "d",
				Labels: map[string]string{"clusterID": "d"},
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
			t.Errorf("Expected cluster %s to not be labeled as managed: %+v", dc.Name, dc)
		}
	})
	t.Run("Discovered list not added to", func(t *testing.T) {
		if len(discovered) != 3 {
			t.Errorf("The discoveredlist should not change in size. Wanted: %d. Got: %d.", 3, len(discovered))
		}
	})
}

func Test_getURLOverride(t *testing.T) {
	tests := []struct {
		name   string
		config *discovery.DiscoveryConfig
		want   string
	}{
		{
			name: "Override annotated",
			config: &discovery.DiscoveryConfig{
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
			config: &discovery.DiscoveryConfig{
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

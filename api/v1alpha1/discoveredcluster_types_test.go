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

package v1alpha1

import (
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Equal reports whether the spec of a is equal to b.
func TestEqual(t *testing.T) {
	time1 := metav1.NewTime(time.Date(2022, 5, 22, 0, 0, 0, 0, time.UTC))
	time2 := metav1.NewTime(time.Date(2022, 6, 22, 0, 0, 0, 0, time.UTC))

	tests := []struct {
		name string
		dc1  *DiscoveredCluster
		dc2  *DiscoveredCluster
		want bool
	}{
		{
			name: "Equal Discovered Cluster Specs",
			dc1: &DiscoveredCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "Cluster 1",
					Namespace: "test",
				},
				Spec: DiscoveredClusterSpec{
					Name:              "managedcluster",
					DisplayName:       "managedcluster",
					Console:           "test",
					APIURL:            "testURL",
					CreationTimestamp: &time1,
					ActivityTimestamp: &time2,
					Type:              "test",
					OpenshiftVersion:  "4.10",
					CloudProvider:     "test",
					Status:            "testing",
					IsManagedCluster:  true,
				},
			},
			dc2: &DiscoveredCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "Cluster 2",
					Namespace: "test",
				},
				Spec: DiscoveredClusterSpec{
					Name:              "managedcluster",
					DisplayName:       "managedcluster",
					Console:           "test",
					APIURL:            "testURL",
					CreationTimestamp: &time1,
					ActivityTimestamp: &time2,
					Type:              "test",
					OpenshiftVersion:  "4.10",
					CloudProvider:     "test",
					Status:            "testing",
					IsManagedCluster:  true,
				},
			},
			want: true,
		},
		{
			name: "Not Equal Discovered Cluster Specs",
			dc1: &DiscoveredCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "Cluster 1",
					Namespace: "test",
				},
				Spec: DiscoveredClusterSpec{
					Name:              "managedcluster",
					DisplayName:       "managedcluster",
					Console:           "test",
					APIURL:            "testURL",
					CreationTimestamp: &time1,
					ActivityTimestamp: &time2,
					Type:              "test",
					OpenshiftVersion:  "4.10",
					CloudProvider:     "test",
					Status:            "testing",
					IsManagedCluster:  true,
				},
			},
			dc2: &DiscoveredCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "Cluster 2",
					Namespace: "test",
				},
				Spec: DiscoveredClusterSpec{
					Name:              "managedcluster2",
					DisplayName:       "managedcluster2",
					Console:           "test2",
					APIURL:            "testURL2",
					CreationTimestamp: &time2,
					ActivityTimestamp: &time1,
					Type:              "test2",
					OpenshiftVersion:  "4.8",
					CloudProvider:     "test2",
					Status:            "testing2",
					IsManagedCluster:  false,
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.dc1.Equal(*tt.dc2); got != tt.want {
				t.Errorf("Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}

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

package common

import "testing"

func Test_IsSupportedClusterType(t *testing.T) {
	tests := []struct {
		name        string
		clusterType string
		want        bool
	}{
		{
			name:        "should support cluster type MultiClusterEngineHCP",
			clusterType: "MultiClusterEngineHCP",
			want:        true,
		},
		{
			name:        "should support cluster type ROSA",
			clusterType: "ROSA",
			want:        true,
		},
		{
			name:        "should not support cluster type GKE",
			clusterType: "GKE",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSupportedClusterType(tt.clusterType); got != tt.want {
				t.Errorf("IsSupportedClusterType(tt.clusterType) = %v, want %v", got, tt.want)
			}
		})
	}
}

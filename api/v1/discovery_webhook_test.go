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

package v1

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newDiscoveredCluster(clusterType, displayName string, importAsManagedCluster bool) *DiscoveredCluster {
	return &DiscoveredCluster{
		ObjectMeta: metav1.ObjectMeta{Name: "cluster", Namespace: "test"},
		Spec: DiscoveredClusterSpec{
			Type:                   clusterType,
			DisplayName:            displayName,
			ImportAsManagedCluster: importAsManagedCluster,
		},
	}
}

func TestDefault(t *testing.T) {
	if err := (&DiscoveredCluster{}).Default(context.TODO(), newDiscoveredCluster("ROSA", "cluster", false)); err != nil {
		t.Errorf("Default() error = %v, want nil", err)
	}
}

func TestValidateCreate(t *testing.T) {
	tests := []struct {
		name    string
		obj     *DiscoveredCluster
		wantErr bool
	}{
		{
			name:    "supported type with auto import",
			obj:     newDiscoveredCluster("ROSA", "rosa-cluster", true),
			wantErr: false,
		},
		{
			name:    "unsupported type without auto import",
			obj:     newDiscoveredCluster("OCP", "ocp-cluster", false),
			wantErr: false,
		},
		{
			name:    "unsupported type with auto import",
			obj:     newDiscoveredCluster("OCP", "ocp-cluster", true),
			wantErr: true,
		},
		{
			name:    "invalid display name with auto import",
			obj:     newDiscoveredCluster("ROSA", "rosa_cluster", true),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := (&DiscoveredCluster{}).ValidateCreate(context.TODO(), tt.obj); (err != nil) != tt.wantErr {
				t.Errorf("ValidateCreate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateUpdate(t *testing.T) {
	tests := []struct {
		name    string
		oldObj  *DiscoveredCluster
		newObj  *DiscoveredCluster
		wantErr bool
	}{
		{
			name:    "enable auto import on supported type",
			oldObj:  newDiscoveredCluster("MultiClusterEngineHCP", "hcp-cluster", false),
			newObj:  newDiscoveredCluster("MultiClusterEngineHCP", "hcp-cluster", true),
			wantErr: false,
		},
		{
			name:    "enable auto import on unsupported type",
			oldObj:  newDiscoveredCluster("OCP", "ocp-cluster", false),
			newObj:  newDiscoveredCluster("OCP", "ocp-cluster", true),
			wantErr: true,
		},
		{
			name:    "enable auto import with invalid display name",
			oldObj:  newDiscoveredCluster("ROSA", "rosa_cluster", false),
			newObj:  newDiscoveredCluster("ROSA", "rosa_cluster", true),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := (&DiscoveredCluster{}).ValidateUpdate(context.TODO(), tt.oldObj, tt.newObj); (err != nil) != tt.wantErr {
				t.Errorf("ValidateUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateDelete(t *testing.T) {
	if _, err := (&DiscoveredCluster{}).ValidateDelete(context.TODO(), newDiscoveredCluster("OCP", "ocp-cluster", true)); err != nil {
		t.Errorf("ValidateDelete() error = %v, want nil", err)
	}
}

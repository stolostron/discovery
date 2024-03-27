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
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ImportStrategyAnnotation = "discovery.open-cluster-management.io/import-strategy"
	ImportCleanUpFinalizer   = "discovery.open-cluster-management.io/import-cleanup"
)

// DiscoveredClusterSpec defines the desired state of DiscoveredCluster
type DiscoveredClusterSpec struct {
	Name              string       `json:"name" yaml:"name"`
	DisplayName       string       `json:"displayName" yaml:"displayName"`
	OCPClusterID      string       `json:"ocpClusterId,omitempty" yaml:"ocpClusterId,omitempty"`
	RHOCMClusterID    string       `json:"rhocmClusterId,omitempty" yaml:"rhocmClusterId,omitempty"`
	Console           string       `json:"console,omitempty" yaml:"console,omitempty"`
	APIURL            string       `json:"apiUrl" yaml:"apiUrl"`
	CreationTimestamp *metav1.Time `json:"creationTimestamp,omitempty" yaml:"creationTimestamp,omitempty"`
	ActivityTimestamp *metav1.Time `json:"activityTimestamp,omitempty" yaml:"activityTimestamp,omitempty"`
	Type              string       `json:"type" yaml:"type"`
	OpenshiftVersion  string       `json:"openshiftVersion,omitempty" yaml:"openshiftVersion,omitempty"`
	CloudProvider     string       `json:"cloudProvider,omitempty" yaml:"cloudProvider,omitempty"`
	Status            string       `json:"status,omitempty" yaml:"status,omitempty"`
	IsManagedCluster  bool         `json:"isManagedCluster" yaml:"isManagedCluster"`

	Credential corev1.ObjectReference `json:"credential,omitempty" yaml:"credential,omitempty"`
}

// DiscoveredClusterStatus defines the observed state of DiscoveredCluster
type DiscoveredClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:storageversion

// DiscoveredCluster is the Schema for the discoveredclusters API
type DiscoveredCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DiscoveredClusterSpec   `json:"spec,omitempty"`
	Status DiscoveredClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DiscoveredClusterList contains a list of DiscoveredCluster
type DiscoveredClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DiscoveredCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DiscoveredCluster{}, &DiscoveredClusterList{})
}

// Equal reports whether the spec of a is equal to b.
func (a DiscoveredCluster) Equal(b DiscoveredCluster) bool {
	if a.Spec.Name != b.Spec.Name ||
		a.Spec.DisplayName != b.Spec.DisplayName ||
		a.Spec.Console != b.Spec.Console ||
		a.Spec.APIURL != b.Spec.APIURL ||
		a.Spec.CreationTimestamp.Truncate(time.Second) != b.Spec.CreationTimestamp.Truncate(time.Second) ||
		a.Spec.ActivityTimestamp.Truncate(time.Second) != b.Spec.ActivityTimestamp.Truncate(time.Second) ||
		a.Spec.Type != b.Spec.Type ||
		a.Spec.OpenshiftVersion != b.Spec.OpenshiftVersion ||
		a.Spec.CloudProvider != b.Spec.CloudProvider ||
		a.Spec.Status != b.Spec.Status ||
		a.Spec.IsManagedCluster != b.Spec.IsManagedCluster ||
		a.Spec.Credential != b.Spec.Credential {
		return false
	}
	return true
}

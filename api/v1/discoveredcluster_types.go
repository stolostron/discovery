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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DiscoveredClusterSpec defines the desired state of DiscoveredCluster
type DiscoveredClusterSpec struct {
	Name              string       `json:"name,omitempty" yaml:"name,omitempty"`
	Console           string       `json:"console,omitempty" yaml:"console,omitempty"`
	CreationTimestamp *metav1.Time `json:"creation_timestamp,omitempty" yaml:"creation_timestamp,omitempty"`
	ActivityTimestamp *metav1.Time `json:"activity_timestamp,omitempty" yaml:"activity_timestamp,omitempty"`
	OpenshiftVersion  string       `json:"openshiftVersion,omitempty" yaml:"openshiftVersion,omitempty"`
	CloudProvider     string       `json:"cloudProvider,omitempty" yaml:"cloudProvider,omitempty"`
	Status            string       `json:"status,omitempty" yaml:"status,omitempty"`
	IsManagedCluster  bool         `json:"isManagedCluster,omitempty" yaml:"isManagedCluster,omitempty"`

	ProviderConnections []corev1.ObjectReference `json:"providerConnections"`
	Credential          corev1.ObjectReference   `json:"credential,omitempty" yaml:"credential,omitempty"`
}

// DiscoveredClusterStatus defines the observed state of DiscoveredCluster
type DiscoveredClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// DiscoveredCluster is the Schema for the discoveredclusters API
type DiscoveredCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DiscoveredClusterSpec   `json:"spec,omitempty"`
	Status DiscoveredClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DiscoveredClusterList contains a list of DiscoveredCluster
type DiscoveredClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DiscoveredCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DiscoveredCluster{}, &DiscoveredClusterList{})
}

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DiscoveredClusterInfo defines the desired state of DiscoveredCluster
type DiscoveredClusterInfo struct {
	Name              string       `json:"name,omitempty" yaml:"name,omitempty"`
	Console           string       `json:"console,omitempty" yaml:"console,omitempty"`
	APIURL            string       `json:"apiUrl,omitempty" yaml:"apiUrl,omitempty"`
	CreationTimestamp *metav1.Time `json:"creation_timestamp,omitempty" yaml:"creation_timestamp,omitempty"`
	ActivityTimestamp *metav1.Time `json:"activity_timestamp,omitempty" yaml:"activity_timestamp,omitempty"`
	OpenshiftVersion  string       `json:"openshiftVersion" yaml:"openshiftVersion"`
	Region            string       `json:"region,omitempty" yaml:"region,omitempty"`
	CloudProvider     string       `json:"cloudProvider,omitempty" yaml:"cloudProvider,omitempty"`
	HealthState       string       `json:"healthState,omitempty" yaml:"healthState,omitempty"`
	State             string       `json:"state,omitempty" yaml:"state,omitempty"`
	Product           string       `json:"product,omitempty" yaml:"product,omitempty"`
	IsManagedCluster  bool         `json:"isManagedCluster,omitempty" yaml:"isManagedCluster,omitempty"`
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

	Info   DiscoveredClusterInfo   `json:"info,omitempty"`
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

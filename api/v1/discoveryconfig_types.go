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

// Filter ...
type Filter struct {
	// LastActive is the last active in days of clusters to discover, determined by activity timestamp
	LastActive int `json:"lastActive,omitempty"`

	// OpenShiftVersions is the list of release versions of OpenShift
	// +optional
	OpenShiftVersions []string `json:"openShiftVersions,omitempty"`
}

// DiscoveryConfigSpec defines the desired state of DiscoveryConfig
type DiscoveryConfigSpec struct {
	// ProviderConnections are the list of secrets containing credentials to connect to the OCM api on behalf of a user
	// +optional
	ProviderConnections []string `json:"providerConnections,omitempty"`

	// Sets restrictions on what kind of clusters to discover
	// +optional
	Filters Filter `json:"filters,omitempty"`
}

// DiscoveryConfigStatus defines the observed state of DiscoveryConfig
type DiscoveryConfigStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// DiscoveryConfig is the Schema for the discoveryconfigs API
type DiscoveryConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DiscoveryConfigSpec   `json:"spec,omitempty"`
	Status DiscoveryConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DiscoveryConfigList contains a list of DiscoveryConfig
type DiscoveryConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DiscoveryConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DiscoveryConfig{}, &DiscoveryConfigList{})
}

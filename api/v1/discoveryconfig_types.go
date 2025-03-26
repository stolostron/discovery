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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Filter defines the criteria for discovering clusters based on specific attributes.
type Filter struct {
	// ClusterTypes is the list of cluster types to discover. These types represent the platform
	// the cluster is running on, such as OpenShift Container Platform (OCP), Azure Red Hat OpenShift (ARO),
	// or others.
	// +optional
	ClusterTypes []string `json:"clusterTypes,omitempty"`

	// InfrastructureProviders is the list of infrastructure providers to discover. This can be
	// a list of cloud providers or platforms (e.g., AWS, Azure, GCP) where clusters might be running.
	// +optional
	InfrastructureProviders []string `json:"infrastructureProviders,omitempty"`

	// LastActive is the last active in days of clusters to discover, determined by activity timestamp
	// +optional
	LastActive int `json:"lastActive,omitempty"`

	// OpenShiftVersions is the list of release versions of OpenShift of the form "<Major>.<Minor>"
	// +optional
	OpenShiftVersions []Semver `json:"openShiftVersions,omitempty"`

	// Regions is the list of regions where OpenShift clusters are located. This helps in filtering
	// clusters based on geographic location or data center region, useful for compliance or latency
	// requirements.
	// +optional
	Regions []string `json:"regions,omitempty"`
}

// Semver represents a partial semver string with the major and minor version
// in the form "<Major>.<Minor>". For example: "4.5"
// +kubebuilder:validation:Pattern="^(?:0|[1-9]\\d*)\\.(?:0|[1-9]\\d*)$"
type Semver string

// DiscoveryConfigSpec defines the desired state of DiscoveryConfig
type DiscoveryConfigSpec struct {
	// Credential is the secret containing credentials to connect to the OCM api on behalf of a user
	// +required
	Credential string `json:"credential"`

	// Sets restrictions on what kind of clusters to discover
	// +optional
	Filters Filter `json:"filters,omitempty"`
}

// DiscoveryConfigStatus defines the observed state of DiscoveryConfig
type DiscoveryConfigStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:storageversion

// DiscoveryConfig is the Schema for the discoveryconfigs API
type DiscoveryConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DiscoveryConfigSpec   `json:"spec,omitempty"`
	Status DiscoveryConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DiscoveryConfigList contains a list of DiscoveryConfig
type DiscoveryConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DiscoveryConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DiscoveryConfig{}, &DiscoveryConfigList{})
}

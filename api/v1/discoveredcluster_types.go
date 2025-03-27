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

// Constants for labels and annotations used in discovery and management operations.
const (
	// AutoDetectLabels is used to specify automatic detection.
	AutoDetectLabels = "auto-detect"

	// ClusterMonitoringLabel is the label indicating cluster monitoring.
	ClusterMonitoringLabel = "openshift.io/cluster-monitoring"

	// CreatedViaAnnotation is the annotation indicating the creation method.
	CreatedViaAnnotation = "open-cluster-management/created-via"

	// ImportStrategyAnnotation is the annotation indicating the import strategy.
	ImportStrategyAnnotation = "discovery.open-cluster-management.io/import-strategy"

	// ImportCleanUpFinalizer is a cleanup finalizer associated with resources created by the discovery operator.
	ImportCleanUpFinalizer = "discovery.open-cluster-management.io/import-cleanup"
)

// DiscoveredClusterSpec defines the desired state of DiscoveredCluster
type DiscoveredClusterSpec struct {
	// ActivityTimestamp records the last observed activity of the cluster.
	ActivityTimestamp *metav1.Time `json:"activityTimestamp,omitempty" yaml:"activityTimestamp,omitempty"`

	// APIURL is the endpoint used to access the cluster's API server.
	APIURL string `json:"apiUrl" yaml:"apiUrl"`

	// CloudProvider specifies the cloud provider where the cluster is hosted (e.g., AWS, Azure, GCP).
	CloudProvider string `json:"cloudProvider,omitempty" yaml:"cloudProvider,omitempty"`

	// Console provides the URL of the cluster's web-based console.
	Console string `json:"console,omitempty" yaml:"console,omitempty"`

	// CreationTimestamp marks when the cluster was initially discovered.
	CreationTimestamp *metav1.Time `json:"creationTimestamp,omitempty" yaml:"creationTimestamp,omitempty"`

	// Credential references the Kubernetes secret containing authentication details for the cluster.
	Credential corev1.ObjectReference `json:"credential,omitempty" yaml:"credential,omitempty"`

	// DisplayName is a human-readable name assigned to the cluster.
	DisplayName string `json:"displayName" yaml:"displayName"`

	// ImportAsManagedCluster determines whether the discovered cluster should be automatically imported as a managed cluster.
	// +kubebuilder:default:=false
	ImportAsManagedCluster bool `json:"importAsManagedCluster,omitempty" yaml:"importAsManagedCluster,omitempty"`

	// IsManagedCluster indicates whether the cluster is currently managed.
	IsManagedCluster bool `json:"isManagedCluster" yaml:"isManagedCluster"`

	// Name represents the unique identifier of the discovered cluster.
	Name string `json:"name" yaml:"name"`

	// OCPClusterID contains the unique identifier assigned by OpenShift to the cluster.
	OCPClusterID string `json:"ocpClusterId,omitempty" yaml:"ocpClusterId,omitempty"`

	// OpenshiftVersion specifies the OpenShift version running on the cluster.
	OpenshiftVersion string `json:"openshiftVersion,omitempty" yaml:"openshiftVersion,omitempty"`

	// Owner identifies the owner or organization responsible for the cluster.
	Owner string `json:"owner,omitempty" yaml:"owner,omitempty"`

	// RHOCMClusterID contains the cluster ID from Red Hat OpenShift Cluster Manager.
	RHOCMClusterID string `json:"rhocmClusterId,omitempty" yaml:"rhocmClusterId,omitempty"`

	// Region specifies the geographical region where the cluster is deployed.
	Region string `json:"region,omitempty" yaml:"region,omitempty"`

	// Status represents the current state of the discovered cluster (e.g Active, Stale).
	Status string `json:"status,omitempty" yaml:"status,omitempty"`

	// Type defines the type of cluster, such as OpenShift, Kubernetes, or a specific managed service type.
	Type string `json:"type" yaml:"type"`
}

type DiscoveredClusterCondition struct {
	// Type is the type of the discovered cluster condition.
	// +required
	Type DiscoveredClusterConditionType `json:"type,omitempty"`

	// Status is the status of the condition. One of True, False, Unknown.
	// +required
	Status metav1.ConditionStatus `json:"status,omitempty"`

	// The last time this condition was updated.
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`

	// LastTransitionTime is the last time the condition changed from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
}

type DiscoveredClusterConditionType string

// These are valid conditions of the multiclusterengine.
const (
	DiscoveredClusterActive   DiscoveredClusterConditionType = "Available"
	DiscoveredClusterReserved DiscoveredClusterConditionType = "Reserved"
	DiscoveredClusterStale    DiscoveredClusterConditionType = "Stale"
)

// DiscoveredClusterStatus defines the observed state of DiscoveredCluster
type DiscoveredClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Conditions []DiscoveredClusterCondition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// +kubebuilder:printcolumn:name="Display Name",type="string",JSONPath=".spec.displayName",description="Human-readable name assigned to the cluster"
// +kubebuilder:printcolumn:name="Cloud Provider",type="string",JSONPath=".spec.cloudProvider",description="Cloud provider where the cluster is hosted (e.g., AWS, Azure, GCP)"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".spec.status",description="Current state of the discovered cluster (e.g Active, Stale)"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
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
	if a.Spec.APIURL != b.Spec.APIURL ||
		a.Spec.ActivityTimestamp.Truncate(time.Second) != b.Spec.ActivityTimestamp.Truncate(time.Second) ||
		a.Spec.CloudProvider != b.Spec.CloudProvider ||
		a.Spec.Console != b.Spec.Console ||
		a.Spec.CreationTimestamp.Truncate(time.Second) != b.Spec.CreationTimestamp.Truncate(time.Second) ||
		a.Spec.Credential != b.Spec.Credential ||
		a.Spec.DisplayName != b.Spec.DisplayName ||
		a.Spec.ImportAsManagedCluster != b.Spec.ImportAsManagedCluster ||
		a.Spec.IsManagedCluster != b.Spec.IsManagedCluster ||
		a.Spec.Name != b.Spec.Name ||
		a.Spec.OpenshiftVersion != b.Spec.OpenshiftVersion ||
		a.Spec.Owner != b.Spec.Owner ||
		a.Spec.Region != b.Spec.Region ||
		a.Spec.Status != b.Spec.Status ||
		a.Spec.Type != b.Spec.Type {
		return false
	}
	return true
}

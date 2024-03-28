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

	discovery "github.com/stolostron/discovery/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var r = &DiscoveredClusterReconciler{
	Client: fake.NewClientBuilder().Build(),
}

func Test_ApplyDefaultImportStrategy(t *testing.T) {
	tests := []struct {
		name string
		obj  metav1.Object
		want bool
	}{
		{
			name: "should reconcile ROSA cluster",
			obj: &discovery.DiscoveredCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
				},
				Spec: discovery.DiscoveredClusterSpec{
					Type: "ROSA",
				},
			},
			want: true,
		},
		{
			name: "should reconcile OCP cluster",
			obj: &discovery.DiscoveredCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
				},
				Spec: discovery.DiscoveredClusterSpec{
					Type: "OCP",
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r.ShouldReconcile(tt.obj)
		})
	}
}

func Test_CreateAutoImportSecret(t *testing.T) {
	tests := []struct {
		name      string
		clusterID string
		nn        types.NamespacedName
		token     string
		want      bool
	}{
		{
			name:      "should create auto import Secret object",
			clusterID: "BASS9-ADJAN-349AS-923SD",
			nn: types.NamespacedName{
				Name:      "foo",
				Namespace: "bar",
			},
			token: "sample-token",
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := r.CreateAutoImportSecret(tt.nn, tt.clusterID, tt.token)

			if got := s.GetName() != tt.nn.Name; got {
				t.Errorf("CreateAutoImportSecret(tt.nn) = want %v, got %v", got, tt.want)
			}
		})
	}
}

func Test_CreateKlusterletAddonConfig(t *testing.T) {
	tests := []struct {
		name string
		nn   types.NamespacedName
		want bool
	}{
		{
			name: "should create KlusterletAddonConfig object",
			nn: types.NamespacedName{
				Name:      "foo",
				Namespace: "bar",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kca := r.CreateKlusterletAddonConfig(tt.nn)

			if got := kca.GetName() != tt.nn.Name; got {
				t.Errorf("CreateKlusterletAddonConfig(tt.nn) = want %v, got %v", got, tt.want)
			}

			if got := kca.GetNamespace() != tt.nn.Namespace; got {
				t.Errorf("CreateKlusterletAddonConfig(tt.nn) = want %v, got %v", got, tt.want)
			}

			if got := kca.Spec.ClusterLabels == nil; got {
				t.Errorf("CreateKlusterletAddonConfig(tt.nn).SearchCollectorConfig.Enabled = want %v, got %v", got,
					tt.want)
			}

			if got := kca.Spec.ApplicationManagerConfig.Enabled; !got {
				t.Errorf("CreateKlusterletAddonConfig(tt.nn).ApplicationManagerConfig.Enabled = want %v, got %v", got,
					tt.want)
			}

			if got := kca.Spec.CertPolicyControllerConfig.Enabled; !got {
				t.Errorf("CreateKlusterletAddonConfig(tt.nn).CertPolicyControllerConfig.Enabled = want %v, got %v", got,
					tt.want)
			}

			if got := kca.Spec.IAMPolicyControllerConfig.Enabled; !got {
				t.Errorf("CreateKlusterletAddonConfig(tt.nn).IAMPolicyControllerConfig.Enabled = want %v, got %v", got,
					tt.want)
			}

			if got := kca.Spec.PolicyController.Enabled; !got {
				t.Errorf("CreateKlusterletAddonConfig(tt.nn).PolicyController.Enabled = want %v, got %v", got, tt.want)
			}

			if got := kca.Spec.SearchCollectorConfig.Enabled; !got {
				t.Errorf("CreateKlusterletAddonConfig(tt.nn).SearchCollectorConfig.Enabled = want %v, got %v", got,
					tt.want)
			}
		})
	}
}

func Test_CreateManagedCluster(t *testing.T) {
	tests := []struct {
		name string
		nn   types.NamespacedName
		want bool
	}{
		{
			name: "should create ManagedCluster object",
			nn: types.NamespacedName{
				Name:      "foo",
				Namespace: "bar",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := r.CreateManagedCluster(tt.nn)

			if got := mc.GetName() != tt.nn.Name; got {
				t.Errorf("CreateManagedCluster(tt.nn) = want %v, got %v", got, tt.want)
			}
		})
	}
}

func Test_CreateNamespaceForDiscoveredCluster(t *testing.T) {
	tests := []struct {
		name string
		dc   *discovery.DiscoveredCluster
		want bool
	}{
		{
			name: "should create Namespace for DiscoveredCluster",
			dc: &discovery.DiscoveredCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
				},
				Spec: discovery.DiscoveredClusterSpec{
					DisplayName: "foo",
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := r.CreateNamespaceForDiscoveredCluster(*tt.dc)

			if got := ns.GetName() != tt.dc.Spec.DisplayName; got {
				t.Errorf("CreateNamespaceForDiscoveredCluster(tt.dc) = want %v, got %v", got, tt.want)
			}
		})
	}
}

// func Test_EnsureAutoImportSecret(t *testing.T) {
// 	tests := []struct {
// 		name   string
// 		config *discovery.DiscoveryConfig
// 		dc     *discovery.DiscoveredCluster
// 		want   bool
// 	}{
// 		{
// 			name:   "should ensure auto-import Secret is created",
// 			config: &discovery.DiscoveryConfig{},
// 			dc: &discovery.DiscoveredCluster{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name:      "foo",
// 					Namespace: "bar",
// 				},
// 				Spec: discovery.DiscoveredClusterSpec{
// 					Type: "ROSA",
// 				},
// 			},
// 			want: true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			r.EnsureAutoImportSecret(context.TODO(), *tt.dc, *tt.config)
// 		})
// 	}
// }

// func Test_EnsureKlusterletAddonConfig(t *testing.T) {
// 	r := DiscoveredClusterReconciler{}

// 	tests := []struct {
// 		name string
// 		dc   *discovery.DiscoveredCluster
// 		want bool
// 	}{
// 		{
// 			name: "should ensure KlusterletAddonConfig created",
// 			dc: &discovery.DiscoveredCluster{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name:      "foo",
// 					Namespace: "bar",
// 				},
// 				Spec: discovery.DiscoveredClusterSpec{
// 					DisplayName: "foo",
// 					Type:        "ROSA",
// 				},
// 			},
// 			want: true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			r.EnsureKlusterletAddonConfig(context.TODO(), *tt.dc)
// 		})
// 	}
// }

// func Test_EnsureManagedCluster(t *testing.T) {
// 	tests := []struct {
// 		name string
// 		dc   *discovery.DiscoveredCluster
// 		want bool
// 	}{
// 		{
// 			name: "should ensure ManagedCluster created",
// 			dc: &discovery.DiscoveredCluster{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name:      "foo",
// 					Namespace: "bar",
// 				},
// 				Spec: discovery.DiscoveredClusterSpec{
// 					DisplayName: "foo",
// 					Type:        "ROSA",
// 				},
// 			},
// 			want: true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			r.EnsureManagedCluster(context.TODO(), *tt.dc)
// 		})
// 	}
// }

// func Test_EnsureNamespaceForDiscoveredCluster(t *testing.T) {
// 	tests := []struct {
// 		name string
// 		dc   *discovery.DiscoveredCluster
// 		want bool
// 	}{
// 		{
// 			name: "should ensure namespace created for DiscoveredCluster",
// 			dc: &discovery.DiscoveredCluster{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name:      "foo",
// 					Namespace: "bar",
// 				},
// 				Spec: discovery.DiscoveredClusterSpec{
// 					DisplayName: "foo",
// 					Type:        "ROSA",
// 				},
// 			},
// 			want: true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			r.EnsureNamespaceForDiscoveredCluster(context.TODO(), *tt.dc)
// 		})
// 	}
// }

// func Test_EnsureFinalizerRemovedFromManagedCluster(t *testing.T) {
// 	tests := []struct {
// 		name string
// 		dc   *discovery.DiscoveredCluster
// 		want bool
// 	}{
// 		{
// 			name: "should ensure finalizers are removed from ManagedCluster",
// 			dc: &discovery.DiscoveredCluster{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name:        "foo",
// 					Namespace:   "bar",
// 					Annotations: map[string]string{},
// 				},
// 				Spec: discovery.DiscoveredClusterSpec{
// 					Type: "OCP",
// 				},
// 			},
// 			want: false,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			r.EnsureFinalizerRemovedFromManagedCluster(context.TODO(), *tt.dc)
// 		})
// 	}
// }

func Test_Reconciler_ShouldReconcile(t *testing.T) {
	tests := []struct {
		name string
		obj  metav1.Object
		want bool
	}{
		{
			name: "should reconcile ROSA cluster",
			obj: &discovery.DiscoveredCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
				},
				Spec: discovery.DiscoveredClusterSpec{
					Type: "ROSA",
				},
			},
			want: true,
		},
		{
			name: "should reconcile OCP cluster",
			obj: &discovery.DiscoveredCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
				},
				Spec: discovery.DiscoveredClusterSpec{
					Type: "OCP",
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r.ShouldReconcile(tt.obj)
		})
	}
}

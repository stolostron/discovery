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
	"context"
	"testing"

	discovery "github.com/stolostron/discovery/api/v1"
	agentv1 "github.com/stolostron/klusterlet-addon-controller/pkg/apis/agent/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	clusterapiv1 "open-cluster-management.io/api/cluster/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var r = &DiscoveredClusterReconciler{
	Client: fake.NewClientBuilder().Build(),
}

func registerScheme() {
	clusterapiv1.AddToScheme(scheme.Scheme)
	discovery.AddToScheme(scheme.Scheme)
	agentv1.SchemeBuilder.AddToScheme(scheme.Scheme)
}

func Test_DiscoveredCluster_Reconciler_Reconcile(t *testing.T) {
	tests := []struct {
		name   string
		config *discovery.DiscoveryConfig
		dc     *discovery.DiscoveredCluster
		ns     *corev1.Namespace
		s      *corev1.Secret
		req    ctrl.Request
	}{
		{
			name: "should create auto import Secret object",
			config: &discovery.DiscoveryConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "discovery",
					Namespace: "discovery",
				},
				Spec: discovery.DiscoveryConfigSpec{
					Credential: "fake-admin",
				},
			},
			ns: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "discovery",
				},
			},
			dc: &discovery.DiscoveredCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "310ac28e-b69b-447b-a51f-08e967cff1ee",
					Namespace: "discovery",
				},
				Spec: discovery.DiscoveredClusterSpec{
					DisplayName:      "fake-cluster",
					EnableAutoImport: true,
					RHOCMClusterID:   "349bcdc1dd6a44f3a1a136b2f98a69ca",
					Type:             "ROSA",
				},
			},
			req: ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      "310ac28e-b69b-447b-a51f-08e967cff1ee",
					Namespace: "discovery",
				},
			},
			s: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "fake-admin",
					Namespace: "discovery",
				},
				Data: map[string][]byte{
					"ocmAPIToken": []byte("fake-token"),
				},
			},
		},
	}

	registerScheme()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			DiscoveryConfig = types.NamespacedName{
				Name: tt.config.Name, Namespace: tt.config.Namespace,
			}

			ns := &corev1.Namespace{}
			mc := &clusterapiv1.ManagedCluster{}
			kac := &agentv1.KlusterletAddonConfig{}
			s := &corev1.Secret{}

			defer func() {
				r.Delete(context.TODO(), kac)
				r.Delete(context.TODO(), s)
				r.Delete(context.TODO(), mc)
				r.EnsureFinalizerRemovedFromManagedCluster(context.TODO(), *tt.dc)
				r.Delete(context.TODO(), ns)

				r.Delete(context.TODO(), tt.dc)
				r.Delete(context.TODO(), tt.config)
				r.Delete(context.TODO(), tt.s)
				r.Delete(context.TODO(), tt.ns)
			}()

			if err := r.Create(context.TODO(), tt.ns); err != nil {
				t.Errorf("failed to create Namespace: %v", err)
			}

			if err := r.Create(context.TODO(), tt.config); err != nil {
				t.Errorf("failed to create DiscoveryConfig: %v", err)
			}

			if err := r.Create(context.TODO(), tt.s); err != nil {
				t.Errorf("failed to create Secret: %v", err)
			}

			if err := r.Create(context.TODO(), tt.dc); err != nil {
				t.Errorf("failed to create DiscoveredCluster: %v", err)
			}

			if _, err := r.Reconcile(context.TODO(), tt.req); err != nil {
				t.Errorf("error: %v", err)
			}

			if err := r.Get(context.TODO(), types.NamespacedName{Name: tt.dc.Spec.DisplayName}, ns); err != nil {
				t.Errorf("failed to get Namespace: %v", err)
			}

			if err := r.Get(context.TODO(), types.NamespacedName{Name: tt.dc.Spec.DisplayName}, mc); err != nil {
				t.Errorf("failed to get ManagedCluster: %v", err)
			}

			if err := r.Get(context.TODO(), types.NamespacedName{Name: tt.dc.Spec.DisplayName,
				Namespace: tt.dc.Spec.DisplayName}, kac); err != nil {
				t.Errorf("failed to get KlusterletAddonConfig: %v", err)
			}

			if err := r.Get(context.TODO(), types.NamespacedName{Name: "auto-import-secret",
				Namespace: tt.dc.Spec.DisplayName}, s); err != nil {
				t.Errorf("failed to get auto-import Secret: %v", err)
			}
		})
	}
}
func Test_Reconciler_CreateAutoImportSecret(t *testing.T) {
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

func Test_Reconciler_CreateKlusterletAddonConfig(t *testing.T) {
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
			kac := r.CreateKlusterletAddonConfig(tt.nn)

			if got := kac.GetName() != tt.nn.Name; got {
				t.Errorf("CreateKlusterletAddonConfig(tt.nn) = want %v, got %v", got, tt.want)
			}

			if got := kac.GetNamespace() != tt.nn.Namespace; got {
				t.Errorf("CreateKlusterletAddonConfig(tt.nn) = want %v, got %v", got, tt.want)
			}

			if got := kac.Spec.ClusterLabels == nil; got {
				t.Errorf("CreateKlusterletAddonConfig(tt.nn).SearchCollectorConfig.Enabled = want %v, got %v", got,
					tt.want)
			}

			if got := kac.Spec.ApplicationManagerConfig.Enabled; !got {
				t.Errorf("CreateKlusterletAddonConfig(tt.nn).ApplicationManagerConfig.Enabled = want %v, got %v", got,
					tt.want)
			}

			if got := kac.Spec.CertPolicyControllerConfig.Enabled; !got {
				t.Errorf("CreateKlusterletAddonConfig(tt.nn).CertPolicyControllerConfig.Enabled = want %v, got %v", got,
					tt.want)
			}

			if got := kac.Spec.IAMPolicyControllerConfig.Enabled; !got {
				t.Errorf("CreateKlusterletAddonConfig(tt.nn).IAMPolicyControllerConfig.Enabled = want %v, got %v", got,
					tt.want)
			}

			if got := kac.Spec.PolicyController.Enabled; !got {
				t.Errorf("CreateKlusterletAddonConfig(tt.nn).PolicyController.Enabled = want %v, got %v", got, tt.want)
			}

			if got := kac.Spec.SearchCollectorConfig.Enabled; !got {
				t.Errorf("CreateKlusterletAddonConfig(tt.nn).SearchCollectorConfig.Enabled = want %v, got %v", got,
					tt.want)
			}
		})
	}
}

func Test_Reconciler_CreateManagedCluster(t *testing.T) {
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

func Test_Reconciler_CreateNamespaceForDiscoveredCluster(t *testing.T) {
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

func Test_Reconciler_EnsureAutoImportSecret(t *testing.T) {
	tests := []struct {
		name   string
		config *discovery.DiscoveryConfig
		dc     *discovery.DiscoveredCluster
		want   bool
	}{
		{
			name: "should ensure auto-import Secret is created",
			config: &discovery.DiscoveryConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "discovery",
					Namespace: "discovery",
				},
				Spec: discovery.DiscoveryConfigSpec{
					Credential: "admin",
				},
			},
			dc: &discovery.DiscoveredCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
				},
				Spec: discovery.DiscoveredClusterSpec{
					DisplayName: "foo",
					Type:        "ROSA",
				},
			},
			want: true,
		},
	}

	registerScheme()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: tt.dc.Spec.DisplayName}}
			s1 := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: tt.config.Spec.Credential,
				Namespace: tt.config.GetNamespace()}, Data: map[string][]byte{"ocmAPIToken": []byte("fake-token")}}
			s2 := &corev1.Secret{}

			defer func() {
				if err := r.Delete(context.TODO(), s1); err != nil {
					t.Errorf("failed to delete Secret: %v", err)
				}

				if err := r.Delete(context.TODO(), s2); err != nil {
					t.Errorf("failed to delete Secret: %v", err)
				}

				if err := r.Delete(context.TODO(), ns); err != nil {
					t.Errorf("failed to delete Namespace: %v", err)
				}
			}()

			if err := r.Create(context.TODO(), ns); err != nil {
				t.Errorf("failed to create namespace: %v", err)
			}

			if err := r.Create(context.TODO(), s1); err != nil {
				t.Errorf("failed to create Secret: %v", err)
			}

			if _, err := r.EnsureAutoImportSecret(context.TODO(), *tt.dc, *tt.config); err != nil {
				t.Errorf("failed to ensure auto import Secret created: %v", err)
			}

			if err := r.Get(context.TODO(), types.NamespacedName{Name: "auto-import-secret",
				Namespace: tt.dc.Spec.DisplayName}, s2); err != nil {
				t.Errorf("failed to get auto import secret: %v", err)
			}
		})
	}
}

func Test_Reconciler_EnsureKlusterletAddonConfig(t *testing.T) {
	tests := []struct {
		name string
		dc   *discovery.DiscoveredCluster
		want bool
	}{
		{
			name: "should ensure KlusterletAddonConfig created",
			dc: &discovery.DiscoveredCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
				},
				Spec: discovery.DiscoveredClusterSpec{
					DisplayName: "foo",
					Type:        "ROSA",
				},
			},
			want: true,
		},
	}

	registerScheme()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: tt.dc.Spec.DisplayName}}
			kac := &agentv1.KlusterletAddonConfig{}

			defer func() {
				if err := r.Delete(context.TODO(), kac); err != nil {
					t.Errorf("failed to delete KlusterletAddonConfig: %v", err)
				}

				if err := r.Delete(context.TODO(), ns); err != nil {
					t.Errorf("failed to delete Namespace: %v", err)
				}
			}()

			if err := r.Create(context.TODO(), ns); err != nil {
				t.Errorf("failed to create Namespace: %v, err: %v", ns.GetName(), err)
			}

			if _, err := r.EnsureKlusterletAddonConfig(context.TODO(), *tt.dc); err != nil {
				t.Errorf("failed to create ManagedCluster resource: %v", err)
			}

			if err := r.Get(context.TODO(), types.NamespacedName{Name: tt.dc.Spec.DisplayName,
				Namespace: tt.dc.Spec.DisplayName}, kac); err != nil {
				t.Errorf("failed to get KlusterletAddonConfig resource: %v", err)
			}
		})
	}
}

func Test_Reconciler_EnsureManagedCluster(t *testing.T) {
	tests := []struct {
		name string
		dc   *discovery.DiscoveredCluster
		want bool
	}{
		{
			name: "should ensure ManagedCluster created",
			dc: &discovery.DiscoveredCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
				},
				Spec: discovery.DiscoveredClusterSpec{
					DisplayName: "foo",
					Type:        "ROSA",
				},
			},
			want: true,
		},
	}

	registerScheme()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: tt.dc.Spec.DisplayName}}
			mc := &clusterapiv1.ManagedCluster{}

			defer func() {
				if err := r.Delete(context.TODO(), mc); err != nil {
					t.Errorf("failed to delete ManagedCluster: %v", err)
				}

				if err := r.Delete(context.TODO(), ns); err != nil {
					t.Errorf("failed to delete Namespace: %v", err)
				}

				if _, err := r.EnsureFinalizerRemovedFromManagedCluster(context.TODO(), *tt.dc); err != nil {
					t.Errorf("failed to ensure finalizer removed from ManagedCluster: %v", err)
				}

				if err := r.Get(context.TODO(), types.NamespacedName{Name: tt.dc.Spec.DisplayName}, mc); err == nil {
					t.Errorf("ManagedCluster still exist: %v", err)
				}
			}()

			if err := r.Create(context.TODO(), ns); err != nil {
				t.Errorf("failed to create namespace: %v", ns.GetName())
			}

			if _, err := r.EnsureManagedCluster(context.TODO(), *tt.dc); err != nil {
				t.Errorf("failed to create ManagedCluster resource: %v", err)
			}

			if err := r.Get(context.TODO(), types.NamespacedName{Name: tt.dc.Spec.DisplayName}, mc); err != nil {
				t.Errorf("failed to get ManagedCluster resource: %v", err)
			}
		})
	}
}

func Test_Reconciler_EnsureNamespaceForDiscoveredCluster(t *testing.T) {
	tests := []struct {
		name string
		dc   *discovery.DiscoveredCluster
		want bool
	}{
		{
			name: "should ensure namespace created for DiscoveredCluster",
			dc: &discovery.DiscoveredCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
				},
				Spec: discovery.DiscoveredClusterSpec{
					DisplayName: "foo",
					Type:        "ROSA",
				},
			},
			want: true,
		},
	}

	registerScheme()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &corev1.Namespace{}

			defer func() {
				if err := r.Delete(context.TODO(), ns); err != nil {
					t.Errorf("failed to delete Namespace: %v", err)
				}
			}()

			if _, err := r.EnsureNamespaceForDiscoveredCluster(context.TODO(), *tt.dc); err != nil {
				t.Errorf("failed to create Namespace for DiscoveredCluster: %v", err)
			}

			if err := r.Get(context.TODO(), types.NamespacedName{Name: tt.dc.Spec.DisplayName}, ns); err != nil {
				t.Errorf("failed to get Namespace for DiscoveredCluster: %v", err)
			}
		})
	}
}

// func Test_Reconciler_SetupWithManager(t *testing.T) {
// 	tests := []struct {
// 		name string
// 		want bool
// 	}{
// 		{
// 			name: "should setup reconciler with manager",
// 			want: true,
// 		},
// 	}

// 	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{Scheme: scheme.Scheme})
// 	if err != nil {
// 		t.Errorf("failed to create manager: %v", err)
// 	}

// 	registerScheme()
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := r.SetupWithManager(mgr); err != nil {
// 				t.Errorf("failed to setup manager: %v", err)
// 			}
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
			name: "should not reconcile OCP cluster",
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

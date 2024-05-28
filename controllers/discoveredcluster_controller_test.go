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
	"fmt"
	"os"
	"path/filepath"
	"testing"

	discovery "github.com/stolostron/discovery/api/v1"
	agentv1 "github.com/stolostron/klusterlet-addon-controller/pkg/apis/agent/v1"
	corev1 "k8s.io/api/core/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes/scheme"
	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
	clusterapiv1 "open-cluster-management.io/api/cluster/v1"
	clusterapiv1beta1 "open-cluster-management.io/api/cluster/v1beta1"
	clusterapiv1beta2 "open-cluster-management.io/api/cluster/v1beta2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var r = &DiscoveredClusterReconciler{
	Client: fake.NewClientBuilder().Build(),
}

var crdDir = "../test/resources/"

func registerScheme() {
	clusterapiv1.AddToScheme(scheme.Scheme)
	discovery.AddToScheme(scheme.Scheme)
	agentv1.SchemeBuilder.AddToScheme(scheme.Scheme)
	apiextv1.AddToScheme(scheme.Scheme)
	addonv1alpha1.AddToScheme(scheme.Scheme)
	clusterapiv1beta1.AddToScheme(scheme.Scheme)
	clusterapiv1beta2.AddToScheme(scheme.Scheme)
}

func deployCRDs(directory string) error {
	files, err := os.ReadDir(directory)

	if err != nil {
		return fmt.Errorf("failed to read directory %s: %v", directory, err)
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		filePath := filepath.Join(directory, f.Name())
		crdFile, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read CRD YAML file: %v", err)
		}

		crd := &unstructured.Unstructured{}
		if err := yaml.Unmarshal(crdFile, crd); err != nil {
			return fmt.Errorf("failed to unmarshal CRD YAML file: %v", err)
		}

		if err := r.Get(context.TODO(), types.NamespacedName{Name: crd.GetName()}, crd); err != nil {
			if apierrors.IsNotFound(err) {
				if err := r.Create(context.TODO(), crd); err != nil {
					return fmt.Errorf("failed to create CRD: %v", err)
				}

			} else {
				return fmt.Errorf("failed to get CRD: %v", err)
			}
		}
	}
	return nil
}

func deleteClusterManagementAddOns() error {
	addonNames := []string{"cluster-proxy", "managed-serviceaccount", "work-manager"}
	for _, addon := range addonNames {
		cma := addonv1alpha1.ClusterManagementAddOn{ObjectMeta: metav1.ObjectMeta{Name: addon}}
		if err := r.Delete(context.TODO(), &cma); err != nil {
			return err
		}
	}

	return nil
}

func deployClusterManagementAddOns() error {
	addonNames := []string{"cluster-proxy", "managed-serviceaccount", "work-manager"}
	for _, addon := range addonNames {
		cma := addonv1alpha1.ClusterManagementAddOn{
			ObjectMeta: metav1.ObjectMeta{Name: addon},
			Spec: addonv1alpha1.ClusterManagementAddOnSpec{
				InstallStrategy: addonv1alpha1.InstallStrategy{Type: ""}},
		}

		if err := r.Create(context.TODO(), &cma); err != nil {
			return err
		}
	}

	return nil
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
					Credential: corev1.ObjectReference{
						Name:      "fake-admin",
						Namespace: "discovery",
					},
					DisplayName:            "fake-cluster",
					ImportAsManagedCluster: true,
					RHOCMClusterID:         "349bcdc1dd6a44f3a1a136b2f98a69ca",
					Type:                   "ROSA",
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
	if err := deployCRDs(crdDir); err != nil {
		t.Errorf("failed to deploy CRDs: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &corev1.Namespace{}
			mc := &clusterapiv1.ManagedCluster{}
			kac := &agentv1.KlusterletAddonConfig{}
			s := &corev1.Secret{}

			defer func() {
				r.Delete(context.TODO(), kac)
				r.Delete(context.TODO(), s)
				r.Delete(context.TODO(), mc)
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

func Test_Reconciler_CreateAddOnDeploymentConfig(t *testing.T) {
	tests := []struct {
		name string
		nn   types.NamespacedName
		want bool
	}{
		{
			name: "should create AddOnDeploymentConfig object",
			nn: types.NamespacedName{
				Name:      "addon-ns-config",
				Namespace: "test-namespace",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("POD_NAMESPACE", tt.nn.Namespace)
			defer func() {
				os.Unsetenv("POD_NAMESPACE")
			}()

			adc := r.CreateAddOnDeploymentConfig(tt.nn)
			if got := adc.GetName() != tt.nn.Name && adc.GetNamespace() != tt.nn.Namespace; got {
				t.Errorf("CreateAddOnDeploymentConfig(tt.nn) = want %v, got %v", got, tt.want)
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
		name        string
		clusterType string
		nn          types.NamespacedName
		want        bool
	}{
		{
			name:        "should create ROSA ManagedCluster object",
			clusterType: "ROSA",
			nn: types.NamespacedName{
				Name:      "foo",
				Namespace: "bar",
			},
			want: true,
		},
		{
			name:        "should create MCE-HCP ManagedCluster object",
			clusterType: "MultiClusterEngineHCP",
			nn: types.NamespacedName{
				Name:      "foo",
				Namespace: "bar",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := r.CreateManagedCluster(tt.nn, tt.clusterType)

			if got := mc.GetName() != tt.nn.Name; got {
				t.Errorf("CreateManagedCluster(tt.nn) = want %v, got %v", got, tt.want)
			}
		})
	}
}

func Test_Reconciler_CreateManagedClusterSetBinding(t *testing.T) {
	tests := []struct {
		name string
		nn   types.NamespacedName
		want bool
	}{
		{
			name: "should create ManagedClusterSetBinding object",
			nn: types.NamespacedName{
				Name:      "foo",
				Namespace: "bar",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mcsb := r.CreateManagedClusterSetBinding(tt.nn)

			if got := mcsb.GetName() != tt.nn.Name; got {
				t.Errorf("CreateManagedClusterSetBinding(tt.nn) = want %v, got %v", got, tt.want)
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

func Test_Reconciler_CreatePlacement(t *testing.T) {
	tests := []struct {
		name string
		nn   types.NamespacedName
		want bool
	}{
		{
			name: "should create Placement object",
			nn: types.NamespacedName{
				Name:      "default",
				Namespace: "test-namespace",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("POD_NAMESPACE", tt.nn.Namespace)
			defer func() {
				os.Unsetenv("POD_NAMESPACE")
			}()

			placement := r.CreatePlacement(tt.nn)
			if got := placement.GetName() != tt.nn.Name && placement.GetNamespace() != tt.nn.Namespace; got {
				t.Errorf("CreatePlacement(tt.nn) = want %v, got %v", got, tt.want)
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
					Credential: corev1.ObjectReference{
						Name:      "admin",
						Namespace: "discovery",
					},
					Type: "ROSA",
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

			if _, err := r.EnsureAutoImportSecret(context.TODO(), *tt.dc); err != nil {
				t.Errorf("failed to ensure auto import Secret created: %v", err)
			}

			if err := r.Get(context.TODO(), types.NamespacedName{Name: "auto-import-secret",
				Namespace: tt.dc.Spec.DisplayName}, s2); err != nil {
				t.Errorf("failed to get auto import secret: %v", err)
			}
		})
	}
}

func Test_Reconciler_EnsureCommonResources(t *testing.T) {
	tests := []struct {
		name  string
		dc    discovery.DiscoveredCluster
		isHCP bool
		sec   corev1.Secret
		want  bool
	}{
		{
			name: "should ensure common non HCP cluster resources are created",
			dc: discovery.DiscoveredCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
				},
				Spec: discovery.DiscoveredClusterSpec{
					DisplayName: "foo",
					Credential: corev1.ObjectReference{
						Name:      "admin",
						Namespace: "bar",
					},
					RHOCMClusterID: "admin-12345",
					Type:           "ROSA",
				},
			},

			isHCP: false,
			sec: corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "admin",
					Namespace: "bar",
				},
				Data: map[string][]byte{
					"ocmAPIToken": []byte("fake-token"),
				},
			},
			want: true,
		},
		{
			name: "should ensure common HCP cluster resources are created",
			dc: discovery.DiscoveredCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
				},
				Spec: discovery.DiscoveredClusterSpec{
					DisplayName: "foo",
					Type:        "MultiClusterEngineHCP",
				},
			},
			isHCP: true,
			sec: corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "admin",
					Namespace: "bar",
				},
			},
			want: true,
		},
	}

	registerScheme()
	if err := deployCRDs(crdDir); err != nil {
		t.Errorf("failed to deploy CRDs: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if err := r.Delete(context.TODO(), &tt.sec); err != nil {
					t.Errorf("failed to delete Secret: %v", err)
				}

				if err := deleteClusterManagementAddOns(); err != nil {
					t.Errorf("failed to delete ClusterManagementAddons: %v", err)
				}
			}()

			if err := r.Create(context.TODO(), &tt.sec); err != nil {
				t.Errorf("failed to create secret: %v", err)
			}

			if err := deployClusterManagementAddOns(); err != nil {
				t.Errorf("failed to deploy clustermanagementaddons: %v", err)
			}

			if _, err := r.EnsureCommonResources(context.TODO(), &tt.dc, tt.isHCP); err != nil {
				t.Errorf("failed to ensure common resources: %v", err)
			}
		})
	}
}

func Test_Reconciler_EnsureAddOnDeploymentConfig(t *testing.T) {
	tests := []struct {
		name string
		nn   types.NamespacedName
		want bool
	}{
		{
			name: "should ensure AddOnDeploymentConfig created",
			nn: types.NamespacedName{
				Name:      "addon-ns-config",
				Namespace: "test-namespace",
			},
			want: true,
		},
	}

	registerScheme()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adc := &addonv1alpha1.AddOnDeploymentConfig{}

			os.Setenv("POD_NAMESPACE", tt.nn.Namespace)
			defer func() {
				os.Unsetenv("POD_NAMESPACE")

				if err := r.Delete(context.TODO(), adc); err != nil {
					t.Errorf("failed to delete AddOnDeploymentConfig: %v", err)
				}
			}()

			if _, err := r.EnsureAddOnDeploymentConfig(context.TODO()); err != nil {
				t.Errorf("failed to create AddOnDeploymentConfig resource: %v", err)
			}

			if err := r.Get(context.TODO(), types.NamespacedName{Name: tt.nn.Name, Namespace: tt.nn.Namespace},
				adc); err != nil {
				t.Errorf("failed to get AddOnDeploymentConfig resource: %v", err)
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
				t.Errorf("failed to create KlusterletAddonConfig resource: %v", err)
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

func Test_Reconciler_EnsureManagedClusterSetBinding(t *testing.T) {
	tests := []struct {
		name string
		nn   types.NamespacedName
		want bool
	}{
		{
			name: "should ensure ManagedClusterSetBinding created",
			nn:   types.NamespacedName{Name: "default", Namespace: "test-namespace"},
			want: true,
		},
	}

	registerScheme()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mcsb := &clusterapiv1beta2.ManagedClusterSetBinding{}

			os.Setenv("POD_NAMESPACE", tt.nn.Namespace)
			defer func() {
				os.Unsetenv("POD_NAMESPACE")
				if err := r.Delete(context.TODO(), mcsb); err != nil {
					t.Errorf("failed to delete ManagedClusterSetBinding: %v", err)
				}
			}()

			if _, err := r.EnsureManagedClusterSetBinding(context.TODO()); err != nil {
				t.Errorf("failed to create ManagedClusterSetBinding resource: %v", err)
			}

			if err := r.Get(context.TODO(), tt.nn, mcsb); err != nil {
				t.Errorf("failed to get ManagedClusterSetBinding resource: %v", err)
			}
		})
	}
}

func Test_Reconciler_EnsureMultiClusterEngineHCP(t *testing.T) {
	tests := []struct {
		name string
		dc   *discovery.DiscoveredCluster
		want bool
	}{
		{
			name: "should ensure MultiClusterEngineHCP created",
			dc: &discovery.DiscoveredCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
				},
				Spec: discovery.DiscoveredClusterSpec{
					DisplayName: "foo",
					Type:        "MultiClusterEngineHCP",
				},
			},
			want: true,
		},
	}

	registerScheme()
	if err := deployCRDs(crdDir); err != nil {
		t.Errorf("failed to deploy CRDs: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &clusterapiv1beta1.Placement{}
			adc := &addonv1alpha1.AddOnDeploymentConfig{}
			kac := &agentv1.KlusterletAddonConfig{}
			mc := &clusterapiv1.ManagedCluster{}
			mcsb := &clusterapiv1beta2.ManagedClusterSetBinding{}

			os.Setenv("POD_NAMESPACE", "test-namespace")
			defer func() {
				os.Unsetenv("POD_NAMESPACE")
				if err := deleteClusterManagementAddOns(); err != nil {
					t.Errorf("failed to delete ClusterManagementAddOns: %v", err)
				}

				if err := r.Delete(context.TODO(), mcsb); err != nil {
					t.Errorf("failed to delete ManagedClusterSetBinding: %v", err)
				}

				if err := r.Delete(context.TODO(), p); err != nil {
					t.Errorf("failed to delete Placement: %v", err)
				}

				if err := r.Delete(context.TODO(), adc); err != nil {
					t.Errorf("failed to delete AddOnDeploymentConfig: %v", err)
				}

				if err := r.Delete(context.TODO(), mc); err != nil {
					t.Errorf("failed to delete ManagedCluster: %v", err)
				}

				if err := r.Delete(context.TODO(), kac); err != nil {
					t.Errorf("failed to delete KlusterletAddOnConfig: %v", err)
				}
			}()

			if err := deployClusterManagementAddOns(); err != nil {
				t.Errorf("failed to create ClusterManagementAddOns: %v", err)
			}

			if _, err := r.EnsureMultiClusterEngineHCP(context.TODO(), tt.dc); err != nil {
				t.Errorf("failed to ensure MCE-HCP resources created: %v", err)
			}

			if err := r.Get(context.TODO(), types.NamespacedName{Name: "default",
				Namespace: os.Getenv("POD_NAMESPACE")}, mcsb); err != nil {
				t.Errorf("failed to get ManagedClusterSetBinding resource: %v", err)
			}

			if err := r.Get(context.TODO(), types.NamespacedName{Name: "default",
				Namespace: os.Getenv("POD_NAMESPACE")}, p); err != nil {
				t.Errorf("failed to get Placement resource: %v", err)
			}

			if err := r.Get(context.TODO(), types.NamespacedName{Name: "addon-ns-config",
				Namespace: os.Getenv("POD_NAMESPACE")}, adc); err != nil {
				t.Errorf("failed to get AddOnDeploymentConfig resource: %v", err)
			}

			if err := r.Get(context.TODO(), types.NamespacedName{Name: tt.dc.Spec.DisplayName}, mc); err != nil {
				t.Errorf("failed to get MCE-HCP ManagedCluster resource: %v", err)
			}

			if err := r.Get(context.TODO(), types.NamespacedName{Name: tt.dc.Spec.DisplayName,
				Namespace: tt.dc.Spec.DisplayName}, kac); err != nil {
				t.Errorf("failed to get KlusterletAddOnConfig resource: %v", err)
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

func Test_Reconciler_EnsurePlacement(t *testing.T) {
	tests := []struct {
		name string
		nn   types.NamespacedName
		want bool
	}{
		{
			name: "should ensure placement created for ClusterManagementAddOn",
			nn: types.NamespacedName{
				Name:      "default",
				Namespace: "test-namespace",
			},
			want: true,
		},
	}

	registerScheme()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			placement := &clusterapiv1beta1.Placement{}

			os.Setenv("POD_NAMESPACE", tt.nn.Namespace)
			defer func() {
				os.Unsetenv("POD_NAMESPACE")
				if err := r.Delete(context.TODO(), placement); err != nil {
					t.Errorf("failed to delete Placement: %v", err)
				}
			}()

			if _, err := r.EnsurePlacement(context.TODO()); err != nil {
				t.Errorf("failed to create Placement: %v", err)
			}

			if err := r.Get(context.TODO(), types.NamespacedName{Name: tt.nn.Name, Namespace: tt.nn.Namespace},
				placement); err != nil {
				t.Errorf("failed to get Placement: %v", err)
			}
		})
	}
}

func Test_Reconciler_EnsureROSA(t *testing.T) {
	tests := []struct {
		name string
		dc   *discovery.DiscoveredCluster
		sec  *corev1.Secret
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
					Credential: corev1.ObjectReference{
						Name:      "admin",
						Namespace: "bar",
					},
					DisplayName: "foo",
					Type:        "ROSA",
				},
			},
			sec: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "admin",
					Namespace: "bar",
				},
				Data: map[string][]byte{
					"ocmAPIToken": []byte("fake-token"),
				},
			},
			want: true,
		},
	}

	registerScheme()
	if err := deployCRDs(crdDir); err != nil {
		t.Errorf("failed to deploy CRDs: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kac := &agentv1.KlusterletAddonConfig{}
			mc := &clusterapiv1.ManagedCluster{}
			autoSec := &corev1.Secret{}

			defer func() {
				if err := r.Delete(context.TODO(), tt.sec); err != nil {
					t.Errorf("failed to delete Secret: %v", err)
				}

				if err := r.Delete(context.TODO(), mc); err != nil {
					t.Errorf("failed to delete ManagedCluster: %v", err)
				}

				if err := r.Delete(context.TODO(), kac); err != nil {
					t.Errorf("failed to delete KlusterletAddOnConfig: %v", err)
				}

				if err := r.Delete(context.TODO(), autoSec); err != nil {
					t.Errorf("failed to delete Secret: %v", err)
				}
			}()

			if err := r.Create(context.TODO(), tt.sec); err != nil {
				t.Errorf("failed to ensure ROSA resources created: %v", err)
			}

			if _, err := r.EnsureROSA(context.TODO(), tt.dc); err != nil {
				t.Errorf("failed to ensure ROSA resources created: %v", err)
			}

			if err := r.Get(context.TODO(), types.NamespacedName{Name: tt.dc.Spec.DisplayName}, mc); err != nil {
				t.Errorf("failed to get ROSA ManagedCluster resource: %v", err)
			}

			if err := r.Get(context.TODO(), types.NamespacedName{Name: tt.dc.Spec.DisplayName,
				Namespace: tt.dc.Spec.DisplayName}, kac); err != nil {
				t.Errorf("failed to get ROSA KlusterletAddOnConfig resource: %v", err)
			}

			if err := r.Get(context.TODO(), types.NamespacedName{Name: "auto-import-secret",
				Namespace: tt.dc.Spec.DisplayName}, autoSec); err != nil {
				t.Errorf("failed to get ROSA AutoImportSecret resource: %v", err)
			}
		})
	}
}

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

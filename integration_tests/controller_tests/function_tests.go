// Copyright Contributors to the Open Cluster Management project

package controller_tests

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	discoveryv1 "github.com/open-cluster-management/discovery/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
)

// Define utility constants for object names and testing timeouts/durations and intervals.
const (
	DiscoveryConfigName = "discoveryconfig"
	DiscoveryNamespace  = "open-cluster-management"
	SecretName          = "test-connection-secret"

	timeout  = time.Second * 10
	duration = time.Second * 10
	interval = time.Millisecond * 250
)

// annotated adds an annotation to modify the baseUrl used with the discoveryconfig
func annotated(dc *discoveryv1.DiscoveryConfig) *discoveryv1.DiscoveryConfig {
	dc.SetAnnotations(map[string]string{"ocmBaseURL": "http://mock-ocm-server.open-cluster-management.svc.cluster.local:3000"})
	return dc
}

var _ = Describe("DiscoveredClusterRefresh controller", func() {
	Context("When creating a DiscoveryConfig", func() {
		It("Should trigger the creation of new discovered clusters ", func() {
			By("By creating a secret with OCM credentials")
			ctx := context.Background()

			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      SecretName,
					Namespace: DiscoveryNamespace,
				},
				StringData: map[string]string{
					"metadata": "ocmAPIToken: dummytoken",
				},
			}

			Expect(k8sClient.Create(ctx, secret)).Should(Succeed())

			secretLookupKey := types.NamespacedName{Name: SecretName, Namespace: DiscoveryNamespace}
			createdSecret := &corev1.Secret{}

			// We'll need to retry getting this newly created secret, given that creation may not immediately happen.
			Eventually(func() bool {
				err := k8sClient.Get(ctx, secretLookupKey, createdSecret)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			By("By creating a new DiscoveryConfig")
			discoveryConfig := &discoveryv1.DiscoveryConfig{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "discovery.open-cluster-management.io/v1",
					Kind:       "DiscoveryConfig",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      DiscoveryConfigName,
					Namespace: DiscoveryNamespace,
				},
				Spec: discoveryv1.DiscoveryConfigSpec{
					ProviderConnections: []string{SecretName},
				},
			}
			discoveryConfig = annotated(discoveryConfig)

			Expect(k8sClient.Create(ctx, discoveryConfig)).Should(Succeed())

			By("By checking that at least one discovered cluster has been created")
			discoveredClusters := &discoveryv1.DiscoveredClusterList{}
			Eventually(func() (int, error) {
				err := k8sClient.List(ctx, discoveredClusters, client.InNamespace(DiscoveryNamespace))
				if err != nil {
					return 0, err
				}
				return len(discoveredClusters.Items), nil
			}, time.Second*15, interval).Should(BeNumerically(">", 0))
		})
	})

	Context("When creating a ManagedCluster", func() {
		It("Should update the matching discovered cluster as managed", func() {
			ctx := context.Background()

			By("By creating a new ManagedCluster")
			mc1 := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "cluster.open-cluster-management.io/v1",
					"kind":       "ManagedCluster",
					"metadata": map[string]interface{}{
						"name":      "test-managedcluster",
						"namespace": DiscoveryNamespace,
						"labels": map[string]string{
							"clusterID": "69aced7c-286d-471c-9482-eac8a1cd2e17",
						},
					},
					"spec": map[string]interface{}{
						"hubAcceptsClient":     true,
						"leaseDurationSeconds": 60,
					},
				},
			}
			Expect(k8sClient.Create(ctx, mc1)).To(Succeed())

			By("Checking that a DiscoveredCluster is now labeled as managed")
			var fetchedDiscoveredClusters discoveryv1.DiscoveredClusterList
			Eventually(func() int {
				err := k8sClient.List(ctx, &fetchedDiscoveredClusters,
					client.InNamespace(DiscoveryNamespace),
					client.MatchingLabels{
						"isManagedCluster": "true",
					})
				if err != nil {
					return 0
				}
				return len(fetchedDiscoveredClusters.Items)
			}, timeout, interval).Should(Not(BeZero()))

		})
	})

	Context("When deleting the ManagedCluster", func() {
		It("Should update the matching discovered cluster to be unmanaged", func() {
			ctx := context.Background()

			By("By deleting a new ManagedCluster")
			mc1 := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "cluster.open-cluster-management.io/v1",
					"kind":       "ManagedCluster",
					"metadata": map[string]interface{}{
						"name":      "test-managedcluster",
						"namespace": DiscoveryNamespace,
						"labels": map[string]string{
							"clusterID": "69aced7c-286d-471c-9482-eac8a1cd2e17",
						},
					},
					"spec": map[string]interface{}{
						"hubAcceptsClient":     true,
						"leaseDurationSeconds": 60,
					},
				},
			}
			Expect(k8sClient.Delete(ctx, mc1)).To(Succeed())

			By("Checking that no discovered clusters are labeled as managed")
			var fetchedDiscoveredClusters discoveryv1.DiscoveredClusterList
			Eventually(func() int {
				err := k8sClient.List(ctx, &fetchedDiscoveredClusters,
					client.InNamespace(DiscoveryNamespace),
					client.MatchingLabels{
						"isManagedCluster": "true",
					})
				if err != nil {
					return 1
				}
				return len(fetchedDiscoveredClusters.Items)
			}, timeout, interval).Should(BeZero())

		})
	})

	Context("When modifying a DiscoveryConfig", func() {
		It("Should restrict the number of discovered clusters based on its filter", func() {
			By("By counting existing discovered clusters")
			ctx := context.Background()

			clusterCount := 0
			discoveredClusters := &discoveryv1.DiscoveredClusterList{}
			// Wait for discovered clusters created to no longer be in flux
			Eventually(func() bool {
				err := k8sClient.List(ctx, discoveredClusters, client.InNamespace(DiscoveryNamespace))
				if err != nil {
					return false
				}
				if count := len(discoveredClusters.Items); count != clusterCount {
					clusterCount = count
					return false
				}
				return true
			}, time.Second*15, time.Second).Should(BeTrue())

			By("By adding a filter to the DiscoveryConfig")
			configLookupKey := types.NamespacedName{Name: DiscoveryConfigName, Namespace: DiscoveryNamespace}
			createdConfig := &discoveryv1.DiscoveryConfig{}

			err := k8sClient.Get(ctx, configLookupKey, createdConfig)
			Expect(err).ToNot(HaveOccurred())

			createdConfig.Spec.Filters.LastActive = 30

			Expect(k8sClient.Update(ctx, createdConfig)).Should(Succeed())

			By("By checking that at least one discovered cluster has been filtered and deleted")
			discoveredClusters = &discoveryv1.DiscoveredClusterList{}
			Eventually(func() (int, error) {
				err := k8sClient.List(ctx, discoveredClusters, client.InNamespace(DiscoveryNamespace))
				if err != nil {
					return 0, err
				}
				return len(discoveredClusters.Items), nil
			}, time.Second*15, interval).Should(BeNumerically("<", clusterCount))

		})
	})

	Context("When deleting a DiscoveryConfig", func() {
		It("Should clean up all discovered clusters via garbage collection", func() {
			By("By deleting the DiscoveryConfig")
			ctx := context.Background()

			configLookupKey := types.NamespacedName{Name: DiscoveryConfigName, Namespace: DiscoveryNamespace}
			createdConfig := &discoveryv1.DiscoveryConfig{}

			err := k8sClient.Get(ctx, configLookupKey, createdConfig)
			Expect(err).ToNot(HaveOccurred())

			Expect(k8sClient.Delete(ctx, createdConfig)).Should(Succeed())

			By("By counting existing discovered clusters")
			discoveredClusters := &discoveryv1.DiscoveredClusterList{}
			// Wait for discovered clusters created to no longer be in flux
			Eventually(func() int {
				err := k8sClient.List(ctx, discoveredClusters, client.InNamespace(DiscoveryNamespace))
				if err != nil {
					return 1
				}
				return len(discoveredClusters.Items)
			}, time.Second*15, time.Second).Should(BeNumerically("==", 0))
		})
	})

})

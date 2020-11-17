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
					"token": "dummytoken",
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

			Expect(k8sClient.Create(ctx, discoveryConfig)).Should(Succeed())

			configLookupKey := types.NamespacedName{Name: DiscoveryConfigName, Namespace: DiscoveryNamespace}
			createdConfig := &discoveryv1.DiscoveryConfig{}

			// We'll need to retry getting this newly created DiscoveryConfig, given that creation may not immediately happen.
			Eventually(func() bool {
				err := k8sClient.Get(ctx, configLookupKey, createdConfig)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			By("By creating a new DiscoveryRefresh")
			refresh := &discoveryv1.DiscoveredClusterRefresh{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "discovery.open-cluster-management.io/v1",
					Kind:       "DiscoveredClusterRefresh",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "refresh",
					Namespace: DiscoveryNamespace,
				},
				Spec: discoveryv1.DiscoveredClusterRefreshSpec{},
			}

			Expect(k8sClient).ToNot(BeNil())
			Expect(k8sClient.Create(ctx, refresh)).Should(Succeed())

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

			createdConfig.Spec.Filters.Age = 30

			Expect(k8sClient.Update(ctx, createdConfig)).Should(Succeed())

			// We'll need to retry getting this newly created DiscoveryConfig, given that creation may not immediately happen.
			Eventually(func() bool {
				createdConfig = &discoveryv1.DiscoveryConfig{}
				err := k8sClient.Get(ctx, configLookupKey, createdConfig)
				if err != nil {
					return false
				}
				if createdConfig.Spec.Filters.Age != 30 {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			By("By creating a new DiscoveryRefresh")
			refresh := &discoveryv1.DiscoveredClusterRefresh{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "discovery.open-cluster-management.io/v1",
					Kind:       "DiscoveredClusterRefresh",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "refresh2",
					Namespace: DiscoveryNamespace,
				},
				Spec: discoveryv1.DiscoveredClusterRefreshSpec{},
			}

			Expect(k8sClient).ToNot(BeNil())
			Expect(k8sClient.Create(ctx, refresh)).Should(Succeed())

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

})

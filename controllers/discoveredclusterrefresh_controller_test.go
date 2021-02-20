// Copyright Contributors to the Open Cluster Management project

package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	discoveryv1 "github.com/open-cluster-management/discovery/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("DiscoveredClusterRefresh controller", func() {
	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		DiscoveryConfigName = "discoveryconfig"
		Namespace           = "refresh"
		SecretName          = "test-connection-secret"

		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When creating a DiscoveredClusterRefresh", func() {
		It("Should signal for a reconcile on the discovery config", func() {
			ctx := context.Background()

			By("Creating a namespace to work in")
			ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: Namespace}}
			Expect(k8sClient.Create(ctx, ns)).Should(Succeed())

			nsKey := types.NamespacedName{Name: Namespace}
			nsCreated := &corev1.Namespace{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, nsKey, nsCreated)
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
					Namespace: Namespace,
				},
				Spec: discoveryv1.DiscoveryConfigSpec{},
			}

			Expect(k8sClient.Create(ctx, discoveryConfig)).Should(Succeed())

			configLookupKey := types.NamespacedName{Name: DiscoveryConfigName, Namespace: Namespace}
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
					Namespace: Namespace,
				},
				Spec: discoveryv1.DiscoveredClusterRefreshSpec{},
			}

			Expect(k8sClient.Create(ctx, refresh)).Should(Succeed())

			refreshKey := types.NamespacedName{Name: "refresh", Namespace: Namespace}
			createdRefresh := &discoveryv1.DiscoveredClusterRefresh{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, refreshKey, createdRefresh)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			By("By checking that the DiscoveryRefresh is deleted once recognized")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, refreshKey, createdRefresh)
				if err != nil {
					return errors.IsNotFound(err)
				}
				return false
			}, timeout, interval).Should(BeTrue())

		})

		It("Should not trigger a reconcile without a discovery config", func() {
			ctx := context.Background()

			By("By deleting the DiscoveryConfig")
			configLookupKey := types.NamespacedName{Name: DiscoveryConfigName, Namespace: Namespace}
			createdConfig := &discoveryv1.DiscoveryConfig{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, configLookupKey, createdConfig)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			Expect(k8sClient.Delete(ctx, createdConfig)).Should(Succeed())

			Eventually(func() bool {
				err := k8sClient.Get(ctx, configLookupKey, createdConfig)
				if err != nil {
					return errors.IsNotFound(err)
				}
				return false
			}, timeout, interval).Should(BeTrue())

			By("By creating a new DiscoveryRefresh")
			refresh := &discoveryv1.DiscoveredClusterRefresh{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "discovery.open-cluster-management.io/v1",
					Kind:       "DiscoveredClusterRefresh",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "refresh",
					Namespace: Namespace,
				},
				Spec: discoveryv1.DiscoveredClusterRefreshSpec{},
			}

			Expect(k8sClient.Create(ctx, refresh)).Should(Succeed())

			refreshKey := types.NamespacedName{Name: "refresh", Namespace: Namespace}
			createdRefresh := &discoveryv1.DiscoveredClusterRefresh{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, refreshKey, createdRefresh)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			By("Checking that the DiscoveryRefresh is not deleted")
			Consistently(func() bool {
				err := k8sClient.Get(ctx, refreshKey, createdRefresh)
				if err != nil {
					return false
				}
				return true
			}, time.Second*3, interval).Should(BeTrue())

		})
	})
})

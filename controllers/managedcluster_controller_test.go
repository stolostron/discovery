// Copyright Contributors to the Open Cluster Management project

package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	discoveryv1 "github.com/open-cluster-management/discovery/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	ref "k8s.io/client-go/tools/reference"
)

var _ = Describe("ManagedCluster controller", func() {
	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		Namespace  = "managed"
		SecretName = "test-connection-secret"

		timeout  = time.Second * 4
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When creating a ManagedCluster", func() {
		It("Should update the corresponding discovered cluster to indicate managed status", func() {
			ctx := context.Background()

			By("Creating a namespace to work in")
			ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: Namespace}}
			Expect(k8sClient.Create(ctx, ns)).To(Succeed())

			secretKey := types.NamespacedName{
				Namespace: Namespace,
				Name:      SecretName,
			}
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      secretKey.Name,
					Namespace: secretKey.Namespace,
				},
				StringData: map[string]string{
					"metadata": "ocmAPIToken: dummytoken",
				},
			}
			Expect(k8sClient.Create(ctx, secret)).To(Succeed())

			createdSecret := corev1.Secret{}
			Eventually(func() error {
				return k8sClient.Get(ctx, secretKey, &createdSecret)
			}, timeout, interval).Should(Succeed())

			By("By manually creating an unmanaged DiscoveredCluster")
			key := types.NamespacedName{
				Namespace: Namespace,
				Name:      "discoveredcluster1",
			}
			dc1 := &discoveryv1.DiscoveredCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: discoveryv1.DiscoveredClusterSpec{
					Name: "cluster-id-1",
				},
			}

			secretRef, err := ref.GetReference(k8sClient.Scheme(), &createdSecret)
			Expect(err).NotTo(HaveOccurred())
			dc1.Spec.ProviderConnections = append(dc1.Spec.ProviderConnections, *secretRef)

			Expect(k8sClient.Create(ctx, dc1)).To(Succeed())

			By("By creating a new ManagedCluster")
			mc1 := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "cluster.open-cluster-management.io/v1",
					"kind":       "ManagedCluster",
					"metadata": map[string]interface{}{
						"name":      "managedcluster1",
						"namespace": Namespace,
						"labels": map[string]string{
							"clusterID": "cluster-id-1",
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, mc1)).To(Succeed())

			By("Checking the DiscoveredCluster is now labeled as managed")
			var fetchedDiscoveredCluster discoveryv1.DiscoveredCluster
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, key, &fetchedDiscoveredCluster); err != nil {
					return false
				}
				return fetchedDiscoveredCluster.Spec.IsManagedCluster
			}, timeout, interval).Should(BeTrue())

			By("By deleting the ManagedCluster")
			Expect(k8sClient.Delete(ctx, mc1)).To(Succeed())

			By("Checking the DiscoveredCluster is now labeled as unmanaged")
			Expect(k8sClient.Get(ctx, key, &fetchedDiscoveredCluster)).To(Succeed())
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, key, &fetchedDiscoveredCluster); err != nil {
					return true
				}
				return fetchedDiscoveredCluster.Spec.IsManagedCluster
			}, timeout, interval).Should(BeFalse())

		})
	})
})

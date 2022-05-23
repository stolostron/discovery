// Copyright Contributors to the Open Cluster Management project

package e2e

import (
	"context"
	"fmt"
	"time"

	discovery "github.com/stolostron/discovery/api/v1"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// Define utility constants for object names and testing timeouts/durations and intervals.
const (
	DiscoveryConfigName = "discovery"
	SecretName          = "test-connection-secret"
	TestserverName      = "mock-ocm-server"

	timeout  = time.Second * 30
	interval = time.Millisecond * 250
)

var (
	ctx                = context.Background()
	globalsInitialized = false
	discoveryNamespace = ""
	baseURL            = ""
	secondNamespace    = "secondary-test-ns"
	k8sClient          client.Client

	discoveryConfig = types.NamespacedName{}
	testserver      = types.NamespacedName{}
	ocmSecret       = types.NamespacedName{}
)

func initializeGlobals() {
	discoveryNamespace = *DiscoveryNamespace
	baseURL = *BaseURL
	discoveryConfig = types.NamespacedName{
		Name:      DiscoveryConfigName,
		Namespace: discoveryNamespace,
	}
	testserver = types.NamespacedName{
		Name:      TestserverName,
		Namespace: discoveryNamespace,
	}
	ocmSecret = types.NamespacedName{
		Name:      SecretName,
		Namespace: discoveryNamespace,
	}
}

var _ = Describe("[P1][Sev1][installer] Discoveryconfig controller", func() {

	BeforeEach(func() {
		if !globalsInitialized {
			initializeGlobals()
			globalsInitialized = true
		}
	})

	AfterEach(func() {
		err := k8sClient.Delete(ctx, defaultDiscoveryConfig(), client.PropagationPolicy(metav1.DeletePropagationForeground))
		if err != nil {
			Expect(apierrors.IsNotFound(err)).To(BeTrue())
		}

		err = k8sClient.Delete(ctx, dummySecret())
		if err != nil {
			Expect(apierrors.IsNotFound(err)).To(BeTrue())
		}

		// Wait for secret to be gone
		Eventually(func() bool {
			err := k8sClient.Get(ctx, ocmSecret, &corev1.Secret{})
			if err == nil {
				return false
			}
			return apierrors.IsNotFound(err)
		}, timeout, interval).Should(BeTrue(), "There was an issue cleaning up the secret.")

		// Once this succeeds, clean up has happened for all owned resources.
		Eventually(func() bool {
			err := k8sClient.Get(ctx, discoveryConfig, &discovery.DiscoveryConfig{})
			if err == nil {
				return false
			}
			return apierrors.IsNotFound(err)
		}, timeout, interval).Should(BeTrue(), "There was an issue cleaning up the DiscoveryConfig.")
	})

	Context("Creating a DiscoveryConfig", func() {
		It("Should create discovered clusters ", func() {
			By("By creating a secret with OCM credentials", func() {
				Expect(k8sClient.Create(ctx, dummySecret())).Should(Succeed())
			})

			By("By creating a new DiscoveryConfig", func() {
				Expect(k8sClient.Create(ctx, annotateWithScenario(defaultDiscoveryConfig(), "tenClusters"))).Should(Succeed())
			})

			By("By checking 10 discovered clusters have been created", func() {
				Eventually(func() (int, error) {
					return countDiscoveredClusters(discoveryNamespace)
				}, timeout, interval).Should(Equal(10))
			})
		})
	})

	// Context("Creating 999 Clusters", func() {
	// 	It("Should create discovered clusters ", func() {
	// 		By("Setting the testserver's response", func() {
	// 			updateTestserverScenario("nineninenineClusters")
	// 		})
	// 		By("By creating a secret with OCM credentials", func() {
	// 			Expect(k8sClient.Create(ctx, dummySecret())).Should(Succeed())
	// 		})

	// 		By("By creating a new DiscoveryConfig", func() {
	// 			Expect(k8sClient.Create(ctx, annotate(defaultDiscoveryConfig()))).Should(Succeed())
	// 		})
	// 		By("By checking 999 discovered clusters have been created", func() {
	// 			Eventually(func() (int, error) {
	// 				return countDiscoveredClusters(discoveryNamespace)
	// 			}, timeout, interval).Should(Equal(999))
	// 		})
	// 	})
	// })

	Context("Tracking ManagedClusters", func() {
		AfterEach(func() {
			err := k8sClient.Delete(ctx, customSecret("badsecret", discoveryNamespace, ""))
			if err != nil {
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			}
		})

		It("Should mark matching discovered clusters as being managed", func() {
			By("Creating unmanaged discovered clusters", func() {
				Expect(k8sClient.Create(ctx, dummySecret())).Should(Succeed())
				Expect(k8sClient.Create(ctx, annotateWithScenario(defaultDiscoveryConfig(), "tenClusters"))).Should(Succeed())
			})

			By("Checking that no DiscoveredClusters are labeled as managed", func() {
				Eventually(func() (int, error) {
					return countDiscoveredClusters(discoveryNamespace)
				}, timeout, interval).Should(Equal(10))
				Expect(countManagedDiscoveredClusters(discoveryNamespace)).To(Equal(0))
			})

			By("Creating ManagedClusters", func() {
				Expect(k8sClient.Create(ctx, newManagedCluster("testmc1", "844b3bf1-8d70-469c-a113-f1cd5db45c63"))).To(Succeed())
				// Managed but not discovered
				Expect(k8sClient.Create(ctx, newManagedCluster("testmc0", "abcdefgh-ijkl-mnop-qrst-uvwxyz123456"))).To(Succeed())
			})

			By("Forcing the DiscoveryConfig to be reconciled on", func() {
				// This is to test that the reconcile doesn't encounter errors
				// when a non-discovered managedcluster is present
				byTriggeringReconcile()
			})

			By("Checking that a DiscoveredCluster is now labeled as managed", func() {
				Eventually(func() (int, error) {
					return countManagedDiscoveredClusters(discoveryNamespace)
				}, timeout, interval).Should(Equal(1))
			})

			By("Verify controller still cleans up discovered clusters", func() {
				By("Changing secret to an invalid one", func() {
					Expect(k8sClient.Create(ctx, customSecret("badsecret", discoveryNamespace, ""))).Should(Succeed())

					config := &discovery.DiscoveryConfig{}
					Expect(k8sClient.Get(ctx, discoveryConfig, config)).To(Succeed())
					config.Spec.Credential = "badsecret"
					Expect(k8sClient.Update(ctx, config)).Should(Succeed())
				})

				By("Checking all discovered clusters are gone", func() {
					Eventually(func() (int, error) {
						return countDiscoveredClusters(discoveryNamespace)
					}, timeout, interval).Should(Equal(0))
				})
			})

			By("Deleting ManagedClusters", func() {
				byDeletingManagedClusters([]string{"testmc0", "testmc1"})
			})
		})

		It("Should unmark discovered clusters when they are no longer managed", func() {
			if inCanary {
				Skip("Skipping test in a canary environment")
			}

			By("Creating ManagedClusters", func() {
				Expect(k8sClient.Create(ctx, newManagedCluster("testmc1", "844b3bf1-8d70-469c-a113-f1cd5db45c63"))).To(Succeed())
				Expect(k8sClient.Create(ctx, newManagedCluster("testmc2", "dbcbbeeb-7a15-4c64-9975-6f6c331255c8"))).To(Succeed())
				Expect(k8sClient.Create(ctx, newManagedCluster("testmc3", "6429154f-583e-4d95-b74d-2cd02b266ecf"))).To(Succeed())
				Expect(k8sClient.Create(ctx, newManagedCluster("testmc4", "d36f6dc3-84b0-4bc6-b126-9f30766f9fae"))).To(Succeed())
				Expect(k8sClient.Create(ctx, newManagedCluster("testmc5", "f1083487-e6ae-4388-9408-af09fcc9c7fc"))).To(Succeed())
			})

			By("Creating discovered clusters", func() {
				Expect(k8sClient.Create(ctx, dummySecret())).Should(Succeed())
				Expect(k8sClient.Create(ctx, annotateWithScenario(defaultDiscoveryConfig(), "tenClusters"))).Should(Succeed())
			})

			By("Checking that all ManagedClusters are recognized in their matching DiscoveredClusters", func() {
				Eventually(func() (int, error) {
					return countManagedDiscoveredClusters(discoveryNamespace)
				}, timeout, interval).Should(Equal(5))
			})

			By("Deleting ManagedClusters", func() {
				byDeletingManagedClusters([]string{"testmc1", "testmc2", "testmc3", "testmc4", "testmc5"})
			})

			By("Checking that no DiscoveredClusters are labeled as managed", func() {
				Eventually(func() (int, error) {
					return countManagedDiscoveredClusters(discoveryNamespace)
				}, timeout, interval).Should(Equal(0))
				Expect(countDiscoveredClusters(discoveryNamespace)).To(Equal(10))
			})
		})
	})

	Context("Unchanged DiscoveryConfig", func() {
		It("Should update discovered clusters when the OCM responses change", func() {
			By("Creating a DiscoveryConfig and discovered clusters", func() {
				Expect(k8sClient.Create(ctx, dummySecret())).Should(Succeed())
				Expect(k8sClient.Create(ctx, annotateWithScenario(defaultDiscoveryConfig(), "tenClusters"))).Should(Succeed())
				Eventually(func() (int, error) {
					return countDiscoveredClusters(discoveryNamespace)
				}, timeout, interval).Should(Equal(10))
			})

			By("Changing the number of clusters returned from the testserver", func() {
				updateTestserverScenario("fiveClusters")
			})

			By("By checking that discovered clusters have changed", func() {
				Eventually(func() (int, error) {
					return countDiscoveredClusters(discoveryNamespace)
				}, timeout, interval).Should(Equal(5))
			})
		})
	})

	Context("Deleting a DiscoveryConfig", func() {
		It("Should delete all discovered clusters via garbage collection", func() {
			By("Creating a DiscoveryConfig and discovered clusters", func() {
				Expect(k8sClient.Create(ctx, dummySecret())).Should(Succeed())
				Expect(k8sClient.Create(ctx, annotateWithScenario(defaultDiscoveryConfig(), "tenClusters"))).Should(Succeed())
				Eventually(func() (int, error) {
					return countDiscoveredClusters(discoveryNamespace)
				}, timeout, interval).Should(Equal(10))
			})

			By("Deleting the DiscoveryConfig", func() {
				Expect(k8sClient.Delete(ctx, defaultDiscoveryConfig(), client.PropagationPolicy(metav1.DeletePropagationForeground))).Should(Succeed())
			})

			By("Checking all discovered clusters are gone", func() {
				Eventually(func() (int, error) {
					return countDiscoveredClusters(discoveryNamespace)
				}, timeout, interval).Should(Equal(0))
			})
		})
	})

	Context("Archived clusters", func() {
		It("Should not create DiscoveredClusters out of archived clusters", func() {
			By("Creating a DiscoveryConfig", func() {
				Expect(k8sClient.Create(ctx, dummySecret())).Should(Succeed())
				Expect(k8sClient.Create(ctx, annotateWithScenario(defaultDiscoveryConfig(), "archivedClusters"))).Should(Succeed())
			})

			By("By checking only 8 discovered clusters have been created", func() {
				Eventually(func() (int, error) {
					return countDiscoveredClusters(discoveryNamespace)
				}, timeout, interval).Should(Equal(8))
			})
		})
	})

	Context("Credentials become invalid", func() {
		AfterEach(func() {
			err := k8sClient.Delete(ctx, customSecret("badsecret", discoveryNamespace, ""))
			if err != nil {
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			}
		})

		It("Should delete discovered clusters when secret changes to an invalid one", func() {
			By("By creating a secret with OCM credentials", func() {
				Expect(k8sClient.Create(ctx, dummySecret())).Should(Succeed())
			})

			By("By creating a new DiscoveryConfig", func() {
				Expect(k8sClient.Create(ctx, annotateWithScenario(defaultDiscoveryConfig(), "tenClusters"))).Should(Succeed())
			})

			By("By checking 10 discovered clusters have been created", func() {
				Eventually(func() (int, error) {
					return countDiscoveredClusters(discoveryNamespace)
				}, timeout, interval).Should(Equal(10))
			})

			By("Change secret to an invalid one", func() {
				Expect(k8sClient.Create(ctx, customSecret("badsecret", discoveryNamespace, ""))).Should(Succeed())

				config := &discovery.DiscoveryConfig{}
				Expect(k8sClient.Get(ctx, discoveryConfig, config)).To(Succeed())
				config.Spec.Credential = "badsecret"
				Expect(k8sClient.Update(ctx, config)).Should(Succeed())
			})

			By("Checking all discovered clusters are gone", func() {
				Eventually(func() (int, error) {
					return countDiscoveredClusters(discoveryNamespace)
				}, timeout, interval).Should(Equal(0))
			})
		})

		It("Should delete discovered clusters after secret is deleted", func() {
			By("By creating a secret with OCM credentials", func() {
				Expect(k8sClient.Create(ctx, dummySecret())).Should(Succeed())
			})

			By("By creating a new DiscoveryConfig", func() {
				Expect(k8sClient.Create(ctx, annotateWithScenario(defaultDiscoveryConfig(), "tenClusters"))).Should(Succeed())
			})

			By("By checking 10 discovered clusters have been created", func() {
				Eventually(func() (int, error) {
					return countDiscoveredClusters(discoveryNamespace)
				}, timeout, interval).Should(Equal(10))
			})

			By("Deleting the secret", func() {
				err := k8sClient.Delete(ctx, dummySecret())
				if err != nil {
					Expect(apierrors.IsNotFound(err)).To(BeTrue())
				}
			})

			By("Forcing the DiscoveryConfig to be reconciled on", func() {
				byTriggeringReconcile()
			})

			By("Checking all discovered clusters are gone", func() {
				Eventually(func() (int, error) {
					return countDiscoveredClusters(discoveryNamespace)
				}, timeout, interval).Should(Equal(0))
			})
		})
	})

	Context("Multiple DiscoveryConfigs across namespaces", func() {
		BeforeEach(func() {
			Expect(k8sClient.Create(ctx, &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{Name: secondNamespace},
			})).To(Succeed())

			// Wait for namespace to be established
			createdNS := &corev1.Namespace{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{Name: secondNamespace}, createdNS)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
		})

		AfterEach(func() {
			err := k8sClient.Delete(ctx, customSecret("badsecret", discoveryNamespace, ""))
			if err != nil {
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			}

			err = k8sClient.Delete(ctx, customSecret("connection1", discoveryNamespace, "connection1"))
			if err != nil {
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			}

			ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: secondNamespace}}
			err = k8sClient.Delete(ctx, ns)
			if err != nil {
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{Name: secondNamespace}, &corev1.Namespace{})
				if err == nil {
					return false
				}
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue(), "There was an issue cleaning up the namespace.")
		})

		It("Should manage DiscoveredClusters across namespaces", func() {
			By("Creating DiscoveryConfigs in separate namespaces", func() {
				Expect(k8sClient.Create(ctx, customSecret("connection1", discoveryNamespace, "connection1"))).Should(Succeed())
				Expect(k8sClient.Create(ctx, customSecret("connection2", secondNamespace, "connection2"))).Should(Succeed())

				config1 := defaultDiscoveryConfig()
				config1.Spec.Credential = "connection1"
				Expect(k8sClient.Create(ctx, annotateWithScenario(config1, "multipleConnections"))).Should(Succeed())

				config2 := defaultDiscoveryConfig()
				config2.SetNamespace(secondNamespace)
				config2.Spec.Credential = "connection2"
				Expect(k8sClient.Create(ctx, annotateWithScenario(config2, "multipleConnections"))).Should(Succeed())
			})

			By("By checking discovered clusters have been created in both namespaces", func() {
				Eventually(func() (int, error) {
					return countDiscoveredClusters(discoveryNamespace)
				}, timeout, interval).Should(Equal(4), "discoveredClusters not created in first namespace")
				Eventually(func() (int, error) {
					return countDiscoveredClusters(secondNamespace)
				}, timeout, interval).Should(Equal(4), "discoveredClusters not created in second namespace")
			})

			By("By verifying a managedcluster change propogates across all namespaces", func() {
				By("Change secret to an invalid one", func() {
					Expect(k8sClient.Create(ctx, customSecret("badsecret", discoveryNamespace, ""))).Should(Succeed())

					config := &discovery.DiscoveryConfig{}
					Expect(k8sClient.Get(ctx, discoveryConfig, config)).To(Succeed())
					config.Spec.Credential = "badsecret"
					Expect(k8sClient.Update(ctx, config)).Should(Succeed())
				})

				By("Checking all discovered clusters are gone", func() {
					Eventually(func() (int, error) {
						return countDiscoveredClusters(discoveryNamespace)
					}, timeout, interval).Should(Equal(0))
				})

				By("Checking discovered clusters in other namespace are still there", func() {
					Expect(countDiscoveredClusters(secondNamespace)).To(Equal(4))
				})
			})
		})

		It("Should update DiscoveredClusters' managed status across namespaces", func() {
			if inCanary {
				Skip("Skipping test in a canary environment")
			}

			By("Creating DiscoveryConfigs in separate namespaces", func() {
				Expect(k8sClient.Create(ctx, customSecret("connection1", discoveryNamespace, "connection1"))).Should(Succeed())
				Expect(k8sClient.Create(ctx, customSecret("connection2", secondNamespace, "connection2"))).Should(Succeed())

				config1 := defaultDiscoveryConfig()
				config1.Spec.Credential = "connection1"
				Expect(k8sClient.Create(ctx, annotateWithScenario(config1, "multipleConnections"))).Should(Succeed())

				config2 := defaultDiscoveryConfig()
				config2.SetNamespace(secondNamespace)
				config2.Spec.Credential = "connection2"
				Expect(k8sClient.Create(ctx, annotateWithScenario(config2, "multipleConnections"))).Should(Succeed())
			})

			By("By checking discovered clusters have been created in both namespaces", func() {
				Eventually(func() (int, error) {
					return countDiscoveredClusters(discoveryNamespace)
				}, timeout, interval).Should(Equal(4), "discoveredClusters not created in first namespace")
				Eventually(func() (int, error) {
					return countDiscoveredClusters(secondNamespace)
				}, timeout, interval).Should(Equal(4), "discoveredClusters not created in second namespace")
			})

			By("By verifying a change to a managedcluster applies to matching discoveredClusters in all namespaces", func() {
				By("Creating ManagedClusters", func() {
					Expect(k8sClient.Create(ctx, newManagedCluster("mc-connection-1", "d36f6dc3-84b0-4bc6-b126-9f30766f9fae"))).To(Succeed())
					Expect(k8sClient.Create(ctx, newManagedCluster("mc-connection-2", "b6ec171b-d733-40ed-ba9c-78e58a9c9a8b"))).To(Succeed())
					Expect(k8sClient.Create(ctx, newManagedCluster("mc-connection-both", "844b3bf1-8d70-469c-a113-f1cd5db45c63"))).To(Succeed())
				})

				By("Checking that all ManagedClusters are recognized in their matching DiscoveredClusters", func() {
					Eventually(func() (int, error) {
						return countManagedDiscoveredClusters(discoveryNamespace)
					}, timeout, interval).Should(Equal(2), fmt.Sprintf("Missing managed labels in namespace %s", discoveryNamespace))
					Eventually(func() (int, error) {
						return countManagedDiscoveredClusters(secondNamespace)
					}, timeout, interval).Should(Equal(2), fmt.Sprintf("Missing managed labels in namespace %s", secondNamespace))
				})

				By("Deleting ManagedClusters", func() {
					byDeletingManagedClusters([]string{"mc-connection-1", "mc-connection-2", "mc-connection-both"})
				})

				By("Checking that no DiscoveredClusters are labeled as managed", func() {
					Eventually(func() (int, error) {
						return countManagedDiscoveredClusters(discoveryNamespace)
					}, timeout, interval).Should(Equal(0))
					Eventually(func() (int, error) {
						return countManagedDiscoveredClusters(secondNamespace)
					}, timeout, interval).Should(Equal(0))
				})
			})
		})
	})

})

// annotate adds an annotation to modify the baseUrl used with the discoveryconfig
func annotate(dc *discovery.DiscoveryConfig) *discovery.DiscoveryConfig {
	if baseURL != "" {
		dc.SetAnnotations(map[string]string{"ocmBaseURL": baseURL, "authBaseURL": baseURL})
		return dc
	} else {
		dc.SetAnnotations(map[string]string{"ocmBaseURL": defaultBaseUrl(), "authBaseURL": defaultBaseUrl()})

		return dc
	}
}

// annotateWithScenario adds an annotation to modify the baseUrl with a scenario path
func annotateWithScenario(dc *discovery.DiscoveryConfig, scenario string) *discovery.DiscoveryConfig {
	if baseURL != "" {
		dc.SetAnnotations(map[string]string{"ocmBaseURL": baseURL + "/" + scenario, "authBaseURL": baseURL + "/" + scenario})
		return dc
	} else {
		dc.SetAnnotations(map[string]string{"ocmBaseURL": defaultBaseUrl() + "/" + scenario, "authBaseURL": defaultBaseUrl() + "/" + scenario})
		return dc
	}
}

func defaultBaseUrl() string {
	return fmt.Sprintf("http://mock-ocm-server.%s.svc.cluster.local:3000", discoveryNamespace)
}

// Updates the entrypoint argument of the testserver deployment. This changes the content the
// deployment serves.
func updateTestserverScenario(scenario string) {
	config := &discovery.DiscoveryConfig{}
	Expect(k8sClient.Get(ctx, discoveryConfig, config)).To(Succeed())
	annotateWithScenario(config, scenario)
	Expect(k8sClient.Update(ctx, config)).Should(Succeed())
}

func listDiscoveredClusters() (*discovery.DiscoveredClusterList, error) {
	discoveredClusters := &discovery.DiscoveredClusterList{}
	err := k8sClient.List(ctx, discoveredClusters, client.InNamespace(discoveryNamespace))
	return discoveredClusters, err
}

func getDiscoveredClusterByID(id string) (*discovery.DiscoveredCluster, error) {
	discoveredClusters, err := listDiscoveredClusters()
	if err != nil {
		return nil, err
	}
	for _, dc := range discoveredClusters.Items {
		dc := dc
		if dc.Spec.Name == id {
			return &dc, nil
		}
	}
	return nil, fmt.Errorf("Cluster not found")
}

func countManagedDiscoveredClusters(namespace string) (int, error) {
	discoveredClusters := &discovery.DiscoveredClusterList{}
	err := k8sClient.List(ctx, discoveredClusters,
		client.InNamespace(namespace),
		client.MatchingLabels{
			"isManagedCluster": "true",
		})
	if err != nil {
		return -1, err
	}
	return len(discoveredClusters.Items), err
}

func countDiscoveredClusters(namespace string) (int, error) {
	discoveredClusters := &discovery.DiscoveredClusterList{}
	err := k8sClient.List(ctx, discoveredClusters, client.InNamespace(namespace))
	if err != nil {
		return -1, err
	}
	return len(discoveredClusters.Items), nil
}

func byDeletingManagedClusters(names []string) {
	mc := func(n string) *unstructured.Unstructured {
		u := &unstructured.Unstructured{}
		u.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   "cluster.open-cluster-management.io",
			Kind:    "ManagedCluster",
			Version: "v1",
		})
		u.SetName(n)
		return u
	}

	for _, name := range names {
		err := k8sClient.Delete(ctx, mc(name))
		if err != nil {
			Expect(apierrors.IsNotFound(err)).To(BeTrue())
		}
	}

	Eventually(func() bool {
		for _, name := range names {
			err := k8sClient.Get(ctx, client.ObjectKey{Name: name}, mc(name))
			if !apierrors.IsNotFound(err) {
				return false
			}
		}
		return true
	}, timeout, interval).Should(BeTrue(), "There was an issue deleting managedclusters.")
}

func byTriggeringReconcile() {
	config := &discovery.DiscoveryConfig{}
	Expect(k8sClient.Get(ctx, discoveryConfig, config)).To(Succeed())
	a := config.GetAnnotations()
	if a == nil {
		a = map[string]string{}
	}
	a["triggerTime"] = time.Now().String()
	config.SetAnnotations(a)
	Expect(k8sClient.Update(ctx, config)).Should(Succeed())
}

func dummySecret() *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      SecretName,
			Namespace: discoveryNamespace,
		},
		StringData: map[string]string{
			"ocmAPIToken": "dummytoken",
		},
	}
}

func customSecret(name, namespace, token string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		StringData: map[string]string{
			"ocmAPIToken": token,
		},
	}
}

func defaultDiscoveryConfig() *discovery.DiscoveryConfig {
	return &discovery.DiscoveryConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DiscoveryConfigName,
			Namespace: discoveryNamespace,
		},
		Spec: discovery.DiscoveryConfigSpec{
			Credential: SecretName,
			Filters:    discovery.Filter{LastActive: 7},
		},
	}
}

func newManagedCluster(name, clusterID string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "cluster.open-cluster-management.io/v1",
			"kind":       "ManagedCluster",
			"metadata": map[string]interface{}{
				"name": name,
				"labels": map[string]string{
					"clusterID": clusterID,
				},
			},
			"spec": map[string]interface{}{
				"hubAcceptsClient":     true,
				"leaseDurationSeconds": 60,
			},
		},
	}
}

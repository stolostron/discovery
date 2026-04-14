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
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	discovery "github.com/stolostron/discovery/api/v1"
	"github.com/stolostron/discovery/pkg/ocm/auth"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	TestDiscoveryConfigName = "discovery"
	TestDiscoveryNamespace  = "discovery"
	TestSecretName          = "test-connection-secret"

	timeout  = time.Second * 60
	interval = time.Millisecond * 250
)

var (
	ctx                 = context.Background()
	testDiscoveryConfig = types.NamespacedName{Name: TestDiscoveryConfigName, Namespace: TestDiscoveryNamespace}
	mockTime            = metav1.NewTime(time.Date(2020, 5, 29, 6, 0, 0, 0, time.UTC))
	mockCluster410      = discovery.DiscoveredCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "t2",
			Namespace: TestDiscoveryNamespace,
		},
		Spec: discovery.DiscoveredClusterSpec{
			Name:              "t2",
			DisplayName:       "t2",
			OpenshiftVersion:  "4.10.0",
			CreationTimestamp: &mockTime,
			ActivityTimestamp: &mockTime,
		},
	}
	mockCluster411 = discovery.DiscoveredCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "t1",
			Namespace: TestDiscoveryNamespace,
		},
		Spec: discovery.DiscoveredClusterSpec{
			Name:              "t1",
			DisplayName:       "t1",
			OpenshiftVersion:  "4.11.0",
			CreationTimestamp: &mockTime,
			ActivityTimestamp: &mockTime,
		},
	}
)

var _ = Describe("Discoveryconfig controller", func() {

	Context("Creating a DiscoveryConfig", func() {
		It("Should create discovered clusters ", func() {
			By("By creating a namespace", func() {
				err := k8sClient.Create(ctx, &corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{Name: TestDiscoveryNamespace},
				})
				Expect(err).NotTo(HaveOccurred())
			})

			By("By creating a secret with OCM credentials", func() {
				Expect(k8sClient.Create(ctx, &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      TestSecretName,
						Namespace: TestDiscoveryNamespace,
					},
					StringData: map[string]string{
						"ocmAPIToken": "dummytoken",
					},
				})).Should(Succeed())
			})

			By("By creating a new DiscoveryConfig", func() {
				Expect(k8sClient.Create(ctx, &discovery.DiscoveryConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name:      TestDiscoveryConfigName,
						Namespace: TestDiscoveryNamespace,
					},
					Spec: discovery.DiscoveryConfigSpec{
						Credential: TestSecretName,
						Filters:    discovery.Filter{LastActive: 7},
					},
				})).Should(Succeed())
			})

			mockDiscoveredCluster = func() ([]discovery.DiscoveredCluster, error) {
				return []discovery.DiscoveredCluster{
					mockCluster410,
					mockCluster411,
				}, nil
			}

			By("By checking 2 discovered clusters have been created", func() {
				Eventually(func() (int, error) {
					return countDiscoveredClusters(TestDiscoveryNamespace)
				}, timeout, interval).Should(Equal(2))
			})

			mockDiscoveredCluster = func() ([]discovery.DiscoveredCluster, error) {
				mockCluster411.Spec.DisplayName = "newname"
				return []discovery.DiscoveredCluster{
					mockCluster411,
				}, nil
			}

			By("By adding a filter to DiscoveryConfig", func() {
				config := &discovery.DiscoveryConfig{}
				Expect(k8sClient.Get(ctx, testDiscoveryConfig, config)).To(Succeed())
				config.Spec.Filters = discovery.Filter{OpenShiftVersions: []discovery.Semver{"4.11"}}
				Expect(k8sClient.Update(ctx, config)).Should(Succeed())
			})

			By("By checking 1 discovered cluster remains and is updated", func() {
				Eventually(func() (int, error) {
					return countDiscoveredClusters(TestDiscoveryNamespace)
				}, timeout, interval).Should(Equal(1))

			})

			mockDiscoveredCluster = func() ([]discovery.DiscoveredCluster, error) {
				return nil, auth.ErrInvalidToken
			}

			By("By removing a filter to DiscoveryConfig", func() {
				config := &discovery.DiscoveryConfig{}
				Expect(k8sClient.Get(ctx, testDiscoveryConfig, config)).To(Succeed())
				config.Spec.Filters = discovery.Filter{OpenShiftVersions: []discovery.Semver{"4.12"}}
				Expect(k8sClient.Update(ctx, config)).Should(Succeed())
			})

			By("By checking no discovered clusters remain", func() {
				Eventually(func() (int, error) {
					return countDiscoveredClusters(TestDiscoveryNamespace)
				}, timeout, interval).Should(Equal(0))

			})

		})
	})

	Context("Creating an invalid DiscoveryConfig", func() {
		It("Should not create discovered clusters ", func() {
			By("By creating a namespace", func() {
				err := k8sClient.Create(ctx, &corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{Name: "invalid"},
				})
				Expect(err).NotTo(HaveOccurred())
			})

			By("By creating a secret with OCM credentials", func() {
				Expect(k8sClient.Create(ctx, &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      TestSecretName,
						Namespace: "invalid",
					},
					StringData: map[string]string{
						"ocmAPIToken": "dummytoken",
					},
				})).Should(Succeed())
			})

			By("By creating a new DiscoveryConfig", func() {
				Expect(k8sClient.Create(ctx, &discovery.DiscoveryConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "invalid-name",
						Namespace: "invalid",
					},
					Spec: discovery.DiscoveryConfigSpec{
						Credential: TestSecretName,
						Filters:    discovery.Filter{LastActive: 7},
					},
				})).Should(Succeed())
			})

			mockDiscoveredCluster = func() ([]discovery.DiscoveredCluster, error) {
				return []discovery.DiscoveredCluster{
					mockCluster410,
					mockCluster411,
				}, nil
			}

			By("By checking no discovered clusters have been created", func() {
				Eventually(func() (int, error) {
					return countDiscoveredClusters(TestDiscoveryNamespace)
				}, timeout, interval).Should(Equal(0))
			})

		})
	})

	Context("Secret change detection during cluster creation", func() {
		const secretChangeNamespace = "secret-change-test"
		const secretChangeName = "secret-change-test"

		It("Should abort cluster creation when secret credentials change", func() {
			By("Creating a test namespace", func() {
				err := k8sClient.Create(ctx, &corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{Name: secretChangeNamespace},
				})
				Expect(err).NotTo(HaveOccurred())
			})

			By("Creating a secret with initial credentials", func() {
				Expect(k8sClient.Create(ctx, &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      secretChangeName,
						Namespace: secretChangeNamespace,
					},
					StringData: map[string]string{
						"ocmAPIToken": "initial-token",
					},
				})).Should(Succeed())
			})

			By("Creating a DiscoveryConfig", func() {
				Expect(k8sClient.Create(ctx, &discovery.DiscoveryConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name:      TestDiscoveryConfigName,
						Namespace: secretChangeNamespace,
					},
					Spec: discovery.DiscoveryConfigSpec{
						Credential: secretChangeName,
						Filters:    discovery.Filter{LastActive: 7},
					},
				})).Should(Succeed())
			})

			// Mock returning 150 clusters to trigger the check at cluster 100
			mockDiscoveredCluster = func() ([]discovery.DiscoveredCluster, error) {
				clusters := make([]discovery.DiscoveredCluster, 150)
				for i := 0; i < 150; i++ {
					clusterName := fmt.Sprintf("cluster-%d", i)
					clusters[i] = discovery.DiscoveredCluster{
						ObjectMeta: metav1.ObjectMeta{
							Name:      clusterName,
							Namespace: secretChangeNamespace,
						},
						Spec: discovery.DiscoveredClusterSpec{
							Name:              clusterName,
							DisplayName:       clusterName,
							OpenshiftVersion:  "4.10.0",
							CreationTimestamp: &mockTime,
							ActivityTimestamp: &mockTime,
						},
					}
				}
				return clusters, nil
			}

			// Set up hook to change secret after 100 clusters
			secretChanged := make(chan bool, 1)
			testClusterApplyHook = func(count int) {
				if count == 100 {
					secret := &corev1.Secret{}
					err := k8sClient.Get(ctx, types.NamespacedName{
						Name:      secretChangeName,
						Namespace: secretChangeNamespace,
					}, secret)
					if err == nil {
						secret.Data["ocmAPIToken"] = []byte("changed-token")
						_ = k8sClient.Update(ctx, secret)
						secretChanged <- true
					}
				}
			}
			defer func() { testClusterApplyHook = nil }()

			By("Waiting for secret change to be triggered", func() {
				// Wait for hook to trigger secret change
				Eventually(secretChanged, timeout).Should(Receive())
			})

			By("Verifying the secret was actually mutated", func() {
				secret := &corev1.Secret{}
				Expect(k8sClient.Get(ctx, types.NamespacedName{
					Name:      secretChangeName,
					Namespace: secretChangeNamespace,
				}, secret)).Should(Succeed())
				Expect(string(secret.Data["ocmAPIToken"])).Should(Equal("changed-token"))
			})

			By("Verifying that cluster creation was aborted at 100", func() {
				// Wait for reconciliation to complete and verify final count
				Eventually(func() (int, error) {
					return countDiscoveredClusters(secretChangeNamespace)
				}, timeout, interval).Should(Equal(100))

				// Verify count remains stable (hasn't continued creating clusters)
				Consistently(func() (int, error) {
					return countDiscoveredClusters(secretChangeNamespace)
				}, time.Second*2, interval).Should(Equal(100))
			})
		})

		It("Should abort cluster creation when secret is deleted", func() {
			const deletionNamespace = "secret-deletion-test"
			const deletionSecretName = "deletion-test-secret"

			By("Creating a test namespace", func() {
				err := k8sClient.Create(ctx, &corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{Name: deletionNamespace},
				})
				Expect(err).NotTo(HaveOccurred())
			})

			By("Creating a secret", func() {
				Expect(k8sClient.Create(ctx, &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      deletionSecretName,
						Namespace: deletionNamespace,
					},
					StringData: map[string]string{
						"ocmAPIToken": "test-token",
					},
				})).Should(Succeed())
			})

			By("Creating a DiscoveryConfig", func() {
				Expect(k8sClient.Create(ctx, &discovery.DiscoveryConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name:      TestDiscoveryConfigName,
						Namespace: deletionNamespace,
					},
					Spec: discovery.DiscoveryConfigSpec{
						Credential: deletionSecretName,
						Filters:    discovery.Filter{LastActive: 7},
					},
				})).Should(Succeed())
			})

			mockDiscoveredCluster = func() ([]discovery.DiscoveredCluster, error) {
				clusters := make([]discovery.DiscoveredCluster, 150)
				for i := 0; i < 150; i++ {
					clusterName := fmt.Sprintf("del-cluster-%d", i)
					clusters[i] = discovery.DiscoveredCluster{
						ObjectMeta: metav1.ObjectMeta{
							Name:      clusterName,
							Namespace: deletionNamespace,
						},
						Spec: discovery.DiscoveredClusterSpec{
							Name:              clusterName,
							DisplayName:       clusterName,
							OpenshiftVersion:  "4.10.0",
							CreationTimestamp: &mockTime,
							ActivityTimestamp: &mockTime,
						},
					}
				}
				return clusters, nil
			}

			// Set up hook to delete secret after 100 clusters
			secretDeleted := make(chan bool, 1)
			testClusterApplyHook = func(count int) {
				if count == 100 {
					GinkgoWriter.Printf("Test hook triggered at count=%d\n", count)
					secret := &corev1.Secret{}
					err := k8sClient.Get(ctx, types.NamespacedName{
						Name:      deletionSecretName,
						Namespace: deletionNamespace,
					}, secret)
					if err != nil {
						GinkgoWriter.Printf("Error getting secret: %v\n", err)
					} else {
						GinkgoWriter.Printf("Deleting secret\n")
						deleteErr := k8sClient.Delete(ctx, secret)
						if deleteErr != nil {
							GinkgoWriter.Printf("Error deleting secret: %v\n", deleteErr)
						}
						secretDeleted <- true
					}
				}
			}
			defer func() {
				GinkgoWriter.Printf("Cleaning up test hook\n")
				testClusterApplyHook = nil
			}()

			By("Waiting for secret deletion to be triggered", func() {
				// Wait for hook to trigger secret deletion
				Eventually(secretDeleted, timeout).Should(Receive())
			})

			By("Verifying the secret was actually deleted", func() {
				secret := &corev1.Secret{}
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      deletionSecretName,
					Namespace: deletionNamespace,
				}, secret)
				Expect(apierrors.IsNotFound(err)).Should(BeTrue())
			})

			By("Verifying that cluster creation was aborted at 100", func() {
				// Wait for reconciliation to complete and verify final count
				Eventually(func() (int, error) {
					return countDiscoveredClusters(deletionNamespace)
				}, timeout, interval).Should(Equal(100))

				// Verify count remains stable (hasn't continued creating clusters)
				Consistently(func() (int, error) {
					return countDiscoveredClusters(deletionNamespace)
				}, time.Second*2, interval).Should(Equal(100))
			})
		})

		It("Should create all clusters when secret does not change", func() {
			const noChangeNamespace = "no-change-test"
			const noChangeSecretName = "no-change-secret"

			By("Creating a test namespace", func() {
				err := k8sClient.Create(ctx, &corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{Name: noChangeNamespace},
				})
				Expect(err).NotTo(HaveOccurred())
			})

			By("Creating a secret", func() {
				Expect(k8sClient.Create(ctx, &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      noChangeSecretName,
						Namespace: noChangeNamespace,
					},
					StringData: map[string]string{
						"ocmAPIToken": "stable-token",
					},
				})).Should(Succeed())
			})

			By("Creating a DiscoveryConfig", func() {
				Expect(k8sClient.Create(ctx, &discovery.DiscoveryConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name:      TestDiscoveryConfigName,
						Namespace: noChangeNamespace,
					},
					Spec: discovery.DiscoveryConfigSpec{
						Credential: noChangeSecretName,
						Filters:    discovery.Filter{LastActive: 7},
					},
				})).Should(Succeed())
			})

			// Return 150 clusters to exercise the check at 100 (happy path: no change detected)
			mockDiscoveredCluster = func() ([]discovery.DiscoveredCluster, error) {
				clusters := make([]discovery.DiscoveredCluster, 150)
				for i := 0; i < 150; i++ {
					clusterName := fmt.Sprintf("stable-cluster-%d", i)
					clusters[i] = discovery.DiscoveredCluster{
						ObjectMeta: metav1.ObjectMeta{
							Name:      clusterName,
							Namespace: noChangeNamespace,
						},
						Spec: discovery.DiscoveredClusterSpec{
							Name:              clusterName,
							DisplayName:       clusterName,
							OpenshiftVersion:  "4.10.0",
							CreationTimestamp: &mockTime,
							ActivityTimestamp: &mockTime,
						},
					}
				}
				return clusters, nil
			}

			By("Verifying all 150 clusters are created (secret check passed at 100)", func() {
				Eventually(func() (int, error) {
					return countDiscoveredClusters(noChangeNamespace)
				}, timeout, interval).Should(Equal(150))
			})
		})
	})

})

func Test_parseSecretForAuth(t *testing.T) {
	tests := []struct {
		name    string
		secret  *corev1.Secret
		want    auth.AuthRequest
		wantErr bool
	}{
		{
			name: "Dummy token set",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				Data: map[string][]byte{
					"auth_method": []byte("offline-token"),
					"ocmAPIToken": []byte("dummytoken"),
				},
			},
			want: auth.AuthRequest{
				AuthMethod: "offline-token",
				Token:      "dummytoken",
			},
			wantErr: false,
		},
		{
			name: "Missing token",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
			},
			want: auth.AuthRequest{
				AuthMethod: "offline-token",
				Token:      "",
			},
			wantErr: true,
		},
		{
			name: "Dummy service account token",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				Data: map[string][]byte{
					"auth_method":   []byte("service-account"),
					"client_id":     []byte("dc05925d-630b-408b-bfb7-02099be7b789"),
					"client_secret": []byte("ZZocNUZWgYSuJHIqK0j0D1mZVdufng6z"), // notsecret
				},
			},
			want: auth.AuthRequest{
				AuthMethod: "service-account",
				ID:         "dc05925d-630b-408b-bfb7-02099be7b789",
				Secret:     "ZZocNUZWgYSuJHIqK0j0D1mZVdufng6z", // notsecret
			},
			wantErr: false,
		},
		{
			name: "Missing field service account",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				Data: map[string][]byte{
					"auth_method": []byte("service-account"),
					"client_id":   []byte("dc05925d-630b-408b-bfb7-02099be7b789"),
				},
			},
			want: auth.AuthRequest{
				AuthMethod: "service-account",
			},
			wantErr: true,
		},
		{
			name: "Invalid authentication method",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				Data: map[string][]byte{
					"auth_method": []byte("invalid-method"),
				},
			},
			want: auth.AuthRequest{
				AuthMethod: "invalid-method",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSecretForAuth(tt.secret)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSecretForAuth() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseSecretForAuth() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_assignManagedStatus(t *testing.T) {
	discovered := map[string]discovery.DiscoveredCluster{
		"a": {
			ObjectMeta: metav1.ObjectMeta{
				Name:      "a",
				Namespace: "test",
				Labels:    map[string]string{"isManagedCluster": "false"},
			},
			Spec: discovery.DiscoveredClusterSpec{
				IsManagedCluster: false,
			},
		},
		"b": {
			ObjectMeta: metav1.ObjectMeta{
				Name:      "b",
				Namespace: "test",
			},
		},
		"c": {
			ObjectMeta: metav1.ObjectMeta{
				Name:      "c",
				Namespace: "test",
			},
		},
	}

	managed := []metav1.PartialObjectMetadata{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "a",
				Labels: map[string]string{"clusterID": "a"},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "b",
				Labels: map[string]string{"clusterID": "b"},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "d",
				Labels: map[string]string{"clusterID": "d"},
			},
		},
	}

	assignManagedStatus(discovered, managed)

	t.Run("Cluster managed status changed", func(t *testing.T) {
		dc := discovered["a"]
		managedLabel := dc.GetLabels()["isManagedCluster"]
		if !dc.Spec.IsManagedCluster || managedLabel != "true" {
			t.Errorf("Expected cluster %s to be labeled as managed: %+v", dc.Name, dc)
		}
	})
	t.Run("Cluster managed status added", func(t *testing.T) {
		dc := discovered["b"]
		managedLabel := dc.GetLabels()["isManagedCluster"]
		if !dc.Spec.IsManagedCluster || managedLabel != "true" {
			t.Errorf("Expected cluster %s to be labeled as managed: %+v", dc.Name, dc)
		}
	})
	t.Run("Cluster managed status not added", func(t *testing.T) {
		dc := discovered["c"]
		managedLabel := dc.GetLabels()["isManagedCluster"]
		if dc.Spec.IsManagedCluster || managedLabel == "true" {
			t.Errorf("Expected cluster %s to not be labeled as managed: %+v", dc.Name, dc)
		}
	})
	t.Run("Discovered list not added to", func(t *testing.T) {
		if len(discovered) != 3 {
			t.Errorf("The discoveredlist should not change in size. Wanted: %d. Got: %d.", 3, len(discovered))
		}
	})
}

func Test_getURLOverride(t *testing.T) {
	tests := []struct {
		name   string
		config *discovery.DiscoveryConfig
		want   string
	}{
		{
			name: "Override annotated",
			config: &discovery.DiscoveryConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "testconfig",
					Namespace:   "test",
					Annotations: map[string]string{baseURLAnnotation: "http://mock-ocm-server.test.svc.cluster.local:3000"},
				},
			},
			want: "http://mock-ocm-server.test.svc.cluster.local:3000",
		},
		{
			name: "No override specified",
			config: &discovery.DiscoveryConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testconfig",
					Namespace: "test",
				},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getURLOverride(tt.config); got != tt.want {
				t.Errorf("getURLOverride() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getAuthURLOverride(t *testing.T) {
	//config *discovery.DiscoveryConfig
	tests := []struct {
		name   string
		config *discovery.DiscoveryConfig
		want   string
	}{
		{
			name: "Override annotated",
			config: &discovery.DiscoveryConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "testconfig",
					Namespace:   "test",
					Annotations: map[string]string{baseAuthURLAnnotation: "http://mock-ocm-server.test.svc.cluster.local:3000"},
				},
			},
			want: "http://mock-ocm-server.test.svc.cluster.local:3000",
		},
		{
			name: "No override specified",
			config: &discovery.DiscoveryConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testconfig",
					Namespace: "test",
				},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getAuthURLOverride(tt.config); got != tt.want {
				t.Errorf("getAuthURLOverride() = %v, want %v", got, tt.want)
			}
		})
	}
}

func countDiscoveredClusters(namespace string) (int, error) {
	discoveredClusters := &discovery.DiscoveredClusterList{}
	err := k8sClient.List(ctx, discoveredClusters, client.InNamespace(namespace))
	if err != nil {
		return -1, err
	}
	return len(discoveredClusters.Items), nil
}

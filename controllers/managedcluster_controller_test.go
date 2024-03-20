// Copyright Contributors to the Open Cluster Management project

package controllers

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	discovery "github.com/stolostron/discovery/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	TestManagedConfigName = "discovery"
	TestManagedNamespace  = "managed"
	TestManagedSecretName = "test-connection-secret"
)

var (
	mockManagedTime = metav1.NewTime(time.Date(2020, 5, 29, 6, 0, 0, 0, time.UTC))
	mockCluster49   = discovery.DiscoveredCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "c3po",
			Namespace: TestManagedNamespace,
		},
		Spec: discovery.DiscoveredClusterSpec{
			Name:              "c3po",
			DisplayName:       "c3po",
			OpenshiftVersion:  "4.9.0",
			CreationTimestamp: &mockManagedTime,
			ActivityTimestamp: &mockManagedTime,
		},
	}
)

var _ = Describe("ManagedCluster controller", func() {

	Context("Creating a DiscoveryConfig", func() {
		It("Should create discovered clusters", func() {
			By("By creating a namespace", func() {
				err := k8sClient.Create(ctx, &corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{Name: TestManagedNamespace},
				})
				Expect(err).NotTo(HaveOccurred())
			})

			By("By creating a secret with OCM credentials", func() {
				Expect(k8sClient.Create(ctx, &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      TestManagedSecretName,
						Namespace: TestManagedNamespace,
					},
					StringData: map[string]string{
						"ocmAPIToken": "dummytoken",
					},
				})).Should(Succeed())
			})

			mockDiscoveredCluster = func() ([]discovery.DiscoveredCluster, error) {
				return []discovery.DiscoveredCluster{
					mockCluster49,
				}, nil
			}

			By("By creating a new DiscoveryConfig", func() {
				Expect(k8sClient.Create(ctx, &discovery.DiscoveryConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name:      TestManagedConfigName,
						Namespace: TestManagedNamespace,
					},
					Spec: discovery.DiscoveryConfigSpec{
						Credential: TestManagedSecretName,
						Filters:    discovery.Filter{LastActive: 7},
					},
				})).Should(Succeed())
			})

			By("By checking 1 unmanaged discovered cluster has been created", func() {
				Eventually(func() (int, error) {
					return countDiscoveredClusters(TestManagedNamespace)
				}, timeout, interval).Should(Equal(1))

				Eventually(func() (int, error) {
					discoveredClusters := &discovery.DiscoveredClusterList{}
					err := k8sClient.List(ctx, discoveredClusters,
						client.InNamespace(TestManagedNamespace),
						client.MatchingLabels{
							"isManagedCluster": "true",
						})
					if err != nil {
						return -1, err
					}
					return len(discoveredClusters.Items), err
				}, timeout, interval).Should(Equal(0))
			})

			By("By creating a new ManagedCluster", func() {
				Expect(k8sClient.Create(ctx, newManagedCluster("mc-connection-1", "c3po"))).To(Succeed())
			})

			By("By checking 1 discovered cluster is labeled as managed", func() {
				Eventually(func() (int, error) {
					discoveredClusters := &discovery.DiscoveredClusterList{}
					err := k8sClient.List(ctx, discoveredClusters,
						client.InNamespace(TestManagedNamespace),
						client.MatchingLabels{
							"isManagedCluster": "true",
						})
					if err != nil {
						return -1, err
					}
					return len(discoveredClusters.Items), err
				}, timeout, interval).Should(Equal(1))
			})

			By("By deleting the ManagedCluster", func() {
				Expect(k8sClient.Delete(ctx, newManagedCluster("mc-connection-1", "c3po"))).To(Succeed())
			})

			By("By checking 0 discovered clusters labeled as managed", func() {
				Eventually(func() (int, error) {
					discoveredClusters := &discovery.DiscoveredClusterList{}
					err := k8sClient.List(ctx, discoveredClusters,
						client.InNamespace(TestManagedNamespace),
						client.MatchingLabels{
							"isManagedCluster": "true",
						})
					if err != nil {
						return -1, err
					}
					return len(discoveredClusters.Items), err
				}, timeout, interval).Should(Equal(0))
			})

		})
	})

})

func Test_getClusterID(t *testing.T) {
	tests := []struct {
		name           string
		managedCluster metav1.PartialObjectMetadata
		want           string
	}{
		{
			name: "Managed cluster without labels populated",
			managedCluster: metav1.PartialObjectMetadata{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "managedcluster1",
					Namespace: "test",
				},
			},
			want: "",
		},
		{
			name: "Managed cluster with labels populated",
			managedCluster: metav1.PartialObjectMetadata{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "managedcluster1",
					Namespace: "test",
					Labels:    map[string]string{"clusterID": "cluster-id-1"},
				},
			},
			want: "cluster-id-1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getClusterID(tt.managedCluster); got != tt.want {
				t.Errorf("getClusterID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getDiscoveredID(t *testing.T) {
	tests := []struct {
		name string
		dc   *discovery.DiscoveredCluster
		want string
	}{
		{
			name: "Managed cluster without labels populated",
			dc: &discovery.DiscoveredCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "Test A",
					Namespace: "b",
				},
				Spec: discovery.DiscoveredClusterSpec{
					Name: "managedcluster1",
				},
			},
			want: "managedcluster1",
		},
		{
			name: "Managed cluster without labels populated",
			dc: &discovery.DiscoveredCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "Test B",
					Namespace: "c",
				},
				Spec: discovery.DiscoveredClusterSpec{
					Name: "managedcluster2",
				},
			},
			want: "managedcluster2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getDiscoveredID(*tt.dc); got != tt.want {
				t.Errorf("getDiscoveredID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_setManagedStatus(t *testing.T) {
	tests := []struct {
		name string
		dc   *discovery.DiscoveredCluster
		want bool
	}{
		{
			name: "Managed labels set",
			dc: &discovery.DiscoveredCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "a",
					Namespace: "b",
					Labels:    map[string]string{"isManagedCluster": "true"},
				},
				Spec: discovery.DiscoveredClusterSpec{
					IsManagedCluster: true,
				},
			},
			want: false,
		},
		{
			name: "Managed label missing",
			dc: &discovery.DiscoveredCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "a",
					Namespace: "b",
					Labels:    map[string]string{"isManagedCluster": "false"},
				},
				Spec: discovery.DiscoveredClusterSpec{
					IsManagedCluster: false,
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := setManagedStatus(tt.dc); got != tt.want {
				t.Errorf("setManagedStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_unsetManagedStatus(t *testing.T) {
	tests := []struct {
		name string
		dc   *discovery.DiscoveredCluster
		want bool
	}{
		{
			name: "Managed labels set",
			dc: &discovery.DiscoveredCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "a",
					Namespace: "b",
					Labels:    map[string]string{"isManagedCluster": "true"},
				},
				Spec: discovery.DiscoveredClusterSpec{
					IsManagedCluster: true,
				},
			},
			want: true,
		},
		{
			name: "Managed label missing",
			dc: &discovery.DiscoveredCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "a",
					Namespace: "b",
					Labels:    map[string]string{"isManagedCluster": "false"},
				},
				Spec: discovery.DiscoveredClusterSpec{
					IsManagedCluster: false,
				},
			},
			want: false,
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := unsetManagedStatus(tt.dc); got != tt.want {
				t.Errorf("unsetManagedStatus() = %v, want %v", got, tt.want)
			}
		})
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

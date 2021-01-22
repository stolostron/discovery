package controllers

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/open-cluster-management/discovery/pkg/api/domain/cluster_domain"
	"github.com/open-cluster-management/discovery/pkg/api/services/auth_service"
	"github.com/open-cluster-management/discovery/pkg/api/services/cluster_service"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	discoveryv1 "github.com/open-cluster-management/discovery/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var (
	getTokenFunc    func(string) (string, error)
	getClustersFunc func() ([]cluster_domain.Cluster, error)
	clusterGetter   = clusterGetterMock{}
)

// This mocks the authService request and returns a dummy access token
type authServiceMock struct{}

func (m *authServiceMock) GetToken(token string) (string, error) {
	return getTokenFunc(token)
}

// The mocks the GetClusters request to return a select few clusters without connection
// to an external datasource
type clusterGetterMock struct{}

func (m *clusterGetterMock) GetClusters() ([]cluster_domain.Cluster, error) {
	return getClustersFunc()
}

// This mocks the NewClient function and returns an instance of the clusterGetterMock
type clusterClientGeneratorMock struct{}

func (m *clusterClientGeneratorMock) NewClient(config cluster_domain.ClusterRequest) cluster_service.ClusterGetter {
	return &clusterGetter
}

var _ = Describe("DiscoveryConfig controller", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		DiscoveryConfigName = "discoveryconfig"
		DiscoveryNamespace  = "default"
		SecretName          = "test-connection-secret"

		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When creating a DiscoveryConfig", func() {
		// this mock return a dummy token
		getTokenFunc = func(string) (string, error) {
			return "valid_access_token", nil
		}
		auth_service.AuthClient = &authServiceMock{}                           // Mocks out the call to auth service
		cluster_service.ClusterClientGenerator = &clusterClientGeneratorMock{} // Mocks out the cluster client creation

		It("Should trigger the creation of new discovered clusters ", func() {
			// this mock returns 3 clusters read from mock_clusters.json
			getClustersFunc = func() ([]cluster_domain.Cluster, error) {
				file, _ := ioutil.ReadFile("mock_clusters.json")
				clusters := []cluster_domain.Cluster{}
				_ = json.Unmarshal([]byte(file), &clusters)
				return clusters, nil
			}

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

			By("By checking that at least one discovered cluster has been created (without needing a discovery refresh)")
			discoveredClusters := &discoveryv1.DiscoveredClusterList{}
			Eventually(func() (int, error) {
				err := k8sClient.List(ctx, discoveredClusters, client.InNamespace(DiscoveryNamespace))
				if err != nil {
					return 0, err
				}
				return len(discoveredClusters.Items), nil
			}, time.Second*15, interval).Should(BeNumerically(">", 0))

			By("By verifying owner references are set on the new discovered clusters")
			for _, c := range discoveredClusters.Items {
				Expect(c.OwnerReferences).ToNot(BeNil())
			}
		})

		It("Should reconcile differences between existing clusters and new discovered clusters ", func() {
			// this mock returns 3 clusters read from mock_clusters2.json
			// it adds 1 new cluster, removes 1 old cluster, and leaves 1 cluster unchanged
			getClustersFunc = func() ([]cluster_domain.Cluster, error) {
				file, _ := ioutil.ReadFile("mock_clusters_2.json")
				clusters := []cluster_domain.Cluster{}
				_ = json.Unmarshal([]byte(file), &clusters)
				return clusters, nil
			}

			By("By creating a new DiscoveryRefresh")
			ctx := context.Background()
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

			Expect(k8sClient.Create(ctx, refresh)).Should(Succeed())

			By("By checking that one of the previous discovered clusters has been removed")
			deletedLookupKey := types.NamespacedName{Name: "temporarycluster", Namespace: DiscoveryNamespace}
			deletedCluster := &discoveryv1.DiscoveredCluster{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, deletedLookupKey, deletedCluster)
				return errors.IsNotFound(err)
			}, time.Second*15, interval).Should(BeTrue())

			By("By checking that one of the new discovered clusters has been created")
			createdLookupKey := types.NamespacedName{Name: "newclusterid", Namespace: DiscoveryNamespace}
			createdCluster := &discoveryv1.DiscoveredCluster{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, createdLookupKey, createdCluster)
				if err != nil {
					return false
				}
				return true
			}, time.Second*15, interval).Should(BeTrue())

			By("By checking that one of the previous discovered clusters has been updated")
			updatedLookupKey := types.NamespacedName{Name: "tobeupdatedcluster", Namespace: DiscoveryNamespace}
			updatedCluster := &discoveryv1.DiscoveredCluster{}
			Eventually(func() string {
				err := k8sClient.Get(ctx, updatedLookupKey, updatedCluster)
				if err != nil {
					return ""
				}
				return updatedCluster.Spec.OpenshiftVersion
			}, time.Second*15, interval).Should(Equal("4.4.0"))

		})
	})
})

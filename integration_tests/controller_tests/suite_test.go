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

package controller_tests

import (
	"context"
	"flag"
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	discoveryv1 "github.com/open-cluster-management/discovery/api/v1"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment
var reportFile string

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(discoveryv1.AddToScheme(scheme))

	utilruntime.Must(corev1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func init() {
	flag.StringVar(&reportFile, "report-file", "../results/functional-results.xml", "Provide the path to where the junit results will be printed.")
}

func TestAPIs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter(reportFile)
	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{junitReporter})
}

var _ = BeforeSuite(func() {
	ctrl.SetLogger(
		zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)),
	)

	By("bootstrapping test environment")
	c, err := config.GetConfig()
	Expect(err).ToNot(HaveOccurred())
	Expect(c).ToNot(BeNil())

	k8sClient, err = client.New(c, client.Options{Scheme: scheme})
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sClient).ToNot(BeNil())

})

// var _ = BeforeSuite(func() {
// 	// Add any setup steps that needs to be executed before each test
// 	By("Cleaning up test objects")
// 	ctx := context.Background()

// 	discoveryRefresh := &discoveryv1.DiscoveredClusterRefresh{}
// 	k8sClient.DeleteAllOf(ctx, discoveryRefresh, client.InNamespace("open-cluster-management"))

// 	discoveryConfig := &discoveryv1.DiscoveryConfig{}
// 	k8sClient.DeleteAllOf(ctx, discoveryConfig, client.InNamespace("open-cluster-management"))

// 	discoveredCluster := &discoveryv1.DiscoveredCluster{}
// 	k8sClient.DeleteAllOf(ctx, discoveredCluster, client.InNamespace("open-cluster-management"))

// 	secretKey := types.NamespacedName{Name: SecretName, Namespace: DiscoveryNamespace}
// 	secret := &corev1.Secret{}
// 	_ = k8sClient.Get(ctx, secretKey, secret)
// 	k8sClient.Delete(ctx, secret)
// })

var _ = AfterSuite(func() {
	// Add any teardown steps that needs to be executed after each test
	By("Cleaning up test objects")
	ctx := context.Background()

	discoveryRefresh := &discoveryv1.DiscoveredClusterRefresh{}
	k8sClient.DeleteAllOf(ctx, discoveryRefresh, client.InNamespace("open-cluster-management"))

	discoveryConfig := &discoveryv1.DiscoveryConfig{}
	k8sClient.DeleteAllOf(ctx, discoveryConfig, client.InNamespace("open-cluster-management"))

	discoveredCluster := &discoveryv1.DiscoveredCluster{}
	k8sClient.DeleteAllOf(ctx, discoveredCluster, client.InNamespace("open-cluster-management"))

	secretKey := types.NamespacedName{Name: SecretName, Namespace: DiscoveryNamespace}
	secret := &corev1.Secret{}
	_ = k8sClient.Get(ctx, secretKey, secret)
	k8sClient.Delete(ctx, secret)
})

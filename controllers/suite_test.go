// Copyright Contributors to the Open Cluster Management project

package controllers

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "github.com/stolostron/discovery/api/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/event"

	clusterapiv1 "open-cluster-management.io/api/cluster/v1"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

func TestControllers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Suite")
}

var k8sClient client.Client
var testEnv *envtest.Environment
var signalHandlerContext context.Context

var _ = BeforeSuite(func() {
	log.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	// SetupSignalHandler can only be called once, so we'll save the
	// context it returns and reuse it each time we start a new
	// manager.
	signalHandlerContext = ctrl.SetupSignalHandler()

	By("bootstrap test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "config", "crd", "bases"),
			filepath.Join("..", "testserver", "build"),
		},
		CRDInstallOptions: envtest.CRDInstallOptions{
			CleanUpAfterUse: true,
		},
		ErrorIfCRDPathMissing: true,
	}

	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = v1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = clusterapiv1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = scheme.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:                 scheme.Scheme,
		MetricsBindAddress:     "0",
		HealthProbeBindAddress: "0",
	})
	Expect(err).ToNot(HaveOccurred())

	err = (&DiscoveryConfigReconciler{
		Client: k8sManager.GetClient(),
		Scheme: k8sManager.GetScheme(),
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	events := make(chan event.GenericEvent)
	err = (&ManagedClusterReconciler{
		Client:  k8sManager.GetClient(),
		Scheme:  k8sManager.GetScheme(),
		Trigger: events,
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	Expect(os.Setenv("UNIT_TEST", "true")).To(Succeed())

	go func() {
		// For explanation of GinkgoRecover in a go routine, see
		// https://onsi.github.io/ginkgo/#mental-model-how-ginkgo-handles-failure
		defer GinkgoRecover()
		err = k8sManager.Start(signalHandlerContext)
		Expect(err).ToNot(HaveOccurred())
	}()
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	Expect(os.Unsetenv("UNIT_TEST")).To(Succeed())
})

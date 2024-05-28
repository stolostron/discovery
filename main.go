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

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	goruntime "runtime"

	"go.uber.org/zap/zapcore"
	admissionregistration "k8s.io/api/admissionregistration/v1"
	apixv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
	clusterapiv1beta1 "open-cluster-management.io/api/cluster/v1beta1"
	clusterapiv1beta2 "open-cluster-management.io/api/cluster/v1beta2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	klusterletconfigv1alpha1 "github.com/stolostron/cluster-lifecycle-api/klusterletconfig/v1alpha1"
	discoveryv1 "github.com/stolostron/discovery/api/v1"
	"github.com/stolostron/discovery/controllers"
	agentv1 "github.com/stolostron/klusterlet-addon-controller/pkg/apis/agent/v1"
	corev1 "k8s.io/api/core/v1"
	clusterapiv1 "open-cluster-management.io/api/cluster/v1"
	// +kubebuilder:scaffold:imports
)

const (
	crdName = "discoveredclusters.discovery.open-cluster-management.io"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")

	ControllerError = "unable to create controller"
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(addonv1alpha1.AddToScheme(scheme))
	utilruntime.Must(apixv1.AddToScheme(scheme))
	utilruntime.Must(clusterapiv1.AddToScheme(scheme))
	utilruntime.Must(corev1.AddToScheme(scheme))
	utilruntime.Must(discoveryv1.AddToScheme(scheme))
	utilruntime.Must(agentv1.SchemeBuilder.AddToScheme(scheme))
	utilruntime.Must(klusterletconfigv1alpha1.AddToScheme(scheme))
	utilruntime.Must(clusterapiv1beta1.AddToScheme(scheme))
	utilruntime.Must(clusterapiv1beta2.AddToScheme(scheme))

	utilruntime.Must(apiextv1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var leaseDuration time.Duration
	var renewDeadline time.Duration
	var retryPeriod time.Duration
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.DurationVar(&leaseDuration, "leader-election-lease-duration", 137*time.Second, ""+
		"The duration that non-leader candidates will wait after observing a leadership "+
		"renewal until attempting to acquire leadership of a led but unrenewed leader "+
		"slot. This is effectively the maximum duration that a leader can be stopped "+
		"before it is replaced by another candidate. This is only applicable if leader "+
		"election is enabled.")
	flag.DurationVar(&renewDeadline, "leader-election-renew-deadline", 107*time.Second, ""+
		"The interval between attempts by the acting master to renew a leadership slot "+
		"before it stops leading. This must be less than or equal to the lease duration. "+
		"This is only applicable if leader election is enabled.")
	flag.DurationVar(&retryPeriod, "leader-election-retry-period", 26*time.Second, ""+
		"The duration the clients should wait between attempting acquisition and renewal "+
		"of a leadership. This is only applicable if leader election is enabled.")
	opts := zap.Options{
		Development: true,
		TimeEncoder: zapcore.ISO8601TimeEncoder,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	setupLog.Info(fmt.Sprintf("Go Version: %s", goruntime.Version()))
	setupLog.Info(fmt.Sprintf("Go OS/Arch: %s/%s", goruntime.GOOS, goruntime.GOARCH))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: metricsAddr,
		},
		WebhookServer: webhook.NewServer(webhook.Options{
			Port: 9443,
		}),
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "744aebb6.open-cluster-management.io",
		LeaseDuration:          &leaseDuration,
		RenewDeadline:          &renewDeadline,
		RetryPeriod:            &retryPeriod,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	uncachedClient, err := client.New(ctrl.GetConfigOrDie(), client.Options{
		Scheme: scheme,
	})
	if err != nil {
		setupLog.Error(err, "unable to create uncached client")
		os.Exit(1)
	}

	events := make(chan event.GenericEvent)

	if err = (&controllers.DiscoveryConfigReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, ControllerError, "controller", "DiscoveryConfig")
		os.Exit(1)
	}

	if err = (&controllers.DiscoveredClusterReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, ControllerError, "controller", "DiscoveredCluster")
		os.Exit(1)
	}

	if err = (&controllers.ManagedClusterReconciler{
		Client:  mgr.GetClient(),
		Scheme:  mgr.GetScheme(),
		Trigger: events,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, ControllerError, "controller", "ManagedCluster")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	if os.Getenv("ENABLE_WEBHOOKS") != "false" {
		// https://book.kubebuilder.io/cronjob-tutorial/running.html#running-webhooks-locally
		// https://book.kubebuilder.io/multiversion-tutorial/webhooks.html#and-maingo
		if err = ensureWebhooks(uncachedClient); err != nil {
			setupLog.Error(err, "unable to ensure webhook", "webhook", "DiscoveredCluster")
			os.Exit(1)
		}

		if err = (&discoveryv1.DiscoveredCluster{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "DiscoveredCluster")
			os.Exit(1)
		}
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func ensureWebhooks(k8sClient client.Client) error {
	ctx := context.Background()

	deploymentNamespace, ok := os.LookupEnv("POD_NAMESPACE")
	if !ok {
		setupLog.Info("Failing due to being unable to locate webhook service namespace")
		os.Exit(1)
	}
	validatingWebhook := discoveryv1.ValidatingWebhook(deploymentNamespace)

	maxAttempts := 10
	for i := 0; i < maxAttempts; i++ {
		setupLog.Info("Applying ValidatingWebhookConfiguration")

		// Get reference to DiscoveredCluster CRD to set as owner of the webhook
		// This way if the CRD is deleted the webhook will be removed with it
		crdKey := types.NamespacedName{Name: crdName}
		owner := &apixv1.CustomResourceDefinition{}
		if err := k8sClient.Get(context.TODO(), crdKey, owner); err != nil {
			setupLog.Error(err, "Failed to get DiscoveredCluster CRD")
			time.Sleep(5 * time.Second)
			continue
		}
		validatingWebhook.SetOwnerReferences([]metav1.OwnerReference{
			{
				APIVersion: "apiextensions.k8s.io/v1",
				Kind:       "CustomResourceDefinition",
				Name:       owner.Name,
				UID:        owner.UID,
			},
		})

		existingWebhook := &admissionregistration.ValidatingWebhookConfiguration{}
		existingWebhook.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   "admissionregistration.k8s.io",
			Version: "v1",
			Kind:    "ValidatingWebhookConfiguration",
		})
		err := k8sClient.Get(ctx, types.NamespacedName{Name: validatingWebhook.GetName()}, existingWebhook)
		if err != nil && errors.IsNotFound(err) {
			// Webhook not found. Create and return
			err = k8sClient.Create(ctx, validatingWebhook)
			if err != nil {
				setupLog.Error(err, "Error creating validatingwebhookconfiguration")
				time.Sleep(5 * time.Second)
				continue
			}
			return nil

		} else if err != nil {
			setupLog.Error(err, "Error getting validatingwebhookconfiguration")
			time.Sleep(5 * time.Second)
			continue

		} else if err == nil {
			// Webhook already exists. Update and return
			setupLog.Info("Updating existing validatingwebhookconfiguration")
			existingWebhook.Webhooks = validatingWebhook.Webhooks
			err = k8sClient.Update(ctx, existingWebhook)
			if err != nil {
				setupLog.Error(err, "Error updating validatingwebhookconfiguration")
				time.Sleep(5 * time.Second)
				continue
			}
			return nil
		}
	}
	return fmt.Errorf("unable to ensure validatingwebhook exists in allotted time")
}

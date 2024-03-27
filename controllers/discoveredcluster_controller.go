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

	"github.com/pkg/errors"
	discovery "github.com/stolostron/discovery/api/v1"
	"github.com/stolostron/discovery/util/reconciler"
	agentv1 "github.com/stolostron/klusterlet-addon-controller/pkg/apis/agent/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clusterapiv1 "open-cluster-management.io/api/cluster/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// DiscoveredClusterReconciler reconciles a DiscoveredCluster object
type DiscoveredClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *DiscoveredClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logf.Info("Reconciling DiscoveredCluster")

	config := &discovery.DiscoveryConfig{}
	err := r.Get(ctx, req.NamespacedName, config)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logf.Info("DiscoveryConfig resource not found. Ignoring since object may have been deleted.")
			return ctrl.Result{}, nil
		}

		// If there's an error other than "Not Found", return with the error.
		return ctrl.Result{}, fmt.Errorf("failed to get DiscoveryConfig %s: %w", req.Name, err)
	}

	ocmAPISecret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{
		Name: config.Spec.Credential, Namespace: config.GetNamespace(),
	}}

	if err := r.Get(ctx, types.NamespacedName{
		Name: ocmAPISecret.GetName(), Namespace: ocmAPISecret.GetNamespace(),
	}, ocmAPISecret); err != nil {
		logf.Info(fmt.Sprintf("failed to get secret %v: %v", ocmAPISecret.GetName(), err))
	}

	discoveredClusters := &discovery.DiscoveredClusterList{}
	if err := r.List(ctx, discoveredClusters); err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue.
			logf.Info("DiscoveredCluster resources not found. Ignoring since objects must be deleted")
			return ctrl.Result{}, nil
		}

		// Error reading the object - requeue the request.
		return ctrl.Result{}, errors.Wrapf(err, "Failed to list DiscoveredCluster CR")
	}

	/*
		Note: We are only interested in automatically importing ROSA cluster, since we currently cannot import other
		cloud services without additional configurations.
	*/
	filteredClusters := r.FilterByRosa(ctx, *discoveredClusters)
	if len(filteredClusters) == 0 {
		logf.Info("No ROSA cluster found. Scheduling next reconcilation for", "next_time", reconciler.RefreshInterval)
		return ctrl.Result{RequeueAfter: reconciler.RefreshInterval}, nil
	}

	for _, dc := range filteredClusters {
		if dc.GetAnnotations() == nil {
			dc.Annotations = make(map[string]string)
		}

		/*
			Check to see if the discovered cluster has a import strategy defined. If it does not, then add the default
			import strategy to the discovered cluster.
		*/
		if dc.Annotations[discovery.ImportStrategyAnnotation] == "" {
			if _, err := r.ApplyDefaultImportStrategy(ctx, dc); err != nil {
				logf.Error(err, "failed to apply default import strategy", "Name", dc.Spec.DisplayName)
				continue // If we weren't able to apply the import strategy, don't fail but move on to the next cluster.
			}
		}

		logf.Info("dc annotation", "Name", dc.Spec.DisplayName, "annotation", dc.GetAnnotations())

		/*
			If the discovered cluster has an Automatic import strategy, we need to ensure that the required resources
			are available. Otherwise, we will ignore that cluster.
		*/
		if !dc.Spec.IsManagedCluster && dc.Annotations[discovery.ImportStrategyAnnotation] == "Automatic" {
			if res, err := r.EnsureNamespaceForDiscoveredCluster(ctx, dc); err != nil {
				logf.Error(err, "failed to ensure namespace for discovered cluster", "Name", dc.Spec.DisplayName)
				return res, err
			}

			if res, err := r.EnsureManagedCluster(ctx, dc); err != nil {
				logf.Error(err, "failed to ensure managed cluster created", "Name", dc.Spec.DisplayName)
				return res, err
			}

			if res, err := r.EnsureKlusterletAddonConfig(ctx, dc); err != nil {
				logf.Error(err, "failed to ensure klusterlet addon config created", "Name", dc.Spec.DisplayName)
				return res, err
			}

			if res, err := r.EnsureAutoImportSecret(ctx, dc, config); err != nil {
				logf.Error(err, "failed to ensure auto-import secret created", "Name", dc.Spec.DisplayName)
				return res, err
			}
		}
	}

	return ctrl.Result{RequeueAfter: reconciler.RefreshInterval}, nil
}

// ApplyDefaultImportStrategy ...
func (r *DiscoveredClusterReconciler) ApplyDefaultImportStrategy(ctx context.Context, dc discovery.DiscoveredCluster) (
	ctrl.Result, error) {
	nn := types.NamespacedName{Name: dc.GetName(), Namespace: dc.GetNamespace()}

	if err := r.Get(ctx, nn, &dc); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, errors.Wrapf(err, "discovered cluster %s/%s was not found", dc.GetNamespace(),
				dc.GetName())
		}

		return ctrl.Result{}, errors.Wrapf(err, "error fetching discovered cluster %s/%s", dc.GetNamespace(),
			dc.GetName())
	}

	// Create a copy of the DiscoveredCluster object to only modify the annotations field
	modifiedDC := dc.DeepCopy()
	if modifiedDC.GetAnnotations() == nil {
		modifiedDC.SetAnnotations(make(map[string]string))
	}
	modifiedDC.GetAnnotations()[discovery.ImportStrategyAnnotation] = "Manual"

	// Apply the object data
	if err := r.Patch(ctx, modifiedDC, client.MergeFrom(&dc), &client.PatchOptions{FieldManager: "discovery-operator"}); err != nil {
		return ctrl.Result{}, errors.Wrapf(err, "error patching discovered cluster %s/%s", dc.GetNamespace(), dc.GetName())
	}
	return ctrl.Result{}, nil
}

// EnsureManagedCluster ...
func (r *DiscoveredClusterReconciler) EnsureManagedCluster(ctx context.Context, dc discovery.DiscoveredCluster) (
	ctrl.Result, error) {
	nn := types.NamespacedName{Name: dc.Spec.DisplayName}
	existingMC := clusterapiv1.ManagedCluster{}

	if err := r.Get(ctx, nn, &existingMC); err != nil {
		if apierrors.IsNotFound(err) {
			logf.Info("Creating managed cluster", "Name", nn.Name)

			mc := clusterapiv1.ManagedCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: nn.Name,
					Labels: map[string]string{
						"name":   nn.Name,
						"cloud":  "auto-detect",
						"vendor": "auto-detect",
					},
					Annotations: map[string]string{
						"open-cluster-management/created-via": "discovery",
					},
					// Finalizers: []string{
					// 	discovery.ImportCleanUpFinalizer,
					// },
				},
				Spec: clusterapiv1.ManagedClusterSpec{
					HubAcceptsClient: true,
				},
			}

			controllerutil.SetControllerReference(&dc, &mc, r.Scheme)
			if err := r.Create(ctx, &mc, &client.CreateOptions{}); err != nil {
				logf.Error(err, "failed to create managed cluster", "Name", nn.Name)
				return ctrl.Result{RequeueAfter: reconciler.ResyncPeriod}, err
			}
		}

		logf.Error(err, "failed to get managed cluster resource", "Name", nn.Name)
		return ctrl.Result{RequeueAfter: reconciler.ResyncPeriod}, err
	}
	return ctrl.Result{}, nil
}

// EnsureKlusterletAddonConfig ...
func (r *DiscoveredClusterReconciler) EnsureKlusterletAddonConfig(ctx context.Context, dc discovery.DiscoveredCluster) (
	ctrl.Result, error) {
	nn := types.NamespacedName{Name: dc.Spec.DisplayName, Namespace: dc.Spec.DisplayName}
	existingKCA := agentv1.KlusterletAddonConfig{}

	if err := r.Get(ctx, nn, &existingKCA); err != nil {
		if apierrors.IsNotFound(err) {
			logf.Info("Creating klusterlet addon config", "Name", nn.Name)

			kca := agentv1.KlusterletAddonConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      nn.Name,
					Namespace: nn.Namespace,
				},
				Spec: agentv1.KlusterletAddonConfigSpec{
					ClusterName:      nn.Name,
					ClusterNamespace: nn.Namespace,
					ClusterLabels: map[string]string{
						"name":   nn.Name,
						"cloud":  "auto-detect",
						"vendor": "auto-detect",
					},
					ApplicationManagerConfig: agentv1.KlusterletAddonAgentConfigSpec{
						Enabled: true,
					},
					CertPolicyControllerConfig: agentv1.KlusterletAddonAgentConfigSpec{
						Enabled: true,
					},
					IAMPolicyControllerConfig: agentv1.KlusterletAddonAgentConfigSpec{
						Enabled: true,
					},
					PolicyController: agentv1.KlusterletAddonAgentConfigSpec{
						Enabled: true,
					},
					SearchCollectorConfig: agentv1.KlusterletAddonAgentConfigSpec{
						Enabled: true,
					},
				},
			}

			controllerutil.SetControllerReference(&dc, &kca, r.Scheme)
			if err := r.Create(ctx, &kca, &client.CreateOptions{}); err != nil {
				logf.Error(err, "failed to create klusterlet addon config", "Name", nn.Name)
				return ctrl.Result{RequeueAfter: reconciler.ResyncPeriod}, err
			}
		}

		logf.Error(err, "failed to get klusterlet addon config", "Name", nn.Name)
		return ctrl.Result{RequeueAfter: reconciler.ResyncPeriod}, err
	}

	return ctrl.Result{}, nil
}

// EnsureNamespaceForDiscoveredCluster ...
func (r *DiscoveredClusterReconciler) EnsureNamespaceForDiscoveredCluster(ctx context.Context,
	dc discovery.DiscoveredCluster) (ctrl.Result, error) {
	nn := types.NamespacedName{Name: dc.Spec.DisplayName}
	existingNs := &corev1.Namespace{}

	if err := r.Get(ctx, nn, existingNs); err != nil {
		if apierrors.IsNotFound(err) {
			ns := r.GenerateNamespaceForDiscoveredCluster(ctx, dc)
			logf.Info("Creating namespace for discovered cluster", "Name", ns.GetName())

			if err := r.Create(ctx, ns, &client.CreateOptions{}); err != nil {
				return ctrl.Result{RequeueAfter: reconciler.ResyncPeriod}, err
			}
		} else {
			logf.Error(err, "Failed to check if namespace exists", "Name", nn.Name)
			return ctrl.Result{RequeueAfter: reconciler.ResyncPeriod}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *DiscoveredClusterReconciler) EnsureFinalizerRemovedFromManagedCluster(ctx context.Context,
	mc clusterapiv1.ManagedCluster) (ctrl.Result, error) {
	nn := types.NamespacedName{Name: mc.GetName()}

	if err := r.Get(ctx, nn, &mc, &client.GetOptions{}); err != nil {
		if apierrors.IsNotFound(err) {
			logf.Info("Managed cluster not found", "Name", nn.Name)

		} else {
			logf.Error(err, "Failed to get managed cluster", "Name", nn.Name)
		}
	}

	if mc.DeletionTimestamp != nil && controllerutil.ContainsFinalizer(&mc, discovery.ImportCleanUpFinalizer) {
		controllerutil.RemoveFinalizer(&mc, discovery.ImportCleanUpFinalizer)
		if err := r.Update(ctx, &mc); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *DiscoveredClusterReconciler) EnsureAutoImportSecret(ctx context.Context, dc discovery.DiscoveredCluster,
	config *discovery.DiscoveryConfig) (ctrl.Result, error) {
	nn := types.NamespacedName{Name: "auto-import-secret", Namespace: dc.Spec.DisplayName}
	existingSecret := corev1.Secret{}

	if err := r.Get(ctx, nn, &existingSecret, &client.GetOptions{}); err != nil {
		if apierrors.IsNotFound(err) {
			logf.Info("Creating auto-import-secret for managed cluster", "Namespace", nn.Namespace)

			s := corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      nn.Name,
					Namespace: nn.Namespace,
				},
				StringData: map[string]string{
					"api_url":     "",
					"api_token":   config.Spec.Credential,
					"token_url":   "",
					"cluster_id":  "",
					"retry_times": "2",
				},
				Type: "auto-import/rosa",
			}

			controllerutil.SetOwnerReference(&dc, &s, r.Scheme)
			if err := r.Create(ctx, &s, &client.CreateOptions{}); err != nil {
				logf.Error(err, "failed to create auto-import-secret for managed cluster", "Name", nn.Name)
				return ctrl.Result{RequeueAfter: reconciler.ResyncPeriod}, err
			}
		}

		logf.Error(err, "failed to get auto-import-secret for managed cluster", "Name", nn.Name)
		return ctrl.Result{RequeueAfter: reconciler.ResyncPeriod}, err
	}

	return ctrl.Result{}, nil
}

// FilterByRosa ...
func (r *DiscoveredClusterReconciler) FilterByRosa(ctx context.Context, discoveredList discovery.DiscoveredClusterList,
) []discovery.DiscoveredCluster {
	rosaClusters := []discovery.DiscoveredCluster{}

	if len(discoveredList.Items) == 0 {
		return rosaClusters
	}

	for _, dc := range discoveredList.Items {
		if dc.Spec.Type == "ROSA" {
			rosaClusters = append(rosaClusters, dc)
		}
	}

	logf.Info("ROSA clusters filtered succesfully", "count", len(rosaClusters))
	return rosaClusters
}

// GenerateNamespaceForDiscoveredClusters ...
func (r *DiscoveredClusterReconciler) GenerateNamespaceForDiscoveredCluster(ctx context.Context,
	dc discovery.DiscoveredCluster) *corev1.Namespace {
	ns := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: dc.Spec.DisplayName,
			Labels: map[string]string{
				"openshift.io/cluster-monitoring": "true",
			},
		},
	}

	controllerutil.SetControllerReference(&dc, ns, r.Scheme)
	return ns
}

// SetupWithManager ...
func (r *DiscoveredClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&discovery.DiscoveredCluster{}).
		Complete(r)
}

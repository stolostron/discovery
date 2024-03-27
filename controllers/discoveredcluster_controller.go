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
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// DiscoveredClusterReconciler reconciles a DiscoveredCluster object
type DiscoveredClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *DiscoveredClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logf.Info("Reconciling DiscoveredCluster", "Name", req.Name, "Namespace", req.Namespace)
	if req.Name == "" || req.Namespace == "" {
		return ctrl.Result{}, nil
	}

	dc := &discovery.DiscoveredCluster{}
	if err := r.Get(ctx, req.NamespacedName, dc); err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue.
			logf.Info("DiscoveredCluster resource not found. Ignoring since objects must be deleted",
				"Name", req.Name, "Namespace", req.Namespace)
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, errors.Wrap(err, "Failed to get DiscoveredCluster CR")
	}

	if dc.GetAnnotations() == nil {
		dc.Annotations = make(map[string]string)
	}

	/*
		Check to see if the discovered cluster has a import strategy defined. If it does not, then add the default
		import strategy to the discovered cluster.
	*/
	if dc.Annotations[discovery.ImportStrategyAnnotation] == "" {
		if _, err := r.ApplyDefaultImportStrategy(ctx, *dc); err != nil {
			logf.Error(err, "failed to apply default import strategy", "Name", dc.Spec.DisplayName)
			return ctrl.Result{Requeue: true}, err
		}

		logf.Info("Applied default import strategy annotation for DiscoveredCluster", "Name", dc.Spec.DisplayName,
			"Namespace", dc.GetNamespace())
	}

	/*
		If the discovered cluster has an Automatic import strategy, we need to ensure that the required resources
		are available. Otherwise, we will ignore that cluster.
	*/
	if !dc.Spec.IsManagedCluster && dc.Annotations[discovery.ImportStrategyAnnotation] == "Automatic" {
		if res, err := r.EnsureNamespaceForDiscoveredCluster(ctx, *dc); err != nil {
			logf.Error(err, "failed to ensure namespace for DiscoveredCluster", "Name", dc.Spec.DisplayName)
			return res, err
		}

		if res, err := r.EnsureManagedCluster(ctx, *dc); err != nil {
			logf.Error(err, "failed to ensure managed cluster created", "Name", dc.Spec.DisplayName)
			return res, err
		}

		if res, err := r.EnsureKlusterletAddonConfig(ctx, *dc); err != nil {
			logf.Error(err, "failed to ensure klusterlet addon config created", "Name", dc.Spec.DisplayName)
			return res, err
		}

		if res, err := r.EnsureAutoImportSecret(ctx, *dc); err != nil {
			logf.Error(err, "failed to ensure auto-import secret created", "Name", dc.Spec.DisplayName)
			return res, err
		}
	}

	return ctrl.Result{RequeueAfter: reconciler.RefreshInterval}, nil
}

/*
ApplyDefaultImportStrategy applies the default import strategy to the discovered cluster.
It sets the import strategy annotation to "Manual" for the specified DiscoveredCluster object.
If the cluster is not found, it returns an error wrapped with a message indicating the absence of the cluster.
If any other error occurs during fetching or patching, it returns an error wrapped with a corresponding message.
*/
func (r *DiscoveredClusterReconciler) ApplyDefaultImportStrategy(ctx context.Context, dc discovery.DiscoveredCluster) (
	ctrl.Result, error) {
	nn := types.NamespacedName{Name: dc.GetName(), Namespace: dc.GetNamespace()}

	if err := r.Get(ctx, nn, &dc); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, errors.Wrapf(err, "DiscoveredCluster %s/%s was not found", dc.GetNamespace(),
				dc.GetName())
		}

		return ctrl.Result{}, errors.Wrapf(err, "error fetching DiscoveredCluster %s/%s", dc.GetNamespace(),
			dc.GetName())
	}

	// Create a copy of the DiscoveredCluster object to only modify the annotations field.
	modifiedDC := dc.DeepCopy()

	if modifiedDC.GetAnnotations() == nil {
		modifiedDC.SetAnnotations(make(map[string]string))
	}
	modifiedDC.GetAnnotations()[discovery.ImportStrategyAnnotation] = "Manual"

	if err := r.Patch(ctx, modifiedDC, client.MergeFrom(&dc),
		&client.PatchOptions{FieldManager: "discovery-operator"}); err != nil {

		return ctrl.Result{Requeue: true}, errors.Wrapf(err, "error patching DiscoveredCluster %s/%s", dc.GetNamespace(),
			dc.GetName())
	}
	return ctrl.Result{}, nil
}

/*
CreateAutoImportSecret creates an auto-import secret for the given NamespacedName and DiscoveryConfig.
It constructs a Secret object with the specified name, namespace, and credential from the DiscoveryConfig.
Other fields like api_url, token_url, and cluster_id are left empty.
It sets the retry_times field to a default value of "2".
The secret type is set to "auto-import/rosa".
*/
func (r *DiscoveredClusterReconciler) CreateAutoImportSecret(nn types.NamespacedName) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nn.Name,
			Namespace: nn.Namespace,
		},
		StringData: map[string]string{
			"api_url":     "",
			"api_token":   "",
			"token_url":   "",
			"cluster_id":  "",
			"retry_times": "2",
		},
		Type: "auto-import/rosa",
	}
}

/*
createKlusterletAddonConfig creates a KlusterletAddonConfig object with the specified NamespacedName.
It sets the basic configuration for the KlusterletAddonConfig, including metadata and spec fields.
It also initializes the ClusterLabels with default values and enables various addon configurations.
*/
func (r *DiscoveredClusterReconciler) CreateKlusterletAddonConfig(nn types.NamespacedName,
) *agentv1.KlusterletAddonConfig {
	return &agentv1.KlusterletAddonConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nn.Name,
			Namespace: nn.Namespace,
		},
		Spec: agentv1.KlusterletAddonConfigSpec{
			ClusterName:      nn.Name,
			ClusterNamespace: nn.Namespace,
			ClusterLabels: map[string]string{
				"name":   nn.Name,
				"cloud":  discovery.AutoDetectLabels,
				"vendor": discovery.AutoDetectLabels,
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
}

/*
CreateManagedCluster creates a ManagedCluster object with the specified NamespacedName. It sets default labels and
annotations for the ManagedCluster. It also adds a finalizer for cleanup.
*/
func (r *DiscoveredClusterReconciler) CreateManagedCluster(nn types.NamespacedName) *clusterapiv1.ManagedCluster {
	return &clusterapiv1.ManagedCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: nn.Name,
			Labels: map[string]string{
				"name":   nn.Name,
				"cloud":  discovery.AutoDetectLabels,
				"vendor": discovery.AutoDetectLabels,
			},
			Annotations: map[string]string{
				discovery.CreatedViaAnnotation: "discovery",
			},
			Finalizers: []string{
				discovery.ImportCleanUpFinalizer,
			},
		},
		Spec: clusterapiv1.ManagedClusterSpec{
			HubAcceptsClient: true,
		},
	}
}

/*
CreateNamespaceForDiscoveredCluster creates a Namespace object for the specified DiscoveredCluster.
It sets the Namespace's metadata including Name and Labels.
The Namespace is labeled for cluster monitoring with a label indicating its purpose.
*/
func (r *DiscoveredClusterReconciler) CreateNamespaceForDiscoveredCluster(dc discovery.DiscoveredCluster,
) *corev1.Namespace {
	return &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: dc.Spec.DisplayName,
			Labels: map[string]string{
				discovery.ClusterMonitoringLabel: "true",
			},
		},
	}
}

/*
EnsureAutoImportSecret ensures the existence of an auto-import secret for the given DiscoveredCluster and
DiscoveryConfig. It checks if an auto-import secret with the specified name and namespace exists.
If not found, it creates a new auto-import secret with the name "auto-import-secret" and the DiscoveredCluster's
namespace. It sets a controller reference to the DiscoveredCluster for ownership management.
If creation fails, it logs an error and returns with a requeue signal. If the auto-import secret already exists or if
an error occurs during retrieval, it logs an error and returns with a requeue signal.
*/
func (r *DiscoveredClusterReconciler) EnsureAutoImportSecret(ctx context.Context, dc discovery.DiscoveredCluster,
) (ctrl.Result, error) {
	nn := types.NamespacedName{Name: "auto-import-secret", Namespace: dc.Spec.DisplayName}
	existingSecret := corev1.Secret{}

	if err := r.Get(ctx, nn, &existingSecret, &client.GetOptions{}); err != nil && apierrors.IsNotFound(err) {
		logf.Info("Creating auto-import-secret for managed cluster", "Namespace", nn.Namespace)

		s := r.CreateAutoImportSecret(nn)
		controllerutil.SetControllerReference(&dc, s, r.Scheme)

		if err := r.Create(ctx, s); err != nil {
			logf.Error(err, "failed to create auto-import Secret for ManagedCluster", "Name", nn.Name)
			return ctrl.Result{RequeueAfter: reconciler.ResyncPeriod}, err
		}
	} else if err != nil {
		logf.Error(err, "failed to get auto-import Secret for ManagedCluster", "Name", nn.Name)
		return ctrl.Result{RequeueAfter: reconciler.ResyncPeriod}, err
	}

	return ctrl.Result{}, nil
}

// EnsureKlusterletAddonConfig ensures the existence of a KlusterletAddonConfig resource for the given DiscoveredCluster.
// It checks if a KlusterletAddonConfig with the specified display name exists.
// If not found, it creates a new KlusterletAddonConfig with the display name and default configurations.
// It sets a controller reference to the DiscoveredCluster for ownership management.
// If creation fails, it logs an error and returns with a requeue signal.
// If the KlusterletAddonConfig already exists or if an error occurs during retrieval, it logs an error and returns with a
// requeue signal.
func (r *DiscoveredClusterReconciler) EnsureKlusterletAddonConfig(ctx context.Context, dc discovery.DiscoveredCluster) (
	ctrl.Result, error) {
	nn := types.NamespacedName{Name: dc.Spec.DisplayName, Namespace: dc.Spec.DisplayName}
	existingKCA := agentv1.KlusterletAddonConfig{}

	if err := r.Get(ctx, nn, &existingKCA); err != nil && apierrors.IsNotFound(err) {
		logf.Info("Creating KlusterletAddonConfig", "Name", nn.Name, "Namespace", nn.Namespace)

		kca := r.CreateKlusterletAddonConfig(nn)
		controllerutil.SetControllerReference(&dc, kca, r.Scheme)

		if err := r.Create(ctx, kca); err != nil {
			logf.Error(err, "failed to create KlusterAddonConfig", "Name", nn.Name, "Namespace", nn.Namespace)
			return ctrl.Result{RequeueAfter: reconciler.ResyncPeriod}, err
		}

	} else if err != nil {
		logf.Error(err, "failed to get KlusterAddonConfig", "Name", nn.Name)
		return ctrl.Result{RequeueAfter: reconciler.ResyncPeriod}, err
	}

	return ctrl.Result{}, nil
}

// EnsureManagedCluster ensures the existence of a ManagedCluster resource for the given DiscoveredCluster.
// It checks if a ManagedCluster with the specified display name exists.
// If not found, it creates a new ManagedCluster with the display name and default configurations.
// It sets a controller reference to the DiscoveredCluster for ownership management.
// If creation fails, it logs an error and returns with a requeue signal.
// If the ManagedCluster already exists or if an error occurs during retrieval, it logs an error and returns with a
// requeue signal.
func (r *DiscoveredClusterReconciler) EnsureManagedCluster(ctx context.Context, dc discovery.DiscoveredCluster) (
	ctrl.Result, error) {
	nn := types.NamespacedName{Name: dc.Spec.DisplayName}
	existingMC := &clusterapiv1.ManagedCluster{}

	if err := r.Get(ctx, nn, existingMC); err != nil && apierrors.IsNotFound(err) {
		logf.Info("Creating ManagedCluster", "Name", nn.Name)

		mc := r.CreateManagedCluster(nn)
		controllerutil.SetControllerReference(&dc, mc, r.Scheme)

		if err := r.Create(ctx, mc); err != nil {
			logf.Error(err, "failed to create ManagedCluster", "Name", nn.Name)
			return ctrl.Result{RequeueAfter: reconciler.ResyncPeriod}, err
		}

	} else if err != nil {
		logf.Error(err, "failed to get ManagedCluster", "Name", nn.Name)
		return ctrl.Result{RequeueAfter: reconciler.ResyncPeriod}, err
	}

	return ctrl.Result{}, nil
}

// EnsureNamespaceForDiscoveredCluster ensures the existence of a Namespace for the given DiscoveredCluster.
// It checks if a Namespace with the specified display name exists.
// If not found, it creates a new Namespace with the display name.
// It sets a controller reference to the DiscoveredCluster for ownership management.
// If creation fails, it returns with a requeue signal.
// If an error occurs during retrieval or creation, it logs an error and returns with a requeue signal.
func (r *DiscoveredClusterReconciler) EnsureNamespaceForDiscoveredCluster(ctx context.Context,
	dc discovery.DiscoveredCluster) (ctrl.Result, error) {
	nn := types.NamespacedName{Name: dc.Spec.DisplayName}
	existingNs := &corev1.Namespace{}

	if err := r.Get(ctx, nn, existingNs); err != nil && apierrors.IsNotFound(err) {
		logf.Info("Creating Namespace for DiscoveredCluster", "Name", nn.Name)

		ns := r.CreateNamespaceForDiscoveredCluster(dc)
		controllerutil.SetControllerReference(&dc, ns, r.Scheme)

		if err := r.Create(ctx, ns); err != nil {
			logf.Error(err, "failed to create Namespace", "Name", nn.Name)
			return ctrl.Result{RequeueAfter: reconciler.ResyncPeriod}, err
		}

	} else if err != nil {
		logf.Error(err, "Failed to get Namespace", "Name", nn.Name)
		return ctrl.Result{RequeueAfter: reconciler.ResyncPeriod}, err
	}

	return ctrl.Result{}, nil
}

// EnsureFinalizerRemovedFromManagedCluster ...
func (r *DiscoveredClusterReconciler) EnsureFinalizerRemovedFromManagedCluster(ctx context.Context,
	mc clusterapiv1.ManagedCluster) (ctrl.Result, error) {
	nn := types.NamespacedName{Name: mc.GetName()}

	if err := r.Get(ctx, nn, &mc, &client.GetOptions{}); err != nil {
		if !apierrors.IsNotFound(err) {
			logf.Error(err, "failed to get ManagedCluster", "Name", nn.Name)
		}

		return ctrl.Result{}, err
	}

	if mc.DeletionTimestamp != nil && controllerutil.ContainsFinalizer(&mc, discovery.ImportCleanUpFinalizer) {
		controllerutil.RemoveFinalizer(&mc, discovery.ImportCleanUpFinalizer)
		if err := r.Update(ctx, &mc); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager ...
func (r *DiscoveredClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&discovery.DiscoveredCluster{}).
		WithEventFilter(predicate.Funcs{
			CreateFunc: func(e event.CreateEvent) bool {
				return r.ShouldReconcile(e.Object)
			},
			UpdateFunc: func(e event.UpdateEvent) bool {
				return r.ShouldReconcile(e.ObjectNew)
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				return true
			},
		}).
		Complete(r)
}

// ShouldReconcile ...
func (r *DiscoveredClusterReconciler) ShouldReconcile(obj metav1.Object) bool {
	dc, ok := obj.(*discovery.DiscoveredCluster)
	if !ok {
		return false
	}

	return dc.Spec.Type == "ROSA"
}

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
	"os"

	"github.com/pkg/errors"
	discovery "github.com/stolostron/discovery/api/v1"
	utils "github.com/stolostron/discovery/util"
	recon "github.com/stolostron/discovery/util/reconciler"
	agentv1 "github.com/stolostron/klusterlet-addon-controller/pkg/apis/agent/v1"
	corev1 "k8s.io/api/core/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
	clusterapiv1 "open-cluster-management.io/api/cluster/v1"
	clusterapiv1beta1 "open-cluster-management.io/api/cluster/v1beta1"
	clusterapiv1beta2 "open-cluster-management.io/api/cluster/v1beta2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// DiscoveredClusterReconciler reconciles a DiscoveredCluster object
type DiscoveredClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const (
	AddOnDeploymentConfigName = "addon-ns-config"
	DefaultName               = "default"
)

// +kubebuilder:rbac:groups=discovery.open-cluster-management.io,resources=discoveredclusters,verbs=create;delete;deletecollection;get;list;patch;update;watch
// +kubebuilder:rbac:groups=discovery.open-cluster-management.io,resources=discoveredclusters/finalizers,verbs=update
// +kubebuilder:rbac:groups=discovery.open-cluster-management.io,resources=discoveredclusters/status,verbs=get;patch;update
// +kubebuilder:rbac:groups=discovery.open-cluster-management.io,resources=discoveredclusters/finalizers,verbs=get;patch;update
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch
// +kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=validatingwebhookconfigurations,verbs=create;get;list;update;watch
// +kubebuilder:rbac:groups=config.open-cluster-management.io,resources=klusterletconfigs,verbs=create;get;list;patch;update;
// +kubebuilder:rbac:groups=addon.open-cluster-management.io,resources=addondeploymentconfigs;clustermanagementaddons,verbs=create;get;list;update;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DiscoveredCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
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
		If the discovered cluster has an Automatic import strategy, we need to ensure that the required resources
		are available. Otherwise, we will ignore that cluster.
	*/
	if !dc.Spec.IsManagedCluster && dc.Spec.ImportAsManagedCluster {
		if !utils.IsAnnotationTrue(dc, utils.AnnotationPreviouslyAutoImported) {
			switch dc.Spec.Type {
			case "MultiClusterEngineHCP":
				if res, err := r.EnsureMultiClusterEngineHCP(ctx, dc); err != nil {
					return res, err
				}

			case "ROSA":
				if res, err := r.EnsureROSA(ctx, dc); err != nil {
					return res, err
				}

			default:
				logf.Info("Unknown cluster type. Skipping automatic import.", "Name", dc.Spec.DisplayName,
					"Type", dc.Spec.Type)
			}
		} else {
			logf.Info(
				fmt.Sprintf("Skipped automatic import for DiscoveredCluster due to existing '%v' annotation",
					utils.AnnotationPreviouslyAutoImported), "Name", dc.Spec.DisplayName)
		}
	}

	return ctrl.Result{RequeueAfter: recon.ShortRefreshInterval}, nil
}

/*
CreateAddOnDeploymentConfig creates a AddOnDeploymentConfig object with the specified NamespacedName.
It sets the basic configuration for the AddOnDeploymentConfig, including metadata and spec fields.
It also initializes the AgentInstallNamespace with a default value.
*/
func (r *DiscoveredClusterReconciler) CreateAddOnDeploymentConfig(nn types.NamespacedName,
) *addonv1alpha1.AddOnDeploymentConfig {
	return &addonv1alpha1.AddOnDeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nn.Name,
			Namespace: nn.Namespace,
		},
		Spec: addonv1alpha1.AddOnDeploymentConfigSpec{
			AgentInstallNamespace: "open-cluster-management-discovered-hcp",
		},
	}
}

/*
CreateAutoImportSecret creates an auto-import secret for the given NamespacedName and DiscoveryConfig.
It constructs a Secret object with the specified name, namespace, and credential from the DiscoveryConfig.
Other fields like api_url, token_url, and cluster_id are left empty.
It sets the retry_times field to a default value of "2".
The secret type is set to "auto-import/rosa".
*/
func (r *DiscoveredClusterReconciler) CreateAutoImportSecret(nn types.NamespacedName, clusterID, apiToken string,
) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nn.Name,
			Namespace: nn.Namespace,
		},
		StringData: map[string]string{
			"api_token":  apiToken,
			"cluster_id": clusterID,
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
func (r *DiscoveredClusterReconciler) CreateManagedCluster(nn types.NamespacedName,
	clusterType string) *clusterapiv1.ManagedCluster {
	annotations := make(map[string]string)
	labels := map[string]string{
		utils.LabelName:   nn.Name,
		utils.LabelCloud:  discovery.AutoDetectLabels,
		utils.LabelVendor: discovery.AutoDetectLabels,
	}

	switch clusterType {
	case "MultiClusterEngineHCP":
		annotations[utils.AnnotationHostingClusterName] = nn.Namespace
		annotations[utils.AnnotationKlusterletDeployMode] = "Hosted"
		annotations[utils.AnnotationCreatedVia] = "hypershift"

		// Set specific label for MultiClusterEngineHCP
		labels[utils.LabelHypershiftDiscoveryType] = "MultiClusterEngineHCP"

	default:
		annotations[utils.AnnotationCreatedVia] = "discovery"
	}

	return &clusterapiv1.ManagedCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:        nn.Name,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: clusterapiv1.ManagedClusterSpec{
			HubAcceptsClient:     true,
			LeaseDurationSeconds: 60,
		},
	}
}

/*
CreateManagedClusterSetBinding ...
*/
func (r *DiscoveredClusterReconciler) CreateManagedClusterSetBinding(nn types.NamespacedName,
) *clusterapiv1beta2.ManagedClusterSetBinding {
	return &clusterapiv1beta2.ManagedClusterSetBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nn.Name,
			Namespace: nn.Namespace,
		},
		Spec: clusterapiv1beta2.ManagedClusterSetBindingSpec{
			ClusterSet: DefaultName,
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
CreatePlacement creates a Placement object for the ClusterManagementAddOn.
*/
func (r *DiscoveredClusterReconciler) CreatePlacement(nn types.NamespacedName) *clusterapiv1beta1.Placement {
	return &clusterapiv1beta1.Placement{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nn.Name,
			Namespace: nn.Namespace,
		},
		Spec: clusterapiv1beta1.PlacementSpec{
			DecisionStrategy: clusterapiv1beta1.DecisionStrategy{
				GroupStrategy: clusterapiv1beta1.GroupStrategy{
					ClustersPerDecisionGroup: intstr.FromString("100%"),
				},
			},
			Predicates: []clusterapiv1beta1.ClusterPredicate{
				{
					RequiredClusterSelector: clusterapiv1beta1.ClusterSelector{
						LabelSelector: metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      utils.LabelHypershiftDiscoveryType,
									Operator: metav1.LabelSelectorOpIn,
									Values:   []string{"MultiClusterEngineHCP"},
								},
							},
						},
					},
				},
			},
		},
	}
}

/*
EnsureAddOnDeploymentConfig ensures the existence of a AddOnDeploymentConfig resource for the given
DiscoveredCluster. It checks if a AddOnDeploymentConfig with the specified name exists.
If not found, it creates a new AddOnDeploymentConfig with the name and default configurations.
If creation fails, it logs an error and returns with a requeue signal.
If the AddOnDeploymentConfig already exists or if an error occurs during retrieval, it logs an error and returns
with a requeue signal.
*/
func (r *DiscoveredClusterReconciler) EnsureAddOnDeploymentConfig(ctx context.Context) (
	ctrl.Result, error) {
	nn := types.NamespacedName{Name: AddOnDeploymentConfigName, Namespace: os.Getenv("POD_NAMESPACE")}
	existingADC := addonv1alpha1.AddOnDeploymentConfig{}

	if err := r.Get(ctx, nn, &existingADC); apierrors.IsNotFound(err) {
		logf.Info("Creating AddOnDeploymentConfig", "Name", nn.Name, "Namespace", nn.Namespace)

		adc := r.CreateAddOnDeploymentConfig(nn)
		if err := r.Create(ctx, adc); err != nil {
			logf.Error(err, "failed to create AddOnDeploymentConfig", "Name", nn.Name)
			return ctrl.Result{RequeueAfter: recon.ErrorRefreshInterval}, err
		}

	} else if err != nil {
		logf.Error(err, "failed to get AddOnDeploymentConfig", "Name", nn.Name)
		return ctrl.Result{RequeueAfter: recon.WarningRefreshInterval}, err
	}

	return ctrl.Result{}, nil
}

/*
EnsureAutoImportSecret ensures the existence of an auto-import secret for the given DiscoveredCluster and
DiscoveryConfig. It checks if an auto-import secret with the specified name and namespace exists.
If not found, it creates a new auto-import secret with the name "auto-import-secret" and the DiscoveredCluster's
namespace. It sets a controller reference to the DiscoveredCluster for ownership management.
If creation fails, it logs an error and returns with a requeue signal. If the auto-import secret already exists or if
an error occurs during retrieval, it logs an error and returns with a requeue signal.
*/
func (r *DiscoveredClusterReconciler) EnsureAutoImportSecret(ctx context.Context, dc discovery.DiscoveredCluster) (
	ctrl.Result, error) {
	nn := types.NamespacedName{Name: dc.Spec.Credential.Name, Namespace: dc.Spec.Credential.Namespace}
	existingSecret := corev1.Secret{}

	if err := r.Get(ctx, nn, &existingSecret); apierrors.IsNotFound(err) {
		logf.Error(err, "Secret was not found", "Name", nn.Name, "Namespace", nn.Namespace)
		return ctrl.Result{RequeueAfter: recon.ShortRefreshInterval}, err

	} else if err != nil {
		logf.Error(err, "failed to get Secret", "Name", nn.Name, "Namespace", nn.Namespace)
		return ctrl.Result{RequeueAfter: recon.WarningRefreshInterval}, err
	}

	if apiToken, err := parseUserToken(&existingSecret); err == nil {
		nn := types.NamespacedName{Name: "auto-import-secret", Namespace: dc.Spec.DisplayName}
		existingSecret = corev1.Secret{}

		if err := r.Get(ctx, nn, &existingSecret, &client.GetOptions{}); apierrors.IsNotFound(err) {
			logf.Info("Creating auto-import-secret for managed cluster", "Namespace", nn.Namespace)

			s := r.CreateAutoImportSecret(nn, dc.Spec.RHOCMClusterID, apiToken)
			if err := r.Create(ctx, s); err != nil {
				logf.Error(err, "failed to create auto-import Secret for ManagedCluster", "Name", nn.Name)
				return ctrl.Result{RequeueAfter: recon.ErrorRefreshInterval}, err
			}

		} else if err != nil {
			logf.Error(err, "failed to get auto-import Secret for ManagedCluster", "Name", nn.Name)
			return ctrl.Result{RequeueAfter: recon.WarningRefreshInterval}, err
		}
	} else {
		logf.Error(err, "failed to parse token from Secret", "Name", nn.Name)
		return ctrl.Result{RequeueAfter: recon.WarningRefreshInterval}, err
	}

	return ctrl.Result{}, nil
}

/*
EnsureCommonResources ...
*/
func (r *DiscoveredClusterReconciler) EnsureCommonResources(ctx context.Context,
	dc *discovery.DiscoveredCluster, isHCP bool) (ctrl.Result, error) {
	if isHCP {
		if res, err := r.EnsureManagedClusterSetBinding(ctx); err != nil {
			logf.Error(err, "failed to ensure ManagedClusterBindingSet created", "Name", DefaultName,
				"Namespace", os.Getenv("POD_NAMESPACE"))
			return res, err
		}

		if res, err := r.EnsurePlacement(ctx); err != nil {
			logf.Error(err, "failed to ensure Placement created", "Name", DefaultName,
				"Namespace", os.Getenv("POD_NAMESPACE"))
			return res, err
		}

		if res, err := r.EnsureAddOnDeploymentConfig(ctx); err != nil {
			logf.Error(err, "failed to ensure AddOnDeploymentConfig created", "Name", AddOnDeploymentConfigName,
				"Namespace", os.Getenv("POD_NAMESPACE"))
			return res, err
		}

		addonNames := []string{"cluster-proxy", "managed-serviceaccount", "work-manager"}
		for _, addon := range addonNames {
			if res, err := r.AddPlacementToClusterManagementAddOn(ctx, addon); err != nil {
				return res, err
			}
		}
	}

	if res, err := r.EnsureManagedCluster(ctx, *dc); err != nil {
		logf.Error(err, "failed to ensure ManagedCluster created", "Name", dc.Spec.DisplayName)
		return res, err
	}

	// Ensure that the KlusterletAddOnConfig CRD exists. In standalone MCE mode, the CRD is not deployed.
	crdName := "klusterletaddonconfigs.agent.open-cluster-management.io"
	if res, err := r.EnsureCRDExist(ctx, crdName); err != nil {
		if !apierrors.IsNotFound(err) {
			logf.Error(err, "failed to ensure custom resource definition exist", "Name", crdName)
			return res, err
		}
	} else {
		if res, err := r.EnsureKlusterletAddonConfig(ctx, *dc); err != nil {
			logf.Error(err, "failed to ensure KlusterletAddonConfig created", "Name", dc.Spec.DisplayName)
			return res, err
		}
	}

	if !isHCP {
		// Ensure that the DiscoveredCluster credentials are available on the cluster.
		if res, err := r.EnsureDiscoveredClusterCredentialExists(ctx, *dc); err != nil {
			logf.Error(err, "failed to ensure DiscoveredCluster credential Secret exist", "Name",
				dc.Spec.DisplayName)
			return res, err
		}

		if res, err := r.EnsureAutoImportSecret(ctx, *dc); err != nil {
			logf.Error(err, "failed to ensure auto import Secret created", "Name", dc.Spec.DisplayName)
			return res, err
		}
	}

	return ctrl.Result{}, nil
}

/*
EnsureCRDExist checks if a Custom Resource Definition (CRD) with the specified name exists in the cluster.
If the CRD exists, it returns indicating that the reconciliation should continue without requeueing.
If the CRD doesn't exist, it logs a message indicating that the CRD was not found and returns.
If an error occurs while getting the CRD (other than IsNotFound), it logs an error message and returns an error.
*/
func (r *DiscoveredClusterReconciler) EnsureCRDExist(ctx context.Context, crdName string) (ctrl.Result, error) {
	crd := &apiextv1.CustomResourceDefinition{}
	nn := types.NamespacedName{Name: crdName}

	if err := r.Get(ctx, nn, crd); apierrors.IsNotFound(err) {
		logf.Info("CRD not found. Ignoring since object must be deleted", "Name", crdName)
		return ctrl.Result{}, err

	} else if err != nil {
		logf.Error(err, "failed to get CRD", "Name", crdName)
		return ctrl.Result{RequeueAfter: recon.ShortRefreshInterval}, err
	}

	return ctrl.Result{}, nil
}

// EnsureDiscoveredClusterCredentialExists ...
func (r *DiscoveredClusterReconciler) EnsureDiscoveredClusterCredentialExists(
	ctx context.Context, dc discovery.DiscoveredCluster) (ctrl.Result, error) {
	nn := types.NamespacedName{Name: dc.Spec.Credential.Name, Namespace: dc.Spec.Credential.Namespace}
	secret := corev1.Secret{}

	if err := r.Get(ctx, nn, &secret); apierrors.IsNotFound(err) {
		logf.Error(err, "Secret was not found", "Name", nn.Name, "Namespace", nn.Namespace)
		return ctrl.Result{RequeueAfter: recon.ShortRefreshInterval}, err

	} else if err != nil {
		logf.Error(err, "failed to get Secret", "Name", nn.Name, "Namespace", nn.Namespace)
		return ctrl.Result{RequeueAfter: recon.WarningRefreshInterval}, err
	}

	return ctrl.Result{}, nil
}

/*
EnsureKlusterletAddonConfig ensures the existence of a KlusterletAddonConfig resource for the given
DiscoveredCluster. It checks if a KlusterletAddonConfig with the specified display name exists.
If not found, it creates a new KlusterletAddonConfig with the display name and default configurations.
It sets a controller reference to the DiscoveredCluster for ownership management.
If creation fails, it logs an error and returns with a requeue signal.
If the KlusterletAddonConfig already exists or if an error occurs during retrieval, it logs an error and returns
with a requeue signal.
*/
func (r *DiscoveredClusterReconciler) EnsureKlusterletAddonConfig(ctx context.Context, dc discovery.DiscoveredCluster) (
	ctrl.Result, error) {
	nn := types.NamespacedName{Name: dc.Spec.DisplayName, Namespace: dc.Spec.DisplayName}
	existingKAC := agentv1.KlusterletAddonConfig{}

	if err := r.Get(ctx, nn, &existingKAC); apierrors.IsNotFound(err) {
		logf.Info("Creating KlusterletAddonConfig", "Name", nn.Name, "Namespace", nn.Namespace)

		kac := r.CreateKlusterletAddonConfig(nn)
		if err := r.Create(ctx, kac); err != nil {
			logf.Error(err, "failed to create KlusterAddonConfig", "Name", nn.Name, "Namespace", nn.Namespace)
			return ctrl.Result{RequeueAfter: recon.ErrorRefreshInterval}, err
		}

	} else if err != nil {
		logf.Error(err, "failed to get KlusterAddonConfig", "Name", nn.Name)
		return ctrl.Result{RequeueAfter: recon.WarningRefreshInterval}, err
	}

	return ctrl.Result{}, nil
}

/*
EnsureManagedCluster ensures the existence of a ManagedCluster resource for the given DiscoveredCluster.
It checks if a ManagedCluster with the specified display name exists.
If not found, it creates a new ManagedCluster with the display name and default configurations.
It sets a controller reference to the DiscoveredCluster for ownership management.
If creation fails, it logs an error and returns with a requeue signal.
If the ManagedCluster already exists or if an error occurs during retrieval, it logs an error and returns with a
requeue signal.
*/
func (r *DiscoveredClusterReconciler) EnsureManagedCluster(ctx context.Context, dc discovery.DiscoveredCluster) (
	ctrl.Result, error) {
	// ManagedCluster resources are cluster scoped resources; therefore we do not need to specify the namespace.
	nn := types.NamespacedName{Name: dc.Spec.DisplayName}
	existingMC := &clusterapiv1.ManagedCluster{}

	if err := r.Get(ctx, nn, existingMC); apierrors.IsNotFound(err) {
		logf.Info("Creating ManagedCluster", "Name", nn.Name)
		nn.Namespace = dc.GetNamespace() // We are setting the namespace only for annotation purposes.

		mc := r.CreateManagedCluster(nn, dc.Spec.Type)
		if err := r.Create(ctx, mc); err != nil {
			logf.Error(err, "failed to create ManagedCluster", "Name", nn.Name)
			return ctrl.Result{RequeueAfter: recon.ErrorRefreshInterval}, err
		}

	} else if err != nil {
		logf.Error(err, "failed to get ManagedCluster", "Name", nn.Name)
		return ctrl.Result{RequeueAfter: recon.WarningRefreshInterval}, err
	}

	return ctrl.Result{}, nil
}

/*
EnsureManagedClusterSetBinding ensures the existence of a ManagedClusterSetBinding.
It checks if a ManagedClusterSetBinding with the specified display name exists.
If not found, it creates a new ManagedClusterSetBinding with the display name and default configurations.
If creation fails, it logs an error and returns with a requeue signal.
If the ManagedClusterSetBinding already exists or if an error occurs during retrieval, it logs an error and returns
with a requeue signal.
*/
func (r *DiscoveredClusterReconciler) EnsureManagedClusterSetBinding(ctx context.Context) (ctrl.Result, error) {
	nn := types.NamespacedName{Name: DefaultName, Namespace: os.Getenv("POD_NAMESPACE")}
	existingMCSB := &clusterapiv1beta2.ManagedClusterSetBinding{}

	if err := r.Get(ctx, nn, existingMCSB); apierrors.IsNotFound(err) {
		logf.Info("Creating ManagedClusterSetBinding", "Name", nn.Name, "Namespace", nn.Namespace)

		mc := r.CreateManagedClusterSetBinding(nn)
		if err := r.Create(ctx, mc); err != nil {
			logf.Error(err, "failed to create ManagedClusterSetBinding", "Name", nn.Name, "Namespace", nn.Namespace)
			return ctrl.Result{RequeueAfter: recon.ErrorRefreshInterval}, err
		}

	} else if err != nil {
		logf.Error(err, "failed to get ManagedClusterSetBinding", "Name", nn.Name, "Namespace", nn.Namespace)
		return ctrl.Result{RequeueAfter: recon.WarningRefreshInterval}, err
	}

	return ctrl.Result{}, nil
}

/*
EnsureMultiClusterEngineHCP ...
*/
func (r *DiscoveredClusterReconciler) EnsureMultiClusterEngineHCP(ctx context.Context, dc *discovery.DiscoveredCluster,
) (ctrl.Result, error) {
	return r.EnsureCommonResources(ctx, dc, true)
}

/*
EnsureNamespaceForDiscoveredCluster ensures the existence of a Namespace for the given DiscoveredCluster.
It checks if a Namespace with the specified display name exists.
If not found, it creates a new Namespace with the display name.
It sets a controller reference to the DiscoveredCluster for ownership management.
If creation fails, it returns with a requeue signal.
If an error occurs during retrieval or creation, it logs an error and returns with a requeue signal.
*/
func (r *DiscoveredClusterReconciler) EnsureNamespaceForDiscoveredCluster(ctx context.Context,
	dc discovery.DiscoveredCluster) (ctrl.Result, error) {
	nn := types.NamespacedName{Name: dc.Spec.DisplayName}
	existingNs := &corev1.Namespace{}

	if err := r.Get(ctx, nn, existingNs); apierrors.IsNotFound(err) {
		logf.Info("Creating Namespace for DiscoveredCluster", "Name", nn.Name)

		ns := r.CreateNamespaceForDiscoveredCluster(dc)
		if err := r.Create(ctx, ns); err != nil {
			logf.Error(err, "failed to create Namespace", "Name", nn.Name)
			return ctrl.Result{RequeueAfter: recon.ErrorRefreshInterval}, err
		}

	} else if err != nil {
		logf.Error(err, "failed to get Namespace", "Name", nn.Name)
		return ctrl.Result{RequeueAfter: recon.WarningRefreshInterval}, err
	}

	return ctrl.Result{}, nil
}

/*
EnsurePlacement ...
*/
func (r *DiscoveredClusterReconciler) EnsurePlacement(ctx context.Context) (ctrl.Result, error) {
	nn := types.NamespacedName{Name: DefaultName, Namespace: os.Getenv("POD_NAMESPACE")}
	existingPlacement := &clusterapiv1beta1.Placement{}

	if err := r.Get(ctx, nn, existingPlacement); apierrors.IsNotFound(err) {
		logf.Info("Creating Placement", "Name", nn.Name, "Namespace", nn.Namespace)

		placement := r.CreatePlacement(nn)
		if err := r.Create(ctx, placement); err != nil {
			logf.Error(err, "failed to create Placement", "Name", nn.Name, "Namespace", nn.Namespace)
			return ctrl.Result{RequeueAfter: recon.ErrorRefreshInterval}, err
		}

	} else if err != nil {
		logf.Error(err, "failed to get Placement", "Name", nn.Name, "Namespace", nn.Namespace)
		return ctrl.Result{RequeueAfter: recon.WarningRefreshInterval}, err
	}

	return ctrl.Result{}, nil
}

/*
EnsureROSA ...
*/
func (r *DiscoveredClusterReconciler) EnsureROSA(ctx context.Context, dc *discovery.DiscoveredCluster) (
	ctrl.Result, error) {
	if res, err := r.EnsureNamespaceForDiscoveredCluster(ctx, *dc); err != nil {
		logf.Error(err, "failed to ensure namespace for DiscoveredCluster", "Name", dc.Spec.DisplayName)
		return res, err
	}

	return r.EnsureCommonResources(ctx, dc, false)
}

/*
EnsurePlacementAddedToClusterManagementAddOn
*/
func (r *DiscoveredClusterReconciler) AddPlacementToClusterManagementAddOn(ctx context.Context, name string) (
	ctrl.Result, error) {
	cma := &addonv1alpha1.ClusterManagementAddOn{}

	if err := r.Get(ctx, types.NamespacedName{Name: name}, cma); err != nil {
		logf.Error(err, "failed to get clusterManagementAddOn", "Name", name)
		return ctrl.Result{RequeueAfter: recon.WarningRefreshInterval}, err
	}

	if cma.Spec.InstallStrategy.Type != "Placements" {
		cma.Spec.InstallStrategy.Type = "Placements"
	}

	placements := cma.Spec.InstallStrategy.Placements
	placementAvailable := false

	for _, p := range placements {
		if p.Name == DefaultName && p.Namespace == os.Getenv("POD_NAMESPACE") {
			placementAvailable = true
			break
		}
	}

	if !placementAvailable {
		placement := addonv1alpha1.PlacementStrategy{
			PlacementRef: addonv1alpha1.PlacementRef{
				Name:      DefaultName,
				Namespace: os.Getenv("POD_NAMESPACE"),
			},
			Configs: []addonv1alpha1.AddOnConfig{
				{
					ConfigReferent: addonv1alpha1.ConfigReferent{
						Name:      AddOnDeploymentConfigName,
						Namespace: os.Getenv("POD_NAMESPACE"),
					},
					ConfigGroupResource: addonv1alpha1.ConfigGroupResource{
						Group:    "addon.open-cluster-management.io",
						Resource: "addondeploymentconfigs",
					},
				},
			},
		}
		cma.Spec.InstallStrategy.Placements = append(cma.Spec.InstallStrategy.Placements, placement)

		if err := r.Update(ctx, cma); err != nil {
			logf.Error(err, "failed to patch ClusterManagementAddOn", "Name", cma.GetName())
			return ctrl.Result{RequeueAfter: recon.ErrorRefreshInterval}, err
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

	return discovery.IsSupportedClusterType(dc.Spec.Type)
}

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

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	discovery "github.com/stolostron/discovery/api/v1"
	utils "github.com/stolostron/discovery/util"
	recon "github.com/stolostron/discovery/util/reconciler"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clusterapiv1 "open-cluster-management.io/api/cluster/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var managedClusterGVK = schema.GroupVersionKind{
	Kind:    "ManagedCluster",
	Group:   "cluster.open-cluster-management.io",
	Version: "v1",
}

// ManagedClusterReconciler reconciles a ManagedCluster object
type ManagedClusterReconciler struct {
	client.Client
	Scheme  *runtime.Scheme
	Trigger chan event.GenericEvent
	Log     logr.Logger
}

// +kubebuilder:rbac:groups=agent.open-cluster-management.io,resources=klusterletaddonconfigs;klusterletaddonconfigs/finalizers;klusterletaddonconfigs/status,verbs=create;delete;get;list;patch;update;watch
// +kubebuilder:rbac:groups=cluster.open-cluster-management.io,resources=managedclusters;managedclusters/accept;managedclusters/finalizers;managedclustersets;managedclustersets/bind;managedclustersets/finalizers;managedclustersets/join;managedclustersetbindings;managedclustersetbindings/finalizers;placements;placements/finalizers;managedclusters/status,verbs=create;get;list;patch;update;watch
// +kubebuilder:rbac:groups=register.open-cluster-management.io,resources=managedclusters/accept,verbs=update

func (r *ManagedClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logf.Info("Reconciling ManagedCluster", "Name", req.Name)
	if req.Name == "" {
		return ctrl.Result{}, nil
	}

	discoveredClusters := &discovery.DiscoveredClusterList{}
	if err := r.List(ctx, discoveredClusters); err != nil {
		return ctrl.Result{}, errors.Wrapf(err, "error listing discovered clusters")
	}

	if len(discoveredClusters.Items) == 0 {
		return ctrl.Result{}, nil
	}

	mc := &clusterapiv1.ManagedCluster{}
	if err := r.Get(ctx, req.NamespacedName, mc); err != nil && !apierrors.IsNotFound(err) {
		logf.Error(err, "failed to get ManagedCluster", "Name", req.Name)
		return ctrl.Result{RequeueAfter: recon.WarningRefreshInterval}, err
	}

	if mc.GetDeletionTimestamp() != nil {
		logf.Info("ManagedCluster is being deleted", "Name", mc.GetName(), "DeletionTimestamp",
			mc.GetDeletionTimestamp())

		for i := range discoveredClusters.Items {
			dc := &discoveredClusters.Items[i]

			if dc.GetName() == req.Name || dc.Spec.DisplayName == req.Name {
				modifiedDC := dc.DeepCopy()

				/*
					Set annotation to true on DiscoveredCluster resource to prevent automatic import.
					The user will need to manually remove the annotation if they want the cluster to be reimported
					automatically.
				*/
				if modifiedDC.Spec.ImportAsManagedCluster {
					modifiedDC.SetAnnotations(map[string]string{
						utils.AnnotationPreviouslyAutoImported: "true",
					})

					logf.Info(fmt.Sprintf("Added '%v' annotation to DiscoveredCluster",
						utils.AnnotationPreviouslyAutoImported), "Name", dc.GetName())
				}
				modifiedDC.Spec.ImportAsManagedCluster = false

				if err := r.Patch(ctx, modifiedDC, client.MergeFrom(dc)); err != nil {
					logf.Error(err, "failed to patch DiscoveredCluster", "Name", dc.GetName())
					return ctrl.Result{RequeueAfter: recon.ErrorRefreshInterval}, err
				}
				break
			}
		}
	}

	managedMeta := &metav1.PartialObjectMetadataList{TypeMeta: metav1.TypeMeta{Kind: "ManagedClusterList",
		APIVersion: "cluster.open-cluster-management.io/v1"}}

	if err := r.Client.List(ctx, managedMeta); err != nil {
		return ctrl.Result{RequeueAfter: recon.WarningRefreshInterval},
			errors.Wrapf(err, "error listing managed clusters")
	}

	if err := r.updateManagedLabels(ctx, managedMeta, discoveredClusters); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager ...
func (r *ManagedClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	c, err := ctrl.NewControllerManagedBy(mgr).
		For(&clusterapiv1.ManagedCluster{}, builder.OnlyMetadata).
		Build(r)
	if err != nil {
		return errors.Wrapf(err, "error creating controller")
	}

	if err := c.Watch(source.Channel(r.Trigger, &handler.EnqueueRequestForObject{})); err != nil {
		return errors.Wrapf(err, "failed adding a watch channel")
	}

	return nil
}

/*
updateManagedLabels adds managed labels to discovered clusters that need them and removes the labels if the
discoveredclusters have the label but should not, based on the list of managedclusters.
*/
func (r *ManagedClusterReconciler) updateManagedLabels(ctx context.Context,
	managedClusters *metav1.PartialObjectMetadataList, discoveredClusters *discovery.DiscoveredClusterList) error {
	isManaged := map[string]bool{}
	for _, m := range managedClusters.Items {
		if id := getClusterID(m); id != "" {
			isManaged[id] = true
		}
	}

	for _, dc := range discoveredClusters.Items {
		dc := dc
		dcID := getDiscoveredID(dc)

		if isManaged[dcID] {
			if updateRequired := setManagedStatus(&dc); updateRequired {
				// Update with managed labels
				if err := r.Update(ctx, &dc); err != nil {
					return errors.Wrapf(err, "error setting managed status `%s`", dc.Name)
				}
				logf.Info("Updated cluster, adding managed status", "discoveredcluster", dc.Name,
					"discoveredcluster namespace", dc.Namespace)
			}

		} else {
			if updateRequired := unsetManagedStatus(&dc); updateRequired {
				// Update with managed labels removed
				if err := r.Update(ctx, &dc); err != nil {
					return errors.Wrapf(err, "error unsetting managed status `%s`", dc.Name)
				}
				logf.Info("Updated cluster, removing managed status", "discoveredcluster", dc.Name,
					"discoveredcluster namespace", dc.Namespace)
			}
		}
	}

	return nil
}

// getClusterID returns the clusterID from a managedCluster
func getClusterID(managedCluster metav1.PartialObjectMetadata) string {
	if labels := managedCluster.GetLabels(); labels != nil {
		return labels["clusterID"]
	}
	return ""
}

// getDiscoveredID returns the clusterID from a discoveredCluster
func getDiscoveredID(dc discovery.DiscoveredCluster) string {
	return dc.Spec.Name
}

// setManagedStatus returns true if labels were added and false if the labels already exist
func setManagedStatus(dc *discovery.DiscoveredCluster) bool {
	updated := false

	if dc.Labels == nil || dc.Labels["isManagedCluster"] != "true" {
		labels := make(map[string]string)
		if dc.Labels != nil {
			labels = dc.Labels
		}
		labels["isManagedCluster"] = "true"
		dc.SetLabels(labels)
		updated = true
	}

	if !dc.Spec.IsManagedCluster {
		dc.Spec.IsManagedCluster = true
		updated = true
	}

	return updated
}

// unsetManagedStatus returns true if labels were removed and false if the labels aren't present
func unsetManagedStatus(dc *discovery.DiscoveredCluster) bool {
	updated := false
	if dc.Labels["isManagedCluster"] == "true" {
		delete(dc.Labels, "isManagedCluster")
		updated = true
	}

	if dc.Spec.IsManagedCluster {
		dc.Spec.IsManagedCluster = false
		updated = true
	}
	return updated
}

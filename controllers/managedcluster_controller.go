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
	"time"

	"github.com/go-logr/logr"
	discoveryv1 "github.com/open-cluster-management/discovery/api/v1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
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
	Name   string
	Scheme *runtime.Scheme
	Log    logr.Logger
}

// +kubebuilder:rbac:groups=cluster.open-cluster-management.io,resources=managedclusters,verbs=get;list;watch

func (r *ManagedClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logr.FromContext(ctx)

	managedClusters := &unstructured.UnstructuredList{}
	managedClusters.SetGroupVersionKind(managedClusterGVK)
	if err := r.List(ctx, managedClusters); err != nil {
		return ctrl.Result{}, errors.Wrapf(err, "error listing managed clusters")
	}

	discoveredClusters := &discoveryv1.DiscoveredClusterList{}
	if err := r.List(ctx, discoveredClusters, client.InNamespace(req.Namespace)); err != nil {
		return ctrl.Result{}, errors.Wrapf(err, "error listing discovered clusters")
	}

	// Update for recently managed clusters
	for _, m := range managedClusters.Items {
		id := getClusterID(m)
		dc := matchingDiscoveredCluster(discoveredClusters, id)
		if dc == nil {
			// No matching discovered cluster
			log.Info("No matching discovered cluster for managed cluster", "managedCluster id", id)
			continue
		}

		if updateRequired := setManagedStatus(dc); updateRequired {
			// Update with managed labels
			if err := r.Update(ctx, dc); err != nil {
				return ctrl.Result{}, errors.Wrapf(err, "error updating discovered cluster `%s`", id)
			}
		}
	}

	// Update for recently unmanaged clusters
	for _, dc := range discoveredClusters.Items {
		if !dc.Spec.IsManagedCluster {
			continue
		}
		if isManagedCluster(dc, managedClusters) {
			continue
		}

		// Discovered cluster is labeled as managed, but does not have a matching managed cluster
		unsetManagedStatus(&dc)

		// Update with managed labels removed
		if err := r.Update(ctx, &dc); err != nil {
			return ctrl.Result{}, errors.Wrapf(err, "error updating discovered cluster `%s`", dc.Name)
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager ...
func (r *ManagedClusterReconciler) SetupWithManager(mgr ctrl.Manager) (controller.Controller, error) {
	managedClusterController, err := controller.NewUnmanaged(r.Name, mgr, controller.Options{
		Reconciler: r,
		Log:        r.Log,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "error creating controller")
	}

	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(managedClusterGVK)
	// Watch for Pod create / update / delete events and call Reconcile
	err = managedClusterController.Watch(
		&source.Kind{Type: u},
		&handler.EnqueueRequestForObject{},
		predicate.LabelChangedPredicate{})
	if err != nil {
		return nil, errors.Wrapf(err, "error watching managedclusters")
	}

	return managedClusterController, nil
}

func StartManagedClusterController(c controller.Controller, mgr ctrl.Manager, log logr.Logger) {
	// Start our controller in a goroutine so that we do not block.
	go func() {
		// Block until our controller manager is elected leader. We presume our
		// entire process will terminate if we lose leadership, so we don't need
		// to handle that.
		<-mgr.Elected()

		for {
			_, err := mgr.GetRESTMapper().RESTMapping(managedClusterGVK.GroupKind(), managedClusterGVK.Version)

			if err != nil {
				// Do not create controller
				log.Info("ManagedCluster resource does not exist: Waiting to start controller")
				time.Sleep(10 * time.Second)
				continue
			}

			// Start our controller. This will block until the stop channel is
			// closed, or the controller returns an error.
			if err := c.Start(context.TODO()); err != nil {
				log.Error(err, "cannot run ManagedCluster controller")
			}
		}
	}()
}

func getClusterID(managedCluster unstructured.Unstructured) string {
	if labels := managedCluster.GetLabels(); labels != nil {
		return labels["clusterID"]
	}
	return ""
}

// matchingDiscoveredCluster returns the discoveredCluster with the provided id or nil if not found
func matchingDiscoveredCluster(discoveredList *discoveryv1.DiscoveredClusterList, id string) *discoveryv1.DiscoveredCluster {
	for i, _ := range discoveredList.Items {
		if discoveredList.Items[i].Spec.Name == id {
			return &discoveredList.Items[i]
		}
	}
	return nil
}

// setManagedStatus returns true if labels were added and false if the labels already exist
func setManagedStatus(dc *discoveryv1.DiscoveredCluster) bool {
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
func unsetManagedStatus(dc *discoveryv1.DiscoveredCluster) bool {
	updated := false
	if dc.Labels["isManagedCluster"] == "true" {
		delete(dc.Labels, "isManagedCluster")
		updated = true
	}
	if dc.Spec.IsManagedCluster == true {
		dc.Spec.IsManagedCluster = false
		updated = true
	}
	return updated
}

func isManagedCluster(dc discoveryv1.DiscoveredCluster, managedClusters *unstructured.UnstructuredList) bool {
	discoveredName := dc.Spec.Name
	for _, mc := range managedClusters.Items {
		id := getClusterID(mc)
		if id != "" && id == discoveredName {
			return true
		}
	}
	return false
}

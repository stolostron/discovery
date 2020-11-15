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

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	discoveryv1 "github.com/open-cluster-management/discovery/api/v1"
)

// DiscoveredClusterRefreshReconciler reconciles a DiscoveredClusterRefresh object
type DiscoveredClusterRefreshReconciler struct {
	client.Client
	Log     logr.Logger
	Scheme  *runtime.Scheme
	Trigger chan event.GenericEvent
}

// +kubebuilder:rbac:groups=discovery.open-cluster-management.io,resources=discoveredclusterrefreshes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=discovery.open-cluster-management.io,resources=discoveredclusterrefreshes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=discovery.open-cluster-management.io,resources=discoveredclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=discovery.open-cluster-management.io,resources=discoveredclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=discovery.open-cluster-management.io,resources=discoveryconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=discovery.open-cluster-management.io,resources=discoveryconfigs/status,verbs=get;update;patch

// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete

func (r *DiscoveredClusterRefreshReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("discoveredclusterrefresh", req.NamespacedName)

	refresh := &discoveryv1.DiscoveredClusterRefresh{}
	if err := r.Get(ctx, types.NamespacedName{
		Name:      req.Name,
		Namespace: req.Namespace,
	}, refresh); err != nil {
		log.Error(err, "unable to find discoveredclusterrefresh")
		return ctrl.Result{}, err
	}

	// Get discovery config in the trigger's namespace
	config := &discoveryv1.DiscoveryConfig{}
	if err := r.Get(ctx, types.NamespacedName{
		Name:      "discoveryconfig",
		Namespace: req.Namespace,
	}, config); err != nil {
		log.Error(err, "unable to find discoveryconfig")
		return ctrl.Result{}, err
	}

	// Trigger reconcile of found DiscoveryConfig
	r.Trigger <- event.GenericEvent{
		Meta:   config.GetObjectMeta(),
		Object: config,
	}

	// Trigger is done. The refresh can now be deleted.
	if err := r.Delete(ctx, refresh); err != nil {
		log.Error(err, "failed to delete refresh object")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager ...
func (r *DiscoveredClusterRefreshReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&discoveryv1.DiscoveredClusterRefresh{}).
		// Watches(&source.Kind{Type: &discoveryv1.DiscoveryConfig{}}, &handler.EnqueueRequestForObject{}).
		WithEventFilter(predicate.Funcs{
			// Skip delete events
			DeleteFunc: func(e event.DeleteEvent) bool {
				return false
			},
		}).
		Complete(r)
}

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
	"k8s.io/apimachinery/pkg/api/errors"
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
	Scheme  *runtime.Scheme
	Trigger chan event.GenericEvent
}

// +kubebuilder:rbac:groups=discovery.open-cluster-management.io,resources=discoveredclusterrefreshes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=discovery.open-cluster-management.io,resources=discoveredclusterrefreshes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=discovery.open-cluster-management.io,resources=discoveryconfigs,verbs=get;list;watch
// +kubebuilder:rbac:groups=discovery.open-cluster-management.io,resources=discoveryconfigs/status,verbs=get;update;patch

func (r *DiscoveredClusterRefreshReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logr.FromContext(ctx)

	refresh := &discoveryv1.DiscoveredClusterRefresh{}
	if err := r.Get(ctx, types.NamespacedName{
		Name:      req.Name,
		Namespace: req.Namespace,
	}, refresh); err != nil {
		if errors.IsNotFound(err) {
			log.Info("refresh not found")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Get discovery config in the same namespace
	config := &discoveryv1.DiscoveryConfig{}
	if err := r.Get(ctx, types.NamespacedName{
		Name:      "discoveryconfig",
		Namespace: req.Namespace,
	}, config); err != nil {
		if errors.IsNotFound(err) {
			log.Info("could not find discoveryconfig to refresh in same namespace")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{RequeueAfter: 5 * time.Second}, err
	}

	// Trigger reconcile of found DiscoveryConfig
	r.Trigger <- event.GenericEvent{
		Object: config,
	}

	// Trigger is done. The refresh object can now be deleted.
	if err := r.Delete(ctx, refresh); err != nil {
		return ctrl.Result{RequeueAfter: 1 * time.Second}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager ...
func (r *DiscoveredClusterRefreshReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&discoveryv1.DiscoveredClusterRefresh{}).
		WithEventFilter(predicate.Funcs{
			// Skip delete events
			DeleteFunc: func(e event.DeleteEvent) bool {
				return false
			},
		}).
		Complete(r)
}

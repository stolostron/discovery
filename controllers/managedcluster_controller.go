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
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
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
	Name   string
	Scheme *runtime.Scheme
	Log    logr.Logger
}

// +kubebuilder:rbac:groups=cluster.open-cluster-management.io,resources=managedclusters,verbs=get;list;watch

func (r *ManagedClusterReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("managedcluster", req.NamespacedName)

	managedClusterList := &unstructured.UnstructuredList{}
	managedClusterList.SetGroupVersionKind(managedClusterGVK)

	if err := r.List(ctx, managedClusterList); err != nil {
		return ctrl.Result{}, errors.Wrapf(err, "error listing managed clusters")
	}

	for _, m := range managedClusterList.Items {
		log.Info("Managed cluster", "Name", m.GetName())
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
	err = managedClusterController.Watch(&source.Kind{Type: u}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return nil, errors.Wrapf(err, "error watching managedclusters")
	}

	return managedClusterController, nil
}

func StartManagedClusterController(c controller.Controller, mgr ctrl.Manager, log logr.Logger) {
	// Create a stop channel for our controller. The controller will stop when
	// this channel is closed.
	stop := make(chan struct{})

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
			if err := c.Start(stop); err != nil {
				log.Error(err, "cannot run ManagedCluster controller")
			}
		}
	}()
}

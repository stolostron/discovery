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
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	discoveryv1 "github.com/open-cluster-management/discovery/api/v1"
	"github.com/open-cluster-management/discovery/pkg/auth"
	"github.com/open-cluster-management/discovery/pkg/ocm"
	corev1 "k8s.io/api/core/v1"
)

var (
	refreshInterval = 10 * time.Minute
)

// DiscoveredClusterRefreshReconciler reconciles a DiscoveredClusterRefresh object
type DiscoveredClusterRefreshReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
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

	// Get discovery config. Die if there is none
	config := &discoveryv1.DiscoveryConfig{}
	if err := r.Get(ctx, types.NamespacedName{
		Name:      "discoveryconfig",
		Namespace: req.Namespace,
	}, config); err != nil {
		log.Error(err, "unable to fetch DiscoveryConfig")
		return ctrl.Result{RequeueAfter: refreshInterval}, client.IgnoreNotFound(err)
	}

	// Get all refresh requests. If there are none then we were triggered by refresh interval
	var refreshRequestList discoveryv1.DiscoveredClusterRefreshList
	err := r.List(ctx, &refreshRequestList, client.InNamespace(req.Namespace))
	if err != nil {
		return ctrl.Result{}, err
	}

	secretName := config.Spec.ProviderConnections[0]
	userToken, err := userToken(r.Client, types.NamespacedName{Name: secretName, Namespace: req.Namespace})
	if err != nil {
		return ctrl.Result{}, err
	}

	accessToken, err := auth.NewAccessToken(userToken)
	if err != nil {
		return ctrl.Result{}, err
	}

	// List all already-discovered clusters
	var discoveredList discoveryv1.DiscoveredClusterList
	if err := r.List(ctx, &discoveredList, client.InNamespace(req.Namespace)); err != nil {
		log.Error(err, "unable to list discovered clusters")
		return ctrl.Result{}, err
	}

	existing := make(map[string]int, len(discoveredList.Items))
	for i, cluster := range discoveredList.Items {
		existing[cluster.Name] = i
	}

	var createClusters []discoveryv1.DiscoveredCluster
	var updateClusters []discoveryv1.DiscoveredCluster
	var deleteClusters []discoveryv1.DiscoveredCluster
	var unchangedClusters []discoveryv1.DiscoveredCluster

	clusterRequest := ocm.ClusterRequest(config)
	page := 1
	size := 1000
	for {
		log.Info("Requesting OCM clusters", "page", page, "size", size)
		newDiscoveredList, err := clusterRequest.Page(page).Size(size).Token(accessToken).Filter(config.Spec.Filters).Get(ctx)
		if err != nil {
			return ctrl.Result{}, err
		}

		// Merge newly discovered clusters with existing list
		for _, cluster := range newDiscoveredList.Items {
			discoveredCluster := ocm.DiscoveredCluster(cluster)
			discoveredCluster.SetNamespace(req.Namespace)

			ind, exists := existing[discoveredCluster.Name]
			if !exists {
				// Newly discovered cluster
				createClusters = append(createClusters, discoveredCluster)
				delete(existing, discoveredCluster.Name)
				continue
			}
			// Cluster has already been discovered. Check for changes.
			if same(discoveredCluster, discoveredList.Items[ind]) {
				unchangedClusters = append(unchangedClusters, discoveredCluster)
				delete(existing, discoveredCluster.Name)
			} else {
				updated := discoveredList.Items[ind]
				updated.Info = discoveredCluster.Info
				updateClusters = append(updateClusters, updated)
				delete(existing, discoveredCluster.Name)
			}
		}

		if len(newDiscoveredList.Items) < size {
			break
		}
		page++
	}

	// Remaining clusters are no longer found by OCM and should be labeled for delete
	for _, ind := range existing {
		deleteClusters = append(deleteClusters, discoveredList.Items[ind])
	}

	// Create new clusters and clean up old ones
	for _, cluster := range createClusters {
		if err := r.Create(ctx, &cluster); err != nil {
			log.Error(err, "unable to create discovered cluster", "name", cluster.Name)
			return ctrl.Result{}, err
		}
		log.Info("Created cluster", "Name", cluster.Name)
	}
	for _, cluster := range updateClusters {
		if err := r.Update(ctx, &cluster); err != nil {
			log.Error(err, "unable to update discovered cluster", "name", cluster.Name)
			return ctrl.Result{}, err
		}
		log.Info("Updated cluster", "Name", cluster.Name)
	}
	for _, cluster := range deleteClusters {
		if err := r.Delete(ctx, &cluster); err != nil {
			log.Error(err, "unable to delete discovered cluster", "name", cluster.Name)
			return ctrl.Result{}, err
		}
		log.Info("Deleted cluster", "Name", cluster.Name)
	}

	log.Info("Cluster categories", "Created", len(createClusters), "Updated", len(updateClusters), "Deleted", len(deleteClusters), "Unchanged", len(unchangedClusters))

	// Remove refresh requests that existed at the beginning of this reconciliation
	for _, request := range refreshRequestList.Items {
		if err := r.Delete(ctx, &request); err != nil {
			log.Error(err, "unable to delete refreshRequest", "Name", request.Name)
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{RequeueAfter: refreshInterval}, nil
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

func same(c1, c2 discoveryv1.DiscoveredCluster) bool {
	c1i, c2i := c1.Info, c2.Info
	if c1i.APIURL != c2i.APIURL {
		return false
	}
	if c1i.CloudProvider != c2i.CloudProvider {
		return false
	}
	if c1i.Console != c2i.Console {
		return false
	}
	if c1i.HealthState != c2i.HealthState {
		return false
	}
	if c1i.Name != c2i.Name {
		return false
	}
	if c1i.OpenshiftVersion != c2i.OpenshiftVersion {
		return false
	}
	if c1i.Product != c2i.Product {
		return false
	}
	if c1i.Region != c2i.Region {
		return false
	}
	if c1i.State != c2i.State {
		return false
	}
	return true
}

// readOCMAPISecret reads the token from the OCM api secret
func userToken(kubeClient client.Client, secret types.NamespacedName) (string, error) {
	ocmSecret := &corev1.Secret{}
	err := kubeClient.Get(context.TODO(), secret, ocmSecret)
	if err != nil {
		return "", err
	}

	if _, ok := ocmSecret.Data["token"]; !ok {
		return "", fmt.Errorf("Secret '%s' does not contain 'token' field", secret.Name)
	}
	return string(ocmSecret.Data["token"]), nil
}

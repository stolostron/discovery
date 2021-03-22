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

	"github.com/circonus-labs/circonus-gometrics/api/config"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"sigs.k8s.io/yaml"

	discoveryv1 "github.com/open-cluster-management/discovery/api/v1"
	"github.com/open-cluster-management/discovery/pkg/ocm"
	"github.com/open-cluster-management/discovery/util/reconciler"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// DiscoveryConfigReconciler reconciles a DiscoveryConfig object
type DiscoveryConfigReconciler struct {
	client.Client
	Scheme  *runtime.Scheme
	Trigger chan event.GenericEvent
}

// CloudRedHatProviderConnection ...
type CloudRedHatProviderConnection struct {
	OCMApiToken string `yaml:"ocmAPIToken"`
}

// +kubebuilder:rbac:groups=discovery.open-cluster-management.io,resources=discoveredclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=discovery.open-cluster-management.io,resources=discoveredclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=discovery.open-cluster-management.io,resources=discoveredclusters/finalizers,verbs=get;update;patch
// +kubebuilder:rbac:groups=discovery.open-cluster-management.io,resources=discoveryconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=discovery.open-cluster-management.io,resources=discoveryconfigs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=discovery.open-cluster-management.io,resources=discoveryconfigs/finalizers,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch

func (r *DiscoveryConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logr.FromContext(ctx)

	config := &discoveryv1.DiscoveryConfig{}
	err := r.Get(ctx, req.NamespacedName, config)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if len(config.Spec.ProviderConnections) == 0 {
		log.Info("No provider connections in config. Returning.")
		return ctrl.Result{}, nil
	}

	if err = r.updateDiscoveredClusters(config); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager ...
func (r *DiscoveryConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	c, err := ctrl.NewControllerManagedBy(mgr).
		For(&discoveryv1.DiscoveryConfig{}).
		// Watches(&source.Kind{Type: &discoveryv1.DiscoveryConfig{}}, &handler.EnqueueRequestForObject{}).
		WithEventFilter(predicate.Funcs{
			// Skip delete events
			DeleteFunc: func(e event.DeleteEvent) bool {
				return false
			},
		}).
		Build(r)
	if err != nil {
		return errors.Wrapf(err, "error creating controller")
	}

	if err := c.Watch(
		&source.Channel{Source: r.Trigger},
		&handler.EnqueueRequestForObject{},
	); err != nil {
		return errors.Wrapf(err, "failed adding a watch channel")
	}

	return nil
}

func (r *DiscoveryConfigReconciler) updateDiscoveredClusters(config *discoveryv1.DiscoveryConfig) error {
	ctx := context.Background()
	var allClusters map[string]discoveryv1.DiscoveredCluster

	// Gather clusters from all provider connections
	for _, secret := range config.Spec.ProviderConnections {
		userToken, err := r.getUserToken(types.NamespacedName{secret, config.Namespace})

		var baseURL string
		if annotations := config.GetAnnotations(); annotations != nil {
			baseURL = annotations["ocmBaseURL"]
		}
		filters := config.Spec.Filters
		discovered, err := ocm.DiscoverClusters(userToken, baseURL, filters)
		if err != nil {
			return err
		}

		for _, dc := range discovered {
			dc.SetNamespace(config.Namespace)
			id := dc.ClusterName
			allClusters[id] = dc
		}
	}

	// Assign managed status
	managed, err := r.getManagedClusters()
	if err != nil {
		return err
	}
	assignManagedStatus(allClusters, managed)

	// Get map to check if discovered is already created
	existing, err := r.getExistingClusterMap()
	if err != nil {
		return err
	}

	for _, discoveredCluster := range allClusters {
		applyCluster(discoveredCluster, existing)
		delete(existing, discoveredCluster.Spec.Name)
	}

	var createClusters []discoveryv1.DiscoveredCluster
	var updateClusters []discoveryv1.DiscoveredCluster
	var deleteClusters []discoveryv1.DiscoveredCluster
	var unchangedClusters []discoveryv1.DiscoveredCluster

	for _, cluster := range newClusters {
		matchingSubscriptionSpec, ok := subscriptionSpecs[cluster.ExternalID]
		if !ok {
			// Ignore clusters without an active subscription
			continue
		}

		// Build a DiscoveredCluster object from the cluster information
		discoveredCluster := discoveredCluster(cluster)
		discoveredCluster.SetNamespace(req.Namespace)

		// Assign dummy status
		discoveredCluster.Spec.Subscription = matchingSubscriptionSpec

		// Assign managed status
		if _, managed := managedClusterIDs[discoveredCluster.Spec.Name]; managed {
			setManagedStatus(&discoveredCluster)
		}

		// // Add reference to secret used for authentication
		// discoveredCluster.Spec.ProviderConnections = nil
		// secretRef, err := ref.GetReference(r.Scheme, ocmSecret)
		// if err != nil {
		// 	log.Error(err, "unable to make reference to secret", "secret", secretRef)
		// }
		// discoveredCluster.Spec.ProviderConnections = append(discoveredCluster.Spec.ProviderConnections, *secretRef)

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
			updated.Spec = discoveredCluster.Spec
			updateClusters = append(updateClusters, updated)
			delete(existing, discoveredCluster.Name)
		}
	}

	// Remaining clusters are no longer found by OCM and should be labeled for delete
	for _, ind := range existing {
		deleteClusters = append(deleteClusters, discoveredList.Items[ind])
	}

	// Create new clusters and clean up old ones
	for _, cluster := range createClusters {
		cluster := cluster
		if err := ctrl.SetControllerReference(config, &cluster, r.Scheme); err != nil {
			log.Error(err, "failed to set controller reference", "name", cluster.Name)
			return ctrl.Result{}, err
		}
		if err := r.Create(ctx, &cluster); err != nil {
			log.Error(err, "unable to create discovered cluster", "name", cluster.Name)
			return ctrl.Result{}, err
		}
		log.Info("Created cluster", "Name", cluster.Name)
	}
	for _, cluster := range updateClusters {
		cluster := cluster
		if err := r.Update(ctx, &cluster); err != nil {
			log.Error(err, "unable to update discovered cluster", "name", cluster.Name)
			return ctrl.Result{}, err
		}
		log.Info("Updated cluster", "Name", cluster.Name)
	}
	for _, cluster := range deleteClusters {
		cluster := cluster
		if err := r.Delete(ctx, &cluster); err != nil {
			log.Error(err, "unable to delete discovered cluster", "name", cluster.Name)
			return ctrl.Result{}, err
		}
		log.Info("Deleted cluster", "Name", cluster.Name)
	}

	log.Info("Cluster categories", "Created", len(createClusters), "Updated", len(updateClusters), "Deleted", len(deleteClusters), "Unchanged", len(unchangedClusters))

	return ctrl.Result{RequeueAfter: reconciler.RefreshInterval}, nil
}

// getUserToken takes a secret cotaining a Provider Connection. It fetches the secret,
// parses it, and returns the stored OCM api token.
func (r *DiscoveryConfigReconciler) getUserToken(secretName types.NamespacedName) (string, error) {
	ocmSecret := &corev1.Secret{}
	if err := r.Get(context.TODO(), secretName, ocmSecret); err != nil {
		return "", err
	}

	if _, ok := ocmSecret.Data["metadata"]; !ok {
		return "", fmt.Errorf("Secret '%' does not contain 'metadata' field", secretName)
	}

	providerConnection := &CloudRedHatProviderConnection{}
	if err := yaml.Unmarshal(ocmSecret.Data["metadata"], providerConnection); err != nil {
		return "", err
	}

	return providerConnection.OCMApiToken, nil
}

// assignManagedStatus marks clusters in the discovered map as managed if they are in the managed list
func assignManagedStatus(discovered map[string]discoveryv1.DiscoveredCluster, managed []unstructured.Unstructured) {
	for _, mc := range managed {
		id := getClusterID(mc)
		if id != "" {
			// Update cluster as managed
			dc := discovered[id]
			setManagedStatus(&dc)
			discovered[id] = dc
		}
	}
}

func (r *DiscoveryConfigReconciler) getManagedClusters() ([]unstructured.Unstructured, error) {
	ctx := context.Background()

	// List all existing managed clusters
	managedList := &unstructured.UnstructuredList{}
	managedList.SetGroupVersionKind(managedClusterGVK)
	if err := r.List(ctx, managedList); err != nil {
		// Capture case were ManagedClusters resource does not exist
		if !apimeta.IsNoMatchError(err) {
			return nil, errors.Wrapf(err, "error listing managed clusters")
		}
	}
	return managedList.Items, nil
}

func (r *DiscoveryConfigReconciler) getExistingClusterMap() (map[string]discoveryv1.DiscoveredCluster, error) {
	ctx := context.Background()

	// List all existing discovered clusters
	var discoveredList discoveryv1.DiscoveredClusterList
	if err := r.List(ctx, &discoveredList, client.InNamespace(config.Namespace)); err != nil {
		return nil, errors.Wrapf(err, "error listing list discovered clusters")
	}
	existingDCs := make(map[string]discoveryv1.DiscoveredCluster, len(discoveredList.Items))
	for _, dc := range discoveredList.Items {
		existingDCs[dc.Spec.Name] = dc
	}
	return existingDCs, nil
}

// applyCluster creates the DiscoveredCluster resources or updates it if necessary. If the cluster already
// exists and doesn't need updating then nothing changes.
func applyCluster(ctx context.Context, config *discoveryv1.DiscoveryConfig, dc discoveryv1.DiscoveredCluster, existing map[string]discoveryv1.DiscoveredCluster) error {
	current, exists := existing[dc.Spec.Name]
	if !exists {
		// Newly discovered cluster
		return createCluster(ctx, dc)
	}

	if same(dc, current) {
		// Discovered cluster has not changed
		return nil
	}

	// Cluster needs to be updated
	return updateCluster(ctx, dc)
}

func createCluster(ctx context.Context, config *discoveryv1.DiscoveryConfig, dc discoveryv1.DiscoveredCluster) error {
	if err := ctrl.SetControllerReference(config, &cluster, r.Scheme); err != nil {
		log.Error(err, "failed to set controller reference", "name", cluster.Name)
		return ctrl.Result{}, err
	}
	if err := r.Create(ctx, &cluster); err != nil {
		log.Error(err, "unable to create discovered cluster", "name", cluster.Name)
		return ctrl.Result{}, err
	}
}

func same(c1, c2 discoveryv1.DiscoveredCluster) bool {
	c1i, c2i := c1.Spec, c2.Spec
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
	if c1i.IsManagedCluster != c2i.IsManagedCluster {
		return false
	}
	return true
}

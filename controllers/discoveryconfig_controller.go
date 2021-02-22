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
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ref "k8s.io/client-go/tools/reference"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"sigs.k8s.io/yaml"

	discoveryv1 "github.com/open-cluster-management/discovery/api/v1"
	"github.com/open-cluster-management/discovery/pkg/api/domain/auth_domain"
	"github.com/open-cluster-management/discovery/pkg/api/domain/cluster_domain"
	"github.com/open-cluster-management/discovery/pkg/api/services/auth_service"
	"github.com/open-cluster-management/discovery/pkg/api/services/cluster_service"
	"github.com/open-cluster-management/discovery/util/reconciler"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	// Get discovery config. Die if there is none
	config := &discoveryv1.DiscoveryConfig{}
	if err := r.Get(ctx, types.NamespacedName{
		Name:      req.Name,
		Namespace: req.Namespace,
	}, config); err != nil {
		log.Error(err, "unable to fetch DiscoveryConfig")
		return ctrl.Result{}, err
	}

	// Update the DiscoveryConfig status
	// config.Status.LastUpdateTime = &metav1.Time{Time: time.Now()}
	// if err := r.Status().Update(ctx, config); err != nil {
	// 	log.Error(err, "unable to update discoveryconfig status")
	// 	return ctrl.Result{}, err
	// }

	// Get user token from secret provided in config
	if len(config.Spec.ProviderConnections) == 0 {
		log.Info("No provider connections in config. Returning.")
		return ctrl.Result{}, nil
	}
	secretName := config.Spec.ProviderConnections[0]
	ocmSecret := &corev1.Secret{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: secretName, Namespace: req.Namespace}, ocmSecret)
	if err != nil {
		return ctrl.Result{}, err
	}
	if _, ok := ocmSecret.Data["metadata"]; !ok {
		return ctrl.Result{}, fmt.Errorf("Secret '%s' does not contain 'metadata' field", secretName)
	}

	providerConnection := &CloudRedHatProviderConnection{}
	err = yaml.Unmarshal(ocmSecret.Data["metadata"], providerConnection)
	if err != nil {
		return ctrl.Result{}, err
	}
	userToken := providerConnection.OCMApiToken

	// Request ephemeral access token with user token. This will be used for OCM requests
	authRequest := auth_domain.AuthRequest{
		Token: userToken,
	}
	if annotations := config.GetAnnotations(); annotations != nil {
		authRequest.BaseURL = annotations["ocmBaseURL"]
	}
	accessToken, err := auth_service.AuthClient.GetToken(authRequest)
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

	// List all managed clusters
	managedClusters := &unstructured.UnstructuredList{}
	managedClusters.SetGroupVersionKind(managedClusterGVK)
	if err := r.List(ctx, managedClusters); err != nil {
		// Capture case were ManagedClusters resource does not exist
		if !apimeta.IsNoMatchError(err) {
			return ctrl.Result{}, errors.Wrapf(err, "error listing managed clusters")
		}
	}

	managedClusterIDs := make(map[string]int, len(managedClusters.Items))
	for i, mc := range managedClusters.Items {
		name := getClusterID(mc)
		if name != "" {
			managedClusterIDs[getClusterID(mc)] = i
		}
	}

	var createClusters []discoveryv1.DiscoveredCluster
	var updateClusters []discoveryv1.DiscoveredCluster
	var deleteClusters []discoveryv1.DiscoveredCluster
	var unchangedClusters []discoveryv1.DiscoveredCluster

	requestConfig := cluster_domain.ClusterRequest{
		Token:  accessToken,
		Filter: config.Spec.Filters,
	}
	if annotations := config.GetAnnotations(); annotations != nil {
		requestConfig.BaseURL = annotations["ocmBaseURL"]
	}
	clusterClient := cluster_service.ClusterClientGenerator.NewClient(requestConfig)

	newClusters, err := clusterClient.GetClusters()
	if err != nil {
		return ctrl.Result{}, err
	}
	for _, cluster := range newClusters {
		// Build a DiscoveredCluster object from the cluster information
		discoveredCluster := discoveredCluster(cluster)
		discoveredCluster.SetNamespace(req.Namespace)

		// Assign dummy status
		discoveredCluster.Spec.Subscription = discoveryv1.SubscriptionSpec{
			Status:       "Active",
			SupportLevel: "None",
			Managed:      false,
			CreatorID:    "abc123",
		}

		// Assign managed status
		if _, managed := managedClusterIDs[discoveredCluster.Spec.Name]; managed {
			setManagedStatus(&discoveredCluster)
		}

		// Add reference to secret used for authentication
		discoveredCluster.Spec.ProviderConnections = nil
		secretRef, err := ref.GetReference(r.Scheme, ocmSecret)
		if err != nil {
			log.Error(err, "unable to make reference to secret", "secret", secretRef)
		}
		discoveredCluster.Spec.ProviderConnections = append(discoveredCluster.Spec.ProviderConnections, *secretRef)

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
		ctrl.SetControllerReference(config, &cluster, r.Scheme)
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

	return ctrl.Result{RequeueAfter: reconciler.RefreshInterval}, nil
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
			UpdateFunc: func(e event.UpdateEvent) bool {
				return e.ObjectNew.GetGeneration() != e.ObjectOld.GetGeneration()
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

// DiscoveredCluster ...
func discoveredCluster(cluster cluster_domain.Cluster) discoveryv1.DiscoveredCluster {
	return discoveryv1.DiscoveredCluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "operator.open-cluster-management.io/v1",
			Kind:       "DiscoveredCluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.ID,
			Namespace: "open-cluster-management",
		},
		Spec: discoveryv1.DiscoveredClusterSpec{
			Console:           cluster.Console.URL,
			CreationTimestamp: cluster.CreationTimestamp,
			ActivityTimestamp: cluster.ActivityTimestamp,
			// ActivityTimestamp: metav1.NewTime(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)),
			OpenshiftVersion: cluster.OpenShiftVersion,
			Name:             cluster.Name,
			Region:           cluster.Region.ID,
			CloudProvider:    cluster.CloudProvider.ID,
			HealthState:      cluster.HealthState,
			State:            cluster.State,
			Product:          cluster.Product.ID,
			// IsManagedCluster:  managedClusterNames[cluster.Name],
			// APIURL: apiurl,
		},
	}
}

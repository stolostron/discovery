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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
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

	discovery "github.com/open-cluster-management/discovery/api/v1alpha1"
	"github.com/open-cluster-management/discovery/pkg/ocm"
	"github.com/open-cluster-management/discovery/util/reconciler"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	defaultDiscoveryConfigName = "discovery"
)

var (
	// baseURLAnnotation is the annotation set in a DiscoveryConfig that overrides the URL base used to find clusters
	baseURLAnnotation = "ocmBaseURL"
)

var ErrBadFormat = errors.New("bad format")

// DiscoveryConfigReconciler reconciles a DiscoveryConfig object
type DiscoveryConfigReconciler struct {
	client.Client
	Scheme  *runtime.Scheme
	Trigger chan event.GenericEvent
}

// CloudRedHatProviderConnection ...
type CloudRedHatCredential struct {
	OCMApiToken string `yaml:"ocmAPIToken"`
}

// +kubebuilder:rbac:groups=discovery.open-cluster-management.io,resources=discoveredclusters,verbs=get;list;watch;create;update;patch;delete;deletecollection
// +kubebuilder:rbac:groups=discovery.open-cluster-management.io,resources=discoveredclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=discovery.open-cluster-management.io,resources=discoveredclusters/finalizers,verbs=get;update;patch
// +kubebuilder:rbac:groups=discovery.open-cluster-management.io,resources=discoveryconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=discovery.open-cluster-management.io,resources=discoveryconfigs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=discovery.open-cluster-management.io,resources=discoveryconfigs/finalizers,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch

func (r *DiscoveryConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logr.FromContext(ctx)
	log.Info("Reconciling DiscoveryConfig")

	if req.Name != defaultDiscoveryConfigName {
		err := fmt.Errorf("DiscoveryConfig resource name must be '%s'", defaultDiscoveryConfigName)
		log.Error(err, "Invalid DiscoveryConfig resource name")
		return ctrl.Result{}, nil
	}

	config := &discovery.DiscoveryConfig{}
	err := r.Get(ctx, req.NamespacedName, config)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if err = r.updateDiscoveredClusters(ctx, config); err != nil {
		log.Error(err, "Error updating DiscoveredClusters")
		return ctrl.Result{}, err
	}

	log.Info("Reconcile complete. Rescheduling.", "time", reconciler.RefreshInterval)
	return ctrl.Result{RequeueAfter: reconciler.RefreshInterval}, nil
}

// SetupWithManager ...
func (r *DiscoveryConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	c, err := ctrl.NewControllerManagedBy(mgr).
		For(&discovery.DiscoveryConfig{}).
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

func (r *DiscoveryConfigReconciler) updateDiscoveredClusters(ctx context.Context, config *discovery.DiscoveryConfig) error {
	allClusters := map[string]discovery.DiscoveredCluster{}
	log := logr.FromContext(ctx)

	// Parse user token from providerconnection secret
	ocmSecret := &corev1.Secret{}
	if err := r.Get(context.TODO(), types.NamespacedName{Name: config.Spec.Credential, Namespace: config.Namespace}, ocmSecret); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Secret does not exist")
			return r.deleteAllClusters(ctx, config)
		}
		return err
	}
	userToken, err := parseUserToken(ocmSecret)
	if err != nil {
		log.Error(err, "Error parsing token from secret")
		return r.deleteAllClusters(ctx, config)
	}

	baseURL := getURLOverride(config)
	filters := config.Spec.Filters
	discovered, err := ocm.DiscoverClusters(userToken, baseURL, filters)
	if err != nil {
		if ocm.IsUnrecoverable(err) {
			log.Info("Error is unrecoverable. Cleaning up clusters.")
			return r.deleteAllClusters(ctx, config)
		}
		return err
	}

	// Get reference to secret used for authentication
	secretRef, err := ref.GetReference(r.Scheme, ocmSecret)
	if err != nil {
		return errors.Wrapf(err, "unable to make reference to secret %s", secretRef)
	}

	for _, dc := range discovered {
		dc.SetNamespace(config.Namespace)
		dc.Spec.ProviderConnections = append(dc.Spec.ProviderConnections, *secretRef)
		dc.Spec.Credential = *secretRef
		merge(allClusters, dc)
	}

	// Assign managed status
	managed, err := r.getManagedClusters()
	if err != nil {
		return err
	}
	if managed != nil && len(managed) > 0 {
		assignManagedStatus(allClusters, managed)
	} else {
		log.Info("No managed clusters available")
	}

	// Create map to check if cluster already discovered
	existing, err := r.getExistingClusterMap(ctx, config)
	if err != nil {
		return err
	}

	// Apply clusters discovered
	for _, discoveredCluster := range allClusters {
		err := r.applyCluster(ctx, config, discoveredCluster, existing)
		if err != nil {
			return err
		}
		delete(existing, discoveredCluster.Spec.Name)
	}

	// Everything remaining in existing should be deleted
	for _, c := range existing {
		err := r.deleteCluster(ctx, c)
		if err != nil {
			return err
		}
	}

	return nil
}

// getUserToken takes a secret cotaining credentials and returns the stored OCM api token.
func parseUserToken(secret *corev1.Secret) (string, error) {
	if _, ok := secret.Data["metadata"]; !ok {
		// return "", fmt.Errorf("Secret '%s' does not contain 'metadata' field", secret.Name)
		return "", fmt.Errorf("%s: %w", secret.Name, ErrBadFormat)
	}

	cred := &CloudRedHatCredential{}
	if err := yaml.Unmarshal(secret.Data["metadata"], cred); err != nil {
		return "", fmt.Errorf("%s: %w", secret.Name, ErrBadFormat)
	}

	return cred.OCMApiToken, nil
}

// assignManagedStatus marks clusters in the discovered map as managed if they are in the managed list
func assignManagedStatus(discovered map[string]discovery.DiscoveredCluster, managed []unstructured.Unstructured) {
	for _, mc := range managed {
		id := getClusterID(mc)
		if id != "" {
			// Update cluster as managed
			if dc, ok := discovered[id]; ok {
				setManagedStatus(&dc)
				discovered[id] = dc
			}
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
		if apimeta.IsNoMatchError(err) {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "error listing managed clusters")
	}
	return managedList.Items, nil
}

func (r *DiscoveryConfigReconciler) getExistingClusterMap(ctx context.Context, config *discovery.DiscoveryConfig) (map[string]discovery.DiscoveredCluster, error) {
	// List all existing discovered clusters
	var discoveredList discovery.DiscoveredClusterList
	if err := r.List(ctx, &discoveredList, client.InNamespace(config.Namespace)); err != nil {
		return nil, errors.Wrapf(err, "error listing list discovered clusters")
	}
	existingDCs := make(map[string]discovery.DiscoveredCluster, len(discoveredList.Items))
	for _, dc := range discoveredList.Items {
		existingDCs[dc.Spec.Name] = dc
	}
	return existingDCs, nil
}

// applyCluster creates the DiscoveredCluster resources or updates it if necessary. If the cluster already
// exists and doesn't need updating then nothing changes.
func (r *DiscoveryConfigReconciler) applyCluster(ctx context.Context, config *discovery.DiscoveryConfig, dc discovery.DiscoveredCluster, existing map[string]discovery.DiscoveredCluster) error {
	current, exists := existing[dc.Spec.Name]
	if !exists {
		// Newly discovered cluster
		return r.createCluster(ctx, config, dc)
	}

	if same(dc, current) {
		// Discovered cluster has not changed
		return nil
	}

	// Cluster needs to be updated
	return r.updateCluster(ctx, config, dc, current)
}

func (r *DiscoveryConfigReconciler) createCluster(ctx context.Context, config *discovery.DiscoveryConfig, dc discovery.DiscoveredCluster) error {
	log := logr.FromContext(ctx)
	if err := ctrl.SetControllerReference(config, &dc, r.Scheme); err != nil {
		return errors.Wrapf(err, "Error setting controller reference on DiscoveredCluster %s", dc.Name)
	}
	if err := r.Create(ctx, &dc); err != nil {
		return errors.Wrapf(err, "Error creating DiscoveredCluster %s", dc.Name)
	}
	log.Info("Created cluster", "Name", dc.Name)
	return nil
}

func (r *DiscoveryConfigReconciler) updateCluster(ctx context.Context, config *discovery.DiscoveryConfig, new, old discovery.DiscoveredCluster) error {
	log := logr.FromContext(ctx)
	updated := old
	updated.Spec = new.Spec
	if err := r.Update(ctx, &updated); err != nil {
		return errors.Wrapf(err, "Error updating DiscoveredCluster %s", updated.Name)
	}
	log.Info("Updated cluster", "Name", updated.Name)
	return nil
}

func (r *DiscoveryConfigReconciler) deleteCluster(ctx context.Context, dc discovery.DiscoveredCluster) error {
	log := logr.FromContext(ctx)
	if err := r.Delete(ctx, &dc); err != nil {
		return errors.Wrapf(err, "Error deleting DiscoveredCluster %s", dc.Name)
	}
	log.Info("Deleted cluster", "Name", dc.Name)
	return nil
}

func (r *DiscoveryConfigReconciler) deleteAllClusters(ctx context.Context, config *discovery.DiscoveryConfig) error {
	log := logr.FromContext(ctx)
	if err := r.DeleteAllOf(ctx, &discovery.DiscoveredCluster{}, client.InNamespace(config.Namespace)); err != nil {
		return errors.Wrapf(err, "Error clearing namespace %s", config.Namespace)
	}
	log.Info("Deleted all clusters", "Namespace", config.Namespace)
	return nil
}

func getURLOverride(config *discovery.DiscoveryConfig) string {
	if annotations := config.GetAnnotations(); annotations != nil {
		return annotations[baseURLAnnotation]
	}
	return ""
}

// merge adds the cluster to the cluster map. If the cluster name is already in the map then it
// appends the credentials to the cluster in the map
func merge(clusters map[string]discovery.DiscoveredCluster, dc discovery.DiscoveredCluster) {
	id := dc.Spec.Name
	current, ok := clusters[id]
	if !ok {
		clusters[id] = dc
		return
	}

	secretRef := dc.Spec.ProviderConnections
	current.Spec.ProviderConnections = append(current.Spec.ProviderConnections, secretRef...)
	clusters[id] = current
}

func same(c1, c2 discovery.DiscoveredCluster) bool {
	c1i, c2i := c1.Spec, c2.Spec
	if c1i.CloudProvider != c2i.CloudProvider {
		return false
	}
	if c1i.Console != c2i.Console {
		return false
	}
	if c1i.Name != c2i.Name {
		return false
	}
	if c1i.DisplayName != c2i.DisplayName {
		return false
	}
	if c1i.OpenshiftVersion != c2i.OpenshiftVersion {
		return false
	}
	if c1i.IsManagedCluster != c2i.IsManagedCluster {
		return false
	}
	if c1i.Credential != c2i.Credential {
		return false
	}
	if len(c1i.ProviderConnections) != len(c2i.ProviderConnections) {
		return false
	}
	for i := 0; i < len(c1i.ProviderConnections); i++ {
		if c1i.ProviderConnections[i] != c2i.ProviderConnections[i] {
			return false
		}
	}
	return true
}

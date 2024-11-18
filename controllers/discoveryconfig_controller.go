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
	"strings"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ref "k8s.io/client-go/tools/reference"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log"

	discovery "github.com/stolostron/discovery/api/v1"
	"github.com/stolostron/discovery/pkg/ocm"
	"github.com/stolostron/discovery/pkg/ocm/auth"
	recon "github.com/stolostron/discovery/util/reconciler"
	corev1 "k8s.io/api/core/v1"
)

const (
	defaultDiscoveryConfigName = "discovery"
)

var logf = log.Log.WithName("reconcile")

var (
	// baseURLAnnotation is the annotation set in a DiscoveryConfig that overrides the URL base used to find clusters
	baseURLAnnotation     = "ocmBaseURL"
	baseAuthURLAnnotation = "authBaseURL"
)

var ErrBadFormat = errors.New("bad format")

var mockDiscoveredCluster = func() ([]discovery.DiscoveredCluster, error) {
	return []discovery.DiscoveredCluster{}, nil
}

// DiscoveryConfigReconciler reconciles a DiscoveryConfig object
type DiscoveryConfigReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups="",resources=namespaces;secrets,verbs=create;get;list;update;watch
// +kubebuilder:rbac:groups=discovery.open-cluster-management.io,resources=discoveryconfigs,verbs=create;delete;get;list;patch;update;watch
// +kubebuilder:rbac:groups=discovery.open-cluster-management.io,resources=discoveryconfigs/finalizers,verbs=get;patch;update
// +kubebuilder:rbac:groups=discovery.open-cluster-management.io,resources=discoveryconfigs/status,verbs=get;patch;update

func (r *DiscoveryConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logf.Info("Reconciling DiscoveryConfig", "Name", req.Name, "Namespace", req.Namespace)
	if req.Name == "" || req.Namespace == "" {
		return ctrl.Result{}, nil
	}

	// Update custom metric based on the number of items in the DiscoveryConfigList.
	if err := r.updateCustomMetrics(ctx); err != nil {
		return ctrl.Result{}, err
	}

	// Validate that the request name matches the discovery config name.
	if err := r.validateDiscoveryConfigName(req.Name); err != nil {
		return ctrl.Result{}, err
	}

	config := &discovery.DiscoveryConfig{}
	err := r.Get(ctx, req.NamespacedName, config)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logf.Info("DiscoveryConfig resource not found. Ignoring since object may have been deleted.")
			return ctrl.Result{}, nil
		}

		// If there's an error other than "Not Found", return with the error.
		return ctrl.Result{}, fmt.Errorf("failed to get DiscoveryConfig %s: %w", req.Name, err)
	}

	if err = r.updateDiscoveredClusters(ctx, config); err != nil {
		logf.Error(err, "Error updating DiscoveredClusters")
		return ctrl.Result{}, err
	}

	logf.Info("Reconciliation complete. Scheduling next reconcilation for", "next_time", recon.DefaultRefreshInterval)
	return ctrl.Result{RequeueAfter: recon.DefaultRefreshInterval}, nil
}

// SetupWithManager ...
func (r *DiscoveryConfigReconciler) SetupWithManager(mgr ctrl.Manager) (controller.Controller, error) {
	return ctrl.NewControllerManagedBy(mgr).
		For(&discovery.DiscoveryConfig{}).
		Build(r)
}

func (r *DiscoveryConfigReconciler) updateDiscoveredClusters(ctx context.Context, config *discovery.DiscoveryConfig) error {
	allClusters := map[string]discovery.DiscoveredCluster{}

	// Fetch secret that contains ocm credentials.
	secretName := config.Spec.Credential
	ocmSecret := &corev1.Secret{}
	if err := r.Get(context.TODO(),
		types.NamespacedName{Name: secretName, Namespace: config.Namespace}, ocmSecret); err != nil {

		if apierrors.IsNotFound(err) {
			logf.Info("Secret does not exist. Deleting all clusters.", "Secret", secretName)
			return r.deleteAllClusters(ctx, config)
		}

		logf.Error(err, "Failed to retrieve secret", "Secret", secretName)
		return err
	}

	// Update secret to include default authentication method if the field is missing.
	if _, found := ocmSecret.Data["auth_method"]; !found {
		if err := r.AddDefaultAuthMethodToSecret(ctx, ocmSecret); err != nil {
			return err
		}
	}

	// Parse user token from ocm secret.
	authRequest, err := parseSecretForAuth(ocmSecret)
	if err != nil {
		logf.Error(err, "Error parsing token from secret. Deleting all clusters.", "Secret", ocmSecret.GetName())
		return r.deleteAllClusters(ctx, config)
	}

	// Set the baseURL for authentication requests.
	authRequest.BaseURL = getURLOverride(config)
	authRequest.BaseAuthURL = getAuthURLOverride(config)
	filters := config.Spec.Filters

	var discovered []discovery.DiscoveredCluster
	if val, ok := os.LookupEnv("UNIT_TEST"); ok && val == "true" {
		discovered, err = mockDiscoveredCluster()
	} else {
		discovered, err = ocm.DiscoverClusters(authRequest, filters)
	}

	if err != nil {
		if ocm.IsUnrecoverable(err) || ocm.IsUnauthorizedClient(err) || ocm.IsInvalidClient(err) {
			logf.Info("Error encountered. Cleaning up clusters.", "Error", err.Error())
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
		dc.Spec.Credential = *secretRef
		allClusters[dc.Spec.Name] = dc
	}

	// Assign managed status
	managed, err := r.getManagedClusters()
	if err != nil {
		return err
	}

	if len(managed) > 0 {
		assignManagedStatus(allClusters, managed)
	} else {
		logf.Info("No managed clusters found in the list of clusters.")
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

/*
ValidateDiscoveryConfigName validates the name of the DiscoveryConfig resource.
It ensures that the provided name matches the defaultDiscoveryConfigName.
*/
func (r *DiscoveryConfigReconciler) validateDiscoveryConfigName(reqName string) error {
	if reqName != defaultDiscoveryConfigName {
		return fmt.Errorf("invalid DiscoveryConfig resource name '%s', it must be '%s'",
			reqName, defaultDiscoveryConfigName)
	}
	return nil
}

/*
parseSecretForAuth parses the given Secret to retrieve authentication credentials.
Depending on the "auth_method" field in the secret, it returns either service account credentials
(client_id, client_secret) or an offline token (ocmAPIToken). If "auth_method" is not set, it
defaults to using the "offline-token" method. Returns an error if the expected fields are missing.
*/
func parseSecretForAuth(secret *corev1.Secret) (auth.AuthRequest, error) {
	// Set the default auth_method to "offline-token"
	authMethod := "offline-token"

	// Check if the "auth_method" key is present in the Secret data
	if method, found := secret.Data["auth_method"]; found {
		authMethod = string(method)
	}

	credentials := auth.AuthRequest{
		AuthMethod: strings.TrimSuffix(string(authMethod), "\n"), // Set the authentication method
	}

	// Handle based on the "auth_method" value
	switch credentials.AuthMethod {
	case "service-account":
		// Retrieve client_id and client_secret for service-account auth method
		clientID, idOk := secret.Data["client_id"]
		clientSecret, secretOk := secret.Data["client_secret"]

		if !idOk || !secretOk {
			return credentials, fmt.Errorf(
				"%s: bad format: secret must contain client_id and client_secret", secret.Name)
		}

		credentials.ID = strings.TrimSuffix(string(clientID), "\n")
		credentials.Secret = strings.TrimSuffix(string(clientSecret), "\n")

	case "offline-token":
		// Retrive ocmAPIToken for offline-token auth method
		token, tokenOk := secret.Data["ocmAPIToken"]
		if !tokenOk {
			return credentials, fmt.Errorf("%s: bad format: secret must contain ocmAPIToken", secret.Name)
		}

		credentials.Token = strings.TrimSuffix(string(token), "\n")

	default:
		return credentials, fmt.Errorf("%s: bad format: unsupported auth_method:  %s", secret.Name, authMethod)
	}

	return credentials, nil
}

func (r *DiscoveryConfigReconciler) AddDefaultAuthMethodToSecret(ctx context.Context, secret *corev1.Secret) error {
	// Set the default auth_method to "offline-token"
	secret.Data["auth_method"] = []byte("offline-token")

	// Check if both client_id and client_secret are present in the secret data
	if _, idOK := secret.Data["client_id"]; idOK {
		if _, secretOk := secret.Data["client_secret"]; secretOk {
			secret.Data["auth_method"] = []byte("service-account")
		}
	}

	// Update the secret
	if err := r.Client.Update(ctx, secret); err != nil {
		logf.Error(err, "failed to update Secret with default auth_method: 'offline-token'", "Name", secret.GetName())
		return err
	}

	return nil
}

// assignManagedStatus marks clusters in the discovered map as managed if they are in the managed list
func assignManagedStatus(discovered map[string]discovery.DiscoveredCluster, managed []metav1.PartialObjectMetadata) {
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

func (r *DiscoveryConfigReconciler) getManagedClusters() ([]metav1.PartialObjectMetadata, error) {
	ctx := context.Background()

	managedMeta := &metav1.PartialObjectMetadataList{TypeMeta: metav1.TypeMeta{Kind: "ManagedClusterList", APIVersion: "cluster.open-cluster-management.io/v1"}}
	if err := r.Client.List(ctx, managedMeta); client.IgnoreNotFound(err) != nil {
		return nil, errors.Wrapf(err, "error listing managed clusters")
	}
	return managedMeta.Items, nil
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

	if dc.Equal(current) {
		// Discovered cluster has not changed
		return nil
	}

	// Cluster needs to be updated
	return r.updateCluster(ctx, dc, current)
}

func (r *DiscoveryConfigReconciler) createCluster(ctx context.Context, config *discovery.DiscoveryConfig, dc discovery.DiscoveredCluster) error {
	// Try to get the existing cluster
	existingCluster := &discovery.DiscoveredCluster{}
	err := r.Get(ctx, types.NamespacedName{Namespace: dc.GetNamespace(), Name: dc.GetName()}, existingCluster)
	if err == nil {
		// Cluster already exists, log and return
		logf.Info("Cluster already exists, skipping creation", "Name", dc.Name)
		return nil
	}

	// Set controller reference
	if err := ctrl.SetControllerReference(config, &dc, r.Scheme); err != nil {
		return errors.Wrapf(err, "Error setting controller reference on DiscoveredCluster %s", dc.Name)
	}

	// Create the cluster
	if err := r.Create(ctx, &dc); err != nil {
		if apierrors.IsAlreadyExists(err) {
			return nil
		}
		return errors.Wrapf(err, "Error creating DiscoveredCluster %s", dc.Name)
	}

	logf.Info("Created cluster", "Name", dc.Name)
	return nil
}

func (r *DiscoveryConfigReconciler) updateCluster(ctx context.Context, new, old discovery.DiscoveredCluster) error {
	updated := old
	updated.Spec = new.Spec
	if err := r.Update(ctx, &updated); err != nil {
		return errors.Wrapf(err, "Error updating DiscoveredCluster %s", updated.Name)
	}

	logf.Info("Updated cluster", "Name", updated.Name)
	return nil
}

func (r *DiscoveryConfigReconciler) deleteCluster(ctx context.Context, dc discovery.DiscoveredCluster) error {
	if err := r.Delete(ctx, &dc); err != nil {
		if apierrors.IsNotFound(err) {
			logf.Info("Cluster does not exist, skipping deletion", "Name", dc.Name)
			return nil
		}
		return errors.Wrapf(err, "Error deleting DiscoveredCluster %s", dc.Name)
	}

	logf.Info("Deleted cluster", "Name", dc.Name)
	return nil
}

func (r *DiscoveryConfigReconciler) deleteAllClusters(ctx context.Context, config *discovery.DiscoveryConfig) error {
	log, _ := logr.FromContext(ctx)
	if err := r.DeleteAllOf(ctx, &discovery.DiscoveredCluster{}, client.InNamespace(config.Namespace)); err != nil {
		return errors.Wrapf(err, "Error clearing namespace %s", config.Namespace)
	}
	log.Info("Deleted all clusters", "Namespace", config.Namespace)
	return nil
}

/*
updateCustomMetrics updates the totalConfigs Prometheus metric based on the number of items in the
DiscoveryConfigList retrieved from the cluster. It retrieves the list of DiscoveryConfigs,
sets the totalConfigs metric to the length of the items list, and reports any errors encountered
during the process.
*/
func (r *DiscoveryConfigReconciler) updateCustomMetrics(ctx context.Context) error {
	configs := &discovery.DiscoveryConfigList{}
	if err := r.List(ctx, configs); err != nil {
		return fmt.Errorf("failed to list DiscoveryConfigs: %w", err)
	}

	totalConfigs.Set(float64(len(configs.Items)))
	return nil
}

func getURLOverride(config *discovery.DiscoveryConfig) string {
	if annotations := config.GetAnnotations(); annotations != nil {
		return annotations[baseURLAnnotation]
	}
	return ""
}

func getAuthURLOverride(config *discovery.DiscoveryConfig) string {
	if annotations := config.GetAnnotations(); annotations != nil {
		return annotations[baseAuthURLAnnotation]
	}
	return ""
}

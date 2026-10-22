// Copyright Contributors to the Open Cluster Management project
/*
Copyright 2021.

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

package v1

import (
	"context"
	"fmt"
	"regexp"

	admissionregistration "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	cl "sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var (
	discoveredclusterLog = logf.Log.WithName("discoveredcluster-resource")
	Client               cl.Client
)

// ValidatingWebhook returns the ValidatingWebhookConfiguration used for the discoveredcluster
// linked to a service in the provided namespace
func ValidatingWebhook(namespace string) *admissionregistration.ValidatingWebhookConfiguration {
	fail := admissionregistration.Fail
	none := admissionregistration.SideEffectClassNone
	path := "/validate-discovery-open-cluster-management-io-v1-discoveredcluster"
	return &admissionregistration.ValidatingWebhookConfiguration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "admissionregistration.k8s.io/v1",
			Kind:       "ValidatingWebhookConfiguration",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        "discovery.open-cluster-management.io",
			Annotations: map[string]string{"service.beta.openshift.io/inject-cabundle": "true"},
		},
		Webhooks: []admissionregistration.ValidatingWebhook{
			{
				AdmissionReviewVersions: []string{
					"v1",
					"v1beta1",
				},
				Name: "discovery.open-cluster-management.io",
				ClientConfig: admissionregistration.WebhookClientConfig{
					Service: &admissionregistration.ServiceReference{
						Name:      "discovery-operator-webhook-service",
						Namespace: namespace,
						Path:      &path,
					},
				},
				FailurePolicy: &fail,
				Rules: []admissionregistration.RuleWithOperations{
					{
						Rule: admissionregistration.Rule{
							APIGroups:   []string{GroupVersion.Group},
							APIVersions: []string{GroupVersion.Version},
							Resources:   []string{"discoveredclusters"},
						},
						Operations: []admissionregistration.OperationType{
							admissionregistration.Create,
							admissionregistration.Update,
							admissionregistration.Delete,
						},
					},
				},
				SideEffects: &none,
			},
		},
	}
}

func (r *DiscoveredCluster) SetupWebhookWithManager(mgr ctrl.Manager) error {
	Client = mgr.GetClient()
	return builder.WebhookManagedBy(mgr, r).
		WithDefaulter(r).
		WithValidator(r).
		Complete()
}

var _ admission.Defaulter[*DiscoveredCluster] = &DiscoveredCluster{}

// Default implements admission.Defaulter so a webhook will be registered for the type
func (r *DiscoveredCluster) Default(_ context.Context, obj *DiscoveredCluster) error {
	discoveredclusterLog.Info("default", "Name", obj.Name)
	return nil
}

var _ admission.Validator[*DiscoveredCluster] = &DiscoveredCluster{}

// ValidateCreate implements admission.Validator so a webhook will be registered for the type
func (r *DiscoveredCluster) ValidateCreate(_ context.Context, obj *DiscoveredCluster) (admission.Warnings, error) {
	discoveredclusterLog.Info("validate create", "Name", obj.Name, "Type", obj.Spec.Type)

	// Validate resource
	if !IsSupportedClusterType(obj.Spec.Type) && obj.Spec.ImportAsManagedCluster {
		err := fmt.Errorf(
			"cannot create DiscoveredCluster '%s': importAsManagedCluster is not allowed for clusters of type '%s'. "+
				"Only ROSA type clusters support auto import", obj.Name, obj.Spec.Type)

		discoveredclusterLog.Error(err, "validation failed")
		return nil, err
	}

	if !IsStringValid(obj.Spec.DisplayName) && obj.Spec.ImportAsManagedCluster {
		err := fmt.Errorf(
			"cannot update DiscoveredCluster '%s': importAsManagedCluster is not allowed for clusters with an invalid display name '%s'. "+
				"Display name must consist of lowercase alphanumeric characters or '-'", obj.Name, obj.Spec.DisplayName)

		discoveredclusterLog.Error(err, "validation failed")
		return nil, err
	}

	return nil, nil
}

// ValidateUpdate implements admission.Validator so a webhook will be registered for the type
func (r *DiscoveredCluster) ValidateUpdate(_ context.Context, oldObj, newObj *DiscoveredCluster) (admission.Warnings, error) {
	discoveredclusterLog.Info("validate update", "Name", newObj.Name, "Type", newObj.Spec.Type)

	// Validate resource
	if !IsSupportedClusterType(oldObj.Spec.Type) && newObj.Spec.ImportAsManagedCluster {
		err := fmt.Errorf(
			"cannot update DiscoveredCluster '%s': importAsManagedCluster is not allowed for clusters of type '%s'. "+
				"Only ROSA type clusters support auto import", newObj.Name, newObj.Spec.Type)

		discoveredclusterLog.Error(err, "validation failed")
		return nil, err
	}

	if !IsStringValid(newObj.Spec.DisplayName) && newObj.Spec.ImportAsManagedCluster {
		err := fmt.Errorf(
			"cannot update DiscoveredCluster '%s': importAsManagedCluster is not allowed for clusters with an invalid display name '%s'. "+
				"Display name must consist of lowercase alphanumeric characters or '-'", newObj.Name, newObj.Spec.DisplayName)

		discoveredclusterLog.Error(err, "validation failed")
		return nil, err
	}

	return nil, nil
}

// ValidateDelete implements admission.Validator so a webhook will be registered for the type
func (r *DiscoveredCluster) ValidateDelete(_ context.Context, obj *DiscoveredCluster) (admission.Warnings, error) {
	discoveredclusterLog.Info("validate delete", "Name", obj.Name, "Type", obj.Spec.Type)
	return nil, nil
}

// IsSupportedClusterType returns true if the cluster type is supported by the registry
func IsSupportedClusterType(clusterType string) bool {
	supportedTypes := map[string]bool{
		"MultiClusterEngineHCP": true,
		"ROSA":                  true,
	}

	return supportedTypes[clusterType]
}

// IsStringValid returns true if the string conforms to the RFC 1123 standards.
func IsStringValid(s string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
	return re.MatchString(s)
}

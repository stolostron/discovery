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
	"fmt"

	"github.com/stolostron/discovery/pkg/common"
	admissionregistration "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	cl "sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
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
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

var _ webhook.Defaulter = &DiscoveredCluster{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *DiscoveredCluster) Default() {
	discoveredclusterLog.Info("default", "Name", r.Name)
}

var _ webhook.Validator = &DiscoveredCluster{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *DiscoveredCluster) ValidateCreate() (admission.Warnings, error) {
	discoveredclusterLog.Info("validate create", "Name", r.Name, "Type", r.Spec.Type)

	// Validate resource
	if !common.IsSupportedClusterType(r.Spec.Type) && r.Spec.ImportAsManagedCluster {
		err := fmt.Errorf(
			"cannot create DiscoveredCluster '%s': importAsManagedCluster is not allowed for clusters of type '%s'. "+
				"Only ROSA type clusters support auto import", r.Name, r.Spec.Type)

		discoveredclusterLog.Error(err, "validation failed")
		return nil, err
	}

	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *DiscoveredCluster) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	discoveredclusterLog.Info("validate update", "Name", r.Name, "Type", r.Spec.Type)

	// Validate resource
	oldDiscoveredCluster := old.(*DiscoveredCluster)
	if !common.IsSupportedClusterType(oldDiscoveredCluster.Spec.Type) && r.Spec.ImportAsManagedCluster {
		err := fmt.Errorf(
			"cannot update DiscoveredCluster '%s': importAsManagedCluster is not allowed for clusters of type '%s'. "+
				"Only ROSA type clusters support auto import", r.Name, r.Spec.Type)

		discoveredclusterLog.Error(err, "validation failed")
		return nil, err
	}

	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *DiscoveredCluster) ValidateDelete() (admission.Warnings, error) {
	discoveredclusterLog.Info("validate delete", "Name", r.Name, "Type", r.Spec.Type)
	return nil, nil
}

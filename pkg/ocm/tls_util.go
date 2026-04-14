// Copyright Contributors to the Open Cluster Management project

package ocm

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"

	configv1 "github.com/openshift/api/config/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// UnitTestEnvVar is the environment variable for unit testing.
	UnitTestEnvVar = "UNIT_TEST"
)

// GetAPIServerTLSProfile retrieves the TLS security profile from the OpenShift APIServer resource.
// Returns the TLSProfileSpec containing minTLSVersion and ciphers.
// If no profile is set, returns the default Intermediate profile.
func GetAPIServerTLSProfile(ctx context.Context, cl client.Client) (*configv1.TLSProfileSpec, error) {
	// If running in unit test mode, return default Intermediate profile
	if val, ok := os.LookupEnv(UnitTestEnvVar); ok && val == "true" {
		return configv1.TLSProfiles[configv1.TLSProfileIntermediateType], nil
	}

	apiServer := &configv1.APIServer{}
	err := cl.Get(ctx, types.NamespacedName{Name: "cluster"}, apiServer)
	if err != nil {
		return nil, fmt.Errorf("failed to get APIServer resource: %w", err)
	}

	// If no TLS profile is set, use the default (Intermediate)
	if apiServer.Spec.TLSSecurityProfile == nil {
		return configv1.TLSProfiles[configv1.TLSProfileIntermediateType], nil
	}

	profile := apiServer.Spec.TLSSecurityProfile

	// For predefined profiles (Old, Intermediate, Modern), use the map
	if profileSpec, ok := configv1.TLSProfiles[profile.Type]; ok {
		return profileSpec, nil
	}

	// For custom profile, return the inline spec
	if profile.Type == configv1.TLSProfileCustomType && profile.Custom != nil {
		return &profile.Custom.TLSProfileSpec, nil
	}

	// Fallback to Intermediate if something unexpected
	return configv1.TLSProfiles[configv1.TLSProfileIntermediateType], nil
}

// ConvertTLSVersion converts OpenShift TLSProtocolVersion string to crypto/tls uint16 constant.
// Returns tls.VersionTLS12 as default if the version string is not recognized.
func ConvertTLSVersion(version configv1.TLSProtocolVersion) uint16 {
	switch version {
	case configv1.VersionTLS10:
		return tls.VersionTLS10
	case configv1.VersionTLS11:
		return tls.VersionTLS11
	case configv1.VersionTLS12:
		return tls.VersionTLS12
	case configv1.VersionTLS13:
		return tls.VersionTLS13
	default:
		// Default to TLS 1.2 for safety
		return tls.VersionTLS12
	}
}

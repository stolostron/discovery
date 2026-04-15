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

// ConvertCipherSuites converts OpenShift cipher suite names (OpenSSL format) to crypto/tls uint16 constants.
// TLS 1.3 cipher suites are managed automatically by Go and cannot be configured, so they are filtered out.
// Only returns cipher suites applicable to TLS ≤ 1.2.
func ConvertCipherSuites(cipherNames []string) []uint16 {
	// Mapping from OpenSSL cipher names to crypto/tls constants
	// Only includes cipher suites that exist in Go's crypto/tls package
	cipherMap := map[string]uint16{
		// TLS 1.2 ECDHE ciphers (GCM and ChaCha20)
		"ECDHE-RSA-AES128-GCM-SHA256":   tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		"ECDHE-RSA-AES256-GCM-SHA384":   tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		"ECDHE-ECDSA-AES128-GCM-SHA256": tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		"ECDHE-ECDSA-AES256-GCM-SHA384": tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		"ECDHE-RSA-CHACHA20-POLY1305":   tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
		"ECDHE-ECDSA-CHACHA20-POLY1305": tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,

		// TLS 1.2 ECDHE ciphers (CBC)
		"ECDHE-RSA-AES128-SHA256":   tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
		"ECDHE-RSA-AES128-SHA":      tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
		"ECDHE-ECDSA-AES128-SHA256": tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
		"ECDHE-ECDSA-AES128-SHA":    tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
		"ECDHE-RSA-AES256-SHA":      tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
		"ECDHE-ECDSA-AES256-SHA":    tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,

		// RSA ciphers
		"AES128-GCM-SHA256": tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
		"AES256-GCM-SHA384": tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
		"AES128-SHA256":     tls.TLS_RSA_WITH_AES_128_CBC_SHA256,
		"AES128-SHA":        tls.TLS_RSA_WITH_AES_128_CBC_SHA,
		"AES256-SHA":        tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		"DES-CBC3-SHA":      tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
	}

	var result []uint16
	for _, name := range cipherNames {
		// Skip TLS 1.3 cipher suites (they start with TLS_ prefix and are auto-managed)
		if len(name) > 4 && name[:4] == "TLS_" {
			continue
		}

		if cipher, ok := cipherMap[name]; ok {
			result = append(result, cipher)
		}
		// Silently skip unsupported cipher suites - Go may not support all OpenSSL ciphers
	}

	return result
}

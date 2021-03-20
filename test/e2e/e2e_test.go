package e2e

import (
	"testing"
	// +kubebuilder:scaffold:imports
)

func TestE2E(t *testing.T) {
	RunE2ETests(t)
}

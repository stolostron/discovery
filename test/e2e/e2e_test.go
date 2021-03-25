package e2e

import (
	"testing"
	// +kubebuilder:scaffold:imports
)

func TestE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	RunE2ETests(t)
}

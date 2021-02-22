// Copyright Contributors to the Open Cluster Management project

package reconciler

import (
	"time"
)

const (
	// RefreshInterval is the maximum time to wait between reconciles
	RefreshInterval = 30 * time.Minute
)

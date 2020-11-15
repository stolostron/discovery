package reconciler

import (
	"time"
)

const (
	// RefreshInterval is the maximum time to wait between reconciles
	RefreshInterval = 90 * time.Minute
)

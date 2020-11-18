package reconciler

import (
	"time"
)

const (
	// RefreshInterval is the maximum time to wait between reconciles
	RefreshInterval = 30 * time.Minute
)

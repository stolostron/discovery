// Copyright Contributors to the Open Cluster Management project

package controllers

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	totalConfigs = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "discovery_config_total",
			Help: "Number of discoveryConfigs across namespaces",
		},
	)
)

func init() {
	// Register custom metrics with the global prometheus registry
	metrics.Registry.MustRegister(totalConfigs)
}

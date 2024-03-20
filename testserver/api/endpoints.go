// Copyright Contributors to the Open Cluster Management project

package api

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// SetupEndpoints ...
func SetupEndpoints(r *gin.Engine, logger zerolog.Logger) {
	r.GET("/scenarios/:scenario", SetScenario)
	r.GET("/api/clusters_mgmt/v1/clusters", GetClusters)
	r.GET("/api/accounts_mgmt/v1/subscriptions", GetSubscriptions)
	r.POST("/auth/realms/redhat-external/protocol/openid-connect/token", GetToken)

	// Add route group for each scenario
	for k := range scenarios {
		addRoutes(r.Group(k), k)
	}

}

func addRoutes(rg *gin.RouterGroup, scenario string) {
	s := Scenario(scenario)
	rg.GET("/api/clusters_mgmt/v1/clusters", s.GetClusters)
	rg.GET("/api/accounts_mgmt/v1/subscriptions", s.GetSubscriptions)
	rg.POST("/auth/realms/redhat-external/protocol/openid-connect/token", GetToken)
}

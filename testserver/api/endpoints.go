package api

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// SetupEndpoints ...
func SetupEndpoints(r *gin.Engine, logger zerolog.Logger) {
	r.GET("/scenarios/:scenario", SetScenario)
	r.GET("/api/clusters_mgmt/v1/clusters", GetCluster)
	r.GET("/api/accounts_mgmt/v1/subscriptions", GetSubscriptions)
	r.POST("/auth/realms/redhat-external/protocol/openid-connect/token", GetToken)
}

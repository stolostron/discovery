package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

var dataFolder = "data"

var scenarios = map[string]string{
	// responds with 10 active clusters
	"tenClusters": "data/scenarios/ten_clusters",
}

// SetScenario sets up predetermined api responses to simulate various scenarios
// for testing
func SetScenario(c *gin.Context) {
	name := c.Param("scenario")
	fmt.Println(name)
	scenarioPath, ok := scenarios[name]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": fmt.Sprintf("'%s' is not a valid scenario", name),
		})
		return
	}

	dataFolder = scenarioPath
	c.Status(http.StatusOK)
}

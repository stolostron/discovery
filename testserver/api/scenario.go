package api

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Scenario is
var scenario string

var dataFolder = "data"

var scenarios = map[string]string{
	// responds with 10 active clusters
	"tenClusters": "data/scenarios/ten_clusters",
}

func init() {
	flag.StringVar(&scenario, "scenario", "tenClusters", "The address the metric endpoint binds to.")
	flag.Parse()

	dataFolder = scenarios[scenario]
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

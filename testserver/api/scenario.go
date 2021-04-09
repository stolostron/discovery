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
	// responds with 10 active/stale clusters
	"tenClusters": "data/scenarios/ten_clusters",
	// responds with 5 active/stale clusters
	"fiveClusters": "data/scenarios/five_clusters",
	// responds with 10 clusters: 8 active/stale, 2 archived
	"archivedClusters": "data/scenarios/archived_clusters",
	// responds with one of two options, depending on auth header content
	"multipleConnections": "data/scenarios/multiple_connections",
}

func init() {
	flag.StringVar(&scenario, "scenario", "tenClusters", "The address the metric endpoint binds to.")
	flag.Parse()
	fmt.Println("Starting with scenario " + scenario)

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

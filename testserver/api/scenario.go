package api

import "github.com/gin-gonic/gin"

var dataFolder = "data"

var scenarios = map[string]string{
	// clustersAdded starts with a set number of clusters and adds more
	"clustersAdded":   "data/clusters_added",
	"clustersRemoved": "data/cluster_removed",
}

// SetScenario sets up predetermined api responses to simulate various scenarios
// for testing
func SetScenario(c *gin.Context) {
	name := c.Param("playbook")
	dataFolder = scenarios[name]

}

package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/open-cluster-management/discovery/pkg/api/domain/cluster_domain"
)

// GetCluster ...
func GetCluster(c *gin.Context) {
	file, err := ioutil.ReadFile(path.Join(dataFolder, "cluster_response.json"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": fmt.Sprintf("Error reading file: %s", err.Error()),
		})
		return
	}

	// Validate file can be unmarshalled into ClusterResponse
	var clusterList cluster_domain.ClusterResponse
	err = json.Unmarshal(file, &clusterList)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": fmt.Sprintf("Error unmarshalling JSON: %s", err.Error()),
		})
		return
	}

	c.Data(http.StatusOK, "application/json", file)
}

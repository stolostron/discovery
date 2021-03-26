package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/open-cluster-management/discovery/pkg/ocm/domain/cluster_domain"
)

// GetCluster ...
func GetCluster(c *gin.Context) {
	header := c.Request.Header["Authorization"]
	auth := strings.Join(header, " ")

	var file []byte
	var err error
	if strings.Contains(auth, "connection1") {
		fmt.Println("Returning connection1 response")
		file, err = ioutil.ReadFile(path.Join(dataFolder, "connection1/cluster_response.json"))
	} else if strings.Contains(auth, "connection2") {
		fmt.Println("Returning connection2 response")
		file, err = ioutil.ReadFile(path.Join(dataFolder, "connection2/cluster_response.json"))
	} else {
		file, err = ioutil.ReadFile(path.Join(dataFolder, "cluster_response.json"))
	}

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

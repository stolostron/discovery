// Copyright Contributors to the Open Cluster Management project

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stolostron/discovery/pkg/ocm/common"
)

// GetClusters ...
func GetClusters(c *gin.Context) {
	header := c.Request.Header["Authorization"]
	auth := strings.Join(header, " ")
	params := c.Request.URL.Query()
	page := params.Get("page")

	var file []byte
	var err error
	if strings.Contains(auth, "connection1") {
		fmt.Println("Returning connection1 response")
		file, err = os.ReadFile(path.Join(dataFolder, "connection1/cluster_response.json"))
	} else if strings.Contains(auth, "connection2") {
		fmt.Println("Returning connection2 response")
		file, err = os.ReadFile(path.Join(dataFolder, "connection2/cluster_response.json"))
	} else {
		file, err = os.ReadFile(paginate(path.Join(dataFolder, "cluster_response.json"), page))
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": fmt.Sprintf("Error reading file: %s", err.Error()),
		})
		return
	}

	file = setTime(file, time.Now())

	// Validate file can be unmarshalled into clusterResponse
	var clusterList common.Response
	err = json.Unmarshal(file, &clusterList)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": fmt.Sprintf("Error unmarshalling JSON: %s", err.Error()),
		})
		return
	}

	c.Data(http.StatusOK, "application/json", file)
}

// Behaves the same as GetClusters, but using a set scenario rather than the dataFolder var
func (s Scenario) GetClusters(c *gin.Context) {
	header := c.Request.Header["Authorization"]
	auth := strings.Join(header, " ")
	params := c.Request.URL.Query()
	page := params.Get("page")

	var file []byte
	var err error
	if strings.Contains(auth, "connection1") {
		fmt.Println("Returning connection1 response")
		file, err = os.ReadFile(path.Join(s.Path(), "connection1/cluster_response.json"))
	} else if strings.Contains(auth, "connection2") {
		fmt.Println("Returning connection2 response")
		file, err = os.ReadFile(path.Join(s.Path(), "connection2/cluster_response.json"))
	} else {
		file, err = os.ReadFile(paginate(path.Join(s.Path(), "cluster_response.json"), page))
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": fmt.Sprintf("Error reading file: %s", err.Error()),
		})
		return
	}

	file = setTime(file, time.Now())

	// Validate file can be unmarshalled into ClusterResponse
	var clusterList common.Response
	err = json.Unmarshal(file, &clusterList)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": fmt.Sprintf("Error unmarshalling JSON: %s", err.Error()),
		})
		return
	}

	c.Data(http.StatusOK, "application/json", file)
}

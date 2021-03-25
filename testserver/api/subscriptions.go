package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/open-cluster-management/discovery/pkg/ocm/domain/subscription_domain"
)

// GetSubscriptions ...
func GetSubscriptions(c *gin.Context) {
	token := strings.Split(c.Request.Header["Authorization"][0], " ")[1]

	var file []byte
	var err error
	if strings.Contains(token, "connection1") {
		fmt.Println("Returning connection1 response")
		file, err = ioutil.ReadFile(path.Join(dataFolder, "connection1/subscription_response.json"))
	} else if strings.Contains(token, "connection2") {
		fmt.Println("Returning connection2 response")
		file, err = ioutil.ReadFile(path.Join(dataFolder, "connection2/subscription_response.json"))
	} else {
		file, err = ioutil.ReadFile(path.Join(dataFolder, "subscription_response.json"))
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": fmt.Sprintf("Error reading file: %s", err.Error()),
		})
		return
	}

	// Validate file can be unmarshalled into SubscriptionResponse
	var subscriptionList subscription_domain.SubscriptionResponse
	err = json.Unmarshal(file, &subscriptionList)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": fmt.Sprintf("Error unmarshalling JSON: %s", err.Error()),
		})
		return
	}

	c.Data(http.StatusOK, "application/json", file)
}

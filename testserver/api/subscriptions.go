package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/open-cluster-management/discovery/pkg/api/domain/subscription_domain"
)

// GetSubscriptions ...
func GetSubscriptions(c *gin.Context) {
	// file, err := ioutil.ReadFile("data/subscriptions.json")
	file, err := ioutil.ReadFile(path.Join(dataFolder, "subscription_response.json"))
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

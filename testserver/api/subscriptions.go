// Copyright Contributors to the Open Cluster Management project

package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stolostron/discovery/pkg/ocm/subscription"
)

type Scenario string

func (s Scenario) Path() string {
	return scenarios[string(s)]
}

// GetSubscriptions ...
func GetSubscriptions(c *gin.Context) {
	header := c.Request.Header["Authorization"]
	auth := strings.Join(header, " ")
	params := c.Request.URL.Query()
	page := params.Get("page")

	var file []byte
	var err error
	if strings.Contains(auth, "connection1") {
		fmt.Println("Returning connection1 response")
		file, err = os.ReadFile(path.Join(dataFolder, "connection1/subscription_response.json"))
	} else if strings.Contains(auth, "connection2") {
		fmt.Println("Returning connection2 response")
		file, err = os.ReadFile(path.Join(dataFolder, "connection2/subscription_response.json"))
	} else {
		file, err = os.ReadFile(paginate(path.Join(dataFolder, "subscription_response.json"), page))
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": fmt.Sprintf("Error reading file: %s", err.Error()),
		})
		return
	}

	file = setTime(file, time.Now())

	// Validate file can be unmarshalled into SubscriptionResponse
	var subscriptionList subscription.SubscriptionResponse
	err = json.Unmarshal(file, &subscriptionList)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": fmt.Sprintf("Error unmarshalling JSON: %s", err.Error()),
		})
		return
	}

	c.Data(http.StatusOK, "application/json", file)
}

// Behaves the same as GetSubscriptions, but using a set scenario rather than the dataFolder var
func (s Scenario) GetSubscriptions(c *gin.Context) {
	header := c.Request.Header["Authorization"]
	auth := strings.Join(header, " ")
	params := c.Request.URL.Query()
	page := params.Get("page")

	var file []byte
	var err error
	if strings.Contains(auth, "connection1") {
		fmt.Println("Returning connection1 response")
		file, err = os.ReadFile(path.Join(s.Path(), "connection1/subscription_response.json"))
	} else if strings.Contains(auth, "connection2") {
		fmt.Println("Returning connection2 response")
		file, err = os.ReadFile(path.Join(s.Path(), "connection2/subscription_response.json"))
	} else {
		file, err = os.ReadFile(paginate(path.Join(s.Path(), "subscription_response.json"), page))
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": fmt.Sprintf("Error reading file: %s", err.Error()),
		})
		return
	}

	file = setTime(file, time.Now())

	// Validate file can be unmarshalled into SubscriptionResponse
	var subscriptionList subscription.SubscriptionResponse
	err = json.Unmarshal(file, &subscriptionList)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": fmt.Sprintf("Error unmarshalling JSON: %s", err.Error()),
		})
		return
	}

	c.Data(http.StatusOK, "application/json", file)
}

// paginate appends the page number to the filename when page > 1
func paginate(file, page string) string {
	if page == "" || page == "1" {
		return file
	}

	extension := path.Ext(file)
	newPath := strings.TrimSuffix(file, extension)

	return fmt.Sprintf("%s_%s%s", newPath, page, extension)
}

// Replaces time placeholders in the subscription template with the current time
func setTime(in []byte, now time.Time) []byte {
	// truncates to the hour so it doesn't change every time
	return bytes.ReplaceAll(in, []byte("--TODAY--"), []byte(now.Truncate(time.Hour).Format("2006-01-02T15:04:05.000000Z")))
}

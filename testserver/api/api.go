// Copyright Contributors to the Open Cluster Management project

package api

// // SetupEndpoints ...
// func SetupEndpoints(r *gin.Engine, logger zerolog.Logger) {
// 	r.GET("/api/clusters_mgmt/v1/clusters/*clusterID", GetCluster)

// 	r.GET("/api/accounts_mgmt/v1/subscriptions", GetSubscriptions)

// 	r.POST("/auth/realms/redhat-external/protocol/openid-connect/token", GetToken)
// }

// // GetSubscriptions ...
// func GetSubscriptions(c *gin.Context) {
// 	file, err := ioutil.ReadFile("data/subscriptions.json")
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"msg": fmt.Sprintf("Error reading file: %s", err.Error()),
// 		})
// 		return
// 	}

// 	var subscriptionList SubscriptionList
// 	err = json.Unmarshal(file, &subscriptionList)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"msg": fmt.Sprintf("Error unmarshalling JSON: %s", err.Error()),
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusOK, subscriptionList)
// }

// // GetCluster ...
// func GetCluster(c *gin.Context) {
// 	clusterID := c.Param("clusterID")
// 	params := c.Request.URL.Query()

// 	var file []byte
// 	var err error
// 	// Return filtered results if search param set
// 	if _, ok := params["search"]; ok {
// 		file, err = ioutil.ReadFile("data/filtered_clusters_list.json")
// 	} else {
// 		file, err = ioutil.ReadFile("data/clusters_list.json")
// 	}

// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"msg": fmt.Sprintf("Error reading file: %s", err.Error()),
// 		})
// 		return
// 	}

// 	var clusterList ClusterList
// 	err = json.Unmarshal(file, &clusterList)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"msg": fmt.Sprintf("Error unmarshalling JSON: %s", err.Error()),
// 		})
// 		return
// 	}

// 	if clusterID != "" {
// 		c.Data(http.StatusOK, "application/json", file)
// 	}

// 	for _, cluster := range clusterList.Items {
// 		if cluster.ID == clusterID {
// 			c.JSON(http.StatusOK, cluster)
// 			return
// 		}
// 	}

// 	c.Status(http.StatusBadRequest)
// }

// // GetToken
// func GetToken(c *gin.Context) {
// 	token := c.PostForm("refresh_token")
// 	if token == "" {
// 		log.Println("Empty token received. Responding with auth error.")
// 		file, err := ioutil.ReadFile("data/auth_error.json")
// 		if err != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{
// 				"msg": fmt.Sprintf("Error reading file: %s", err.Error()),
// 			})
// 			return
// 		}
// 		c.Data(http.StatusBadRequest, "application/json", file)
// 	} else {
// 		log.Println("Auth token received. Responding with auth success.")
// 		file, err := ioutil.ReadFile("data/auth_success.json")
// 		if err != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{
// 				"msg": fmt.Sprintf("Error reading file: %s", err.Error()),
// 			})
// 			return
// 		}
// 		c.Data(http.StatusOK, "application/json", file)
// 	}
// }

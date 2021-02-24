// Copyright Contributors to the Open Cluster Management project

package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// ClusterList ...
type ClusterList struct {
	Kind   string    `yaml:"kind"`
	Page   int       `yaml:"page"`
	Size   int       `yaml:"size"`
	Total  int       `yaml:"total"`
	Items  []Cluster `yaml:"items"`
	Reason string    `yaml:"reason"`
}

// SubscriptionList ...
type SubscriptionList struct {
	Kind   string         `yaml:"kind" json:"kind"`
	Page   int            `yaml:"page" json:"page"`
	Size   int            `yaml:"size" json:"size"`
	Total  int            `yaml:"total" json:"total"`
	Items  []Subscription `yaml:"items" json:"items"`
	Reason string         `yaml:"reason" json:"reason"`
}

// Subscription ...
type Subscription struct {
	ID                string       `yaml:"id" json:"id"`
	Kind              string       `yaml:"kind" json:"kind"`
	Href              string       `yaml:"href" json:"href"`
	Plan              StandardKind `yaml:"plan,omitempty" json:"plan,omitempty"`
	ClusterID         string       `yaml:"cluster_id,omitempty" json:"cluster_id,omitempty"`
	ExternalClusterID string       `yaml:"external_cluster_id,omitempty" json:"external_cluster_id,omitempty"`
	OrganizationID    string       `yaml:"organization_id,omitempty" json:"organization_id,omitempty"`
	LastTelemetryDate string       `yaml:"last_telemetry_date,omitempty" json:"last_telemetry_date,omitempty"`
	CreatedAt         string       `yaml:"created_at,omitempty" json:"created_at,omitempty"`
	UpdatedAt         string       `yaml:"updated_at,omitempty" json:"updated_at,omitempty"`
	SupportLevel      string       `yaml:"support_level,omitempty" json:"support_level,omitempty"`
	DisplayName       string       `yaml:"display_name,omitempty" json:"display_name,omitempty"`
	Creator           StandardKind `yaml:"creator" json:"creator"`
	Managed           bool         `yaml:"managed,omitempty" json:"managed,omitempty"`
	Status            string       `yaml:"status" json:"status"`
	Provenance        string       `yaml:"provenance,omitempty" json:"provenance,omitempty"`
	LastReconcileDate string       `yaml:"last_reconcile_date,omitempty" json:"last_reconcile_date,omitempty"`
	LastReleasedAt    string       `yaml:"last_released_at,omitempty" json:"last_released_at,omitempty"`
	Reason            string       `yaml:"reason,omitempty" json:"reason,omitempty"`
}

// Console ...
type Console struct {
	URL string `yaml:"url,omitempty"`
}

// StandardKind ...
type StandardKind struct {
	Kind string `yaml:"kind,omitempty"`
	ID   string `yaml:"id,omitempty"`
	Href string `href:"kind,omitempty"`
}

// Cluster ...
type Cluster struct {
	Kind                     string                 `yaml:"kind"`
	ID                       string                 `yaml:"id"`
	Href                     string                 `yaml:"href"`
	Name                     string                 `yaml:"name"`
	ExternalID               string                 `yaml:"external_id"`
	DisplayName              string                 `yaml:"display_name"`
	CreationTimestamp        metav1.Time            `yaml:"creation_timestamp,omitempty"`
	ActivityTimestamp        metav1.Time            `yaml:"activity_timestamp,omitempty"`
	CloudProvider            StandardKind           `yaml:"cloud_provider,omitempty"`
	OpenShiftVersion         string                 `yaml:"openshift_version"`
	Subscription             StandardKind           `yaml:"subscription,omitempty"`
	Region                   StandardKind           `yaml:"region,omitempty"`
	Console                  Console                `yaml:"console,omitempty"`
	Nodes                    map[string]interface{} `yaml:"nodes,omitempty"`
	State                    string                 `yaml:"state"`
	Groups                   StandardKind           `yaml:"groups,omitempty"`
	Network                  interface{}            `yaml:"network,omitempty"`
	ExternalConfig           map[string]interface{} `yaml:"external_configuration,omitempty"`
	MultiAZ                  bool                   `yaml:"multi_az,omitempty"`
	Managed                  bool                   `yaml:"managed,omitempty"`
	BYOC                     bool                   `yaml:"byoc,omitempty"`
	CCS                      map[string]interface{} `yaml:"ccs,omitempty"`
	IdentityProviders        StandardKind           `yaml:"identity_providers,omitempty"`
	AWSInfraAccessRoleGrants map[string]interface{} `yaml:"aws_infrastructure_access_role_grants,omitempty"`
	Metrics                  map[string]interface{} `yaml:"metrics,omitempty"`
	Addons                   StandardKind           `yaml:"addons,omitempty"`
	Ingresses                StandardKind           `yaml:"ingresses,omitempty"`
	HealthState              string                 `yaml:"health_state,omitempty"`
	Product                  StandardKind           `yaml:"product,omitempty"`
	DNSReady                 bool                   `yaml:"dns_ready,omitempty"`
	Reason                   string                 `yaml:"reason,omitempty" json:"reason,omitempty"`
}

// SetupEndpoints ...
func SetupEndpoints(r *gin.Engine, logger zerolog.Logger) {
	r.GET("/api/clusters_mgmt/v1/clusters/*clusterID", GetCluster)

	r.GET("/api/accounts_mgmt/v1/subscriptions", GetSubscriptions)

	r.POST("/auth/realms/redhat-external/protocol/openid-connect/token", GetToken)
}

// GetSubscriptions ...
func GetSubscriptions(c *gin.Context) {
	file, err := ioutil.ReadFile("data/subscriptions.json")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": fmt.Sprintf("Error reading file: %s", err.Error()),
		})
		return
	}

	var subscriptionList SubscriptionList
	err = json.Unmarshal(file, &subscriptionList)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": fmt.Sprintf("Error unmarshalling JSON: %s", err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, subscriptionList)
}

// GetCluster ...
func GetCluster(c *gin.Context) {
	clusterID := c.Param("clusterID")
	params := c.Request.URL.Query()

	var file []byte
	var err error
	// Return filtered results if search param set
	if _, ok := params["search"]; ok {
		file, err = ioutil.ReadFile("data/filtered_clusters_list.json")
	} else {
		file, err = ioutil.ReadFile("data/clusters_list.json")
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": fmt.Sprintf("Error reading file: %s", err.Error()),
		})
		return
	}

	var clusterList ClusterList
	err = json.Unmarshal(file, &clusterList)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": fmt.Sprintf("Error unmarshalling JSON: %s", err.Error()),
		})
		return
	}

	if clusterID != "" {
		c.Data(http.StatusOK, "application/json", file)
	}

	for _, cluster := range clusterList.Items {
		if cluster.ID == clusterID {
			c.JSON(http.StatusOK, cluster)
			return
		}
	}

	c.Status(http.StatusBadRequest)
}

// GetToken
func GetToken(c *gin.Context) {
	token := c.PostForm("refresh_token")
	if token == "" {
		log.Println("Empty token received. Responding with auth error.")
		file, err := ioutil.ReadFile("data/auth_error.json")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"msg": fmt.Sprintf("Error reading file: %s", err.Error()),
			})
			return
		}
		c.Data(http.StatusBadRequest, "application/json", file)
	} else {
		log.Println("Auth token received. Responding with auth success.")
		file, err := ioutil.ReadFile("data/auth_success.json")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"msg": fmt.Sprintf("Error reading file: %s", err.Error()),
			})
			return
		}
		c.Data(http.StatusOK, "application/json", file)
	}
}

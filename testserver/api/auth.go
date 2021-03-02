package api

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetToken returns successfully if refresh_token is non-empty, and an auth error if empty
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

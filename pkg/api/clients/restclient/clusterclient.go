// Copyright Contributors to the Open Cluster Management project

package restclient

import (
	"net/http"
)

type clusterClient struct{}

type ClusterGetInterface interface {
	Get(*http.Request) (*http.Response, error)
}

var (
	ClusterHTTPClient ClusterGetInterface = &clusterClient{}
)

func (c *clusterClient) Get(request *http.Request) (*http.Response, error) {
	client := http.Client{}
	return client.Do(request)
}

// Copyright Contributors to the Open Cluster Management project

package cluster_provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	discoveryv1 "github.com/open-cluster-management/discovery/api/v1"
	"github.com/open-cluster-management/discovery/pkg/api/clients/restclient"
	"github.com/open-cluster-management/discovery/pkg/api/domain/cluster_domain"
)

const (
	clusterURL = "%s/api/clusters_mgmt/v1/clusters"
)

type clusterProvider struct{}

type IClusterProvider interface {
	GetClusters(request cluster_domain.ClusterRequest) (*cluster_domain.ClusterResponse, *cluster_domain.ClusterError)
}

var (
	ClusterProvider IClusterProvider = &clusterProvider{}
)

func (c *clusterProvider) GetClusters(request cluster_domain.ClusterRequest) (*cluster_domain.ClusterResponse, *cluster_domain.ClusterError) {
	getURL := fmt.Sprintf(clusterURL, request.BaseURL)
	query := &url.Values{}
	query.Add("size", fmt.Sprintf("%d", request.Size))
	query.Add("page", fmt.Sprintf("%d", request.Page))
	applyPreFilters(query, request.Filter)

	getRequest, err := http.NewRequest("GET", getURL, nil)
	getRequest.URL.RawQuery = query.Encode()
	getRequest.Header.Add("Authorization", fmt.Sprintf("Bearer %s", request.Token))
	getRequest = getRequest.WithContext(context.Background())

	response, err := restclient.ClusterHTTPClient.Get(getRequest)
	if err != nil {
		return nil, &cluster_domain.ClusterError{
			Error: err,
		}
	}
	defer response.Body.Close()

	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, &cluster_domain.ClusterError{
			Error: err,
		}
	}

	// The api owner can decide to change datatypes, etc. When this happen, it might affect the error format returned
	if response.StatusCode > 299 {
		var errResponse cluster_domain.ClusterError
		if err := json.Unmarshal(bytes, &errResponse); err != nil {
			return nil, &cluster_domain.ClusterError{
				Error:    err,
				Response: bytes}
		}

		if errResponse.Reason == "" {
			errResponse.Error = fmt.Errorf("invalid json response body")
			errResponse.Response = bytes
		}
		return nil, &errResponse
	}

	var result cluster_domain.ClusterResponse
	if err := json.Unmarshal(bytes, &result); err != nil {
		return nil, &cluster_domain.ClusterError{
			Error:    fmt.Errorf("error unmarshaling response"),
			Response: bytes,
		}
	}

	return &result, nil
}

// applyPreFilters adds fields to the http query to limit the number of items returned
func applyPreFilters(query *url.Values, filters discoveryv1.Filter) {
	if filters.LastActive != 0 {
		query.Add("search", fmt.Sprintf("activity_timestamp >= '%s'", lastActiveDate(time.Now(), filters.LastActive)))
	}
}

// return the date that is `daysAgo` days before `currentDate` in 'YYYY-MM-DD' format
func lastActiveDate(currentDate time.Time, daysAgo int) string {
	if daysAgo < 0 {
		daysAgo = 0
	}
	cutoffDay := currentDate.AddDate(0, 0, -daysAgo)
	return cutoffDay.Format("2006-01-02")
}

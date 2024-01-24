// Copyright Contributors to the Open Cluster Management project

package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	discovery "github.com/stolostron/discovery/api/v1"
)

const (
	clusterURL = "%s/api/clusters_mgmt/v1/clusters"
)

var (
	httpClient           ClusterGetInterface = &clusterRestClient{}
	SubscriptionProvider IClusterProvider    = &clusterProvider{}
)

type ClusterGetInterface interface {
	Get(*http.Request) (*http.Response, error)
}

type clusterRestClient struct{}

func (c *clusterRestClient) Get(request *http.Request) (*http.Response, error) {
	client := http.Client{}
	return client.Do(request)
}

type IClusterProvider interface {
	GetClusters(request ClusterRequest) (*ClusterResponse, *ClusterError)
}

type clusterProvider struct{}

var (
	ClusterProvider IClusterProvider = &clusterProvider{}
)

func (c *clusterProvider) GetClusters(request ClusterRequest) (retRes *ClusterResponse, retErr *ClusterError) {
	getRequest, _ := prepareRequest(request)

	response, err := httpClient.Get(getRequest)
	if err != nil {
		return nil, &ClusterError{
			Error: err,
		}
	}

	defer func() {
		err := response.Body.Close()
		if err != nil && retErr == nil {
			retErr = &ClusterError{
				Error: fmt.Errorf("%s: %w", "error closing response body", err),
			}
		}
	}()

	retRes, retErr = parseResponse(response)
	return
}

func prepareRequest(request ClusterRequest) (*http.Request, error) {
	getURL := fmt.Sprintf(clusterURL, request.BaseURL)
	query := &url.Values{}
	query.Add("size", fmt.Sprintf("%d", request.Size))
	query.Add("page", fmt.Sprintf("%d", request.Page))
	applyPreFilters(query, request.Filter)

	getRequest, err := http.NewRequest("GET", getURL, nil)
	if err != nil {
		return nil, err
	}
	getRequest.URL.RawQuery = query.Encode()
	getRequest.Header.Add("Authorization", fmt.Sprintf("Bearer %s", request.Token))
	getRequest = getRequest.WithContext(context.Background())
	return getRequest, nil
}

func parseResponse(response *http.Response) (*ClusterResponse, *ClusterError) {
	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, &ClusterError{
			Error: fmt.Errorf("%s: %w", "couldn't read response body", err),
		}
	}

	if response.StatusCode > 299 {
		var errResponse ClusterError
		if err := json.Unmarshal(bytes, &errResponse); err != nil {
			return nil, &ClusterError{
				Error:    fmt.Errorf("%s: %w", "couldn't unmarshal cluster error response", err),
				Response: bytes,
			}
		}

		if errResponse.Reason == "" {
			errResponse.Error = fmt.Errorf("unexpected json response body")
			errResponse.Response = bytes
		}
		return nil, &errResponse
	}

	var result ClusterResponse
	if err := json.Unmarshal(bytes, &result); err != nil {
		return nil, &ClusterError{
			Error:    fmt.Errorf("%s: %w", "couldn't unmarshal cluster response", err),
			Response: bytes,
		}
	}

	return &result, nil
}

// applyPreFilters adds fields to the http query to limit the number of items returned
func applyPreFilters(query *url.Values, filters discovery.Filter) {
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

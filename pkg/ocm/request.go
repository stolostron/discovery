package ocm

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	discoveryv1 "github.com/open-cluster-management/discovery/api/v1"
	"sigs.k8s.io/yaml"
)

var (
	OCMClusterPath = "https://api.openshift.com/api/clusters_mgmt/v1/clusters"
)

type OCMRequest struct {
	path   string
	token  string
	page   *int
	size   *int
	filter discoveryv1.Filter
}

func (r *OCMRequest) Page(n int) *OCMRequest {
	r.page = &n
	return r
}

func (r *OCMRequest) Size(n int) *OCMRequest {
	r.size = &n
	return r
}

func (r *OCMRequest) Token(token string) *OCMRequest {
	r.token = token
	return r
}

func (r *OCMRequest) Filter(filter discoveryv1.Filter) *OCMRequest {
	r.filter = filter
	return r
}

// Get ...
func (r *OCMRequest) Get(ctx context.Context) (ClusterList, error) {
	r.path = "https://api.openshift.com/api/clusters_mgmt/v1/clusters"
	if r.token == "" {
		return ClusterList{}, fmt.Errorf("Missing token")
	}

	query := &url.Values{}
	if r.size != nil {
		query.Add("size", fmt.Sprintf("%d", *r.size))
	}
	if r.page != nil {
		query.Add("page", fmt.Sprintf("%d", *r.page))
	}
	if r.filter.Age != 0 {
		query.Add("search", fmt.Sprintf("creation_timestamp >= '%s'", ageDate(time.Now(), r.filter.Age)))
	}

	request, err := http.NewRequest("GET", r.path, nil)
	request.URL.RawQuery = query.Encode()

	bearer := "Bearer " + r.token
	request.Header.Add("Authorization", bearer)

	if ctx != nil {
		request = request.WithContext(ctx)
	}

	// Send req using http Client
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return ClusterList{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ClusterList{}, err
	}

	m := ClusterList{}
	err = yaml.Unmarshal([]byte(body), &m)
	if err != nil {
		return ClusterList{}, err
	}

	if m.Reason != "" {
		return ClusterList{}, fmt.Errorf("Authentication Error: %s", m.Reason)
	}

	return m, nil
}

// return the date that is `daysAgo` days before `currentDate` to  in 'YYYY-MM-DD' format
func ageDate(currentDate time.Time, daysAgo int) string {
	if daysAgo < 0 {
		daysAgo = 0
	}
	cutoffDay := currentDate.AddDate(0, 0, -daysAgo)
	return cutoffDay.Format("2006-01-02")
}

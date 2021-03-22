// Copyright Contributors to the Open Cluster Management project

package restclient

import (
	"net/http"
	"net/url"
)

type authClient struct{}

type AuthPostInterface interface {
	Post(url string, data url.Values) (resp *http.Response, err error)
}

var (
	AuthHTTPClient AuthPostInterface = &authClient{}
)

func (c *authClient) Post(url string, data url.Values) (resp *http.Response, err error) {
	return http.PostForm(url, data) // #nosec G107 (url needs to be configurable to target mock servers)
}

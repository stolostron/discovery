// Copyright Contributors to the Open Cluster Management project

package common

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/stolostron/discovery/pkg/ocm/cluster"
)

var (
	// HTTPClient is used to make HTTP requests.
	httpClient HTTPRequester = &RestClient{}

	// Provider is used for loading resources.
	provider ResourceLoader = &RestProvider{}
)

// Get is used to make HTTP GET requests
func (c *RestClient) Get(request *http.Request) (*http.Response, error) {
	client := http.Client{}

	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("error making HTTP request: %w", err)
	}

	return response, nil
}

// GetResources retrieves resources based on the provided request
func (p *RestProvider) GetResources(request Request, endpointURL string) (retRes *Response, retErr *Error) {
	getRequest, err := prepareRequest(request, endpointURL)
	if err != nil {
		return nil, &Error{
			Error: fmt.Errorf("%s: %w", "error forming request", err),
		}
	}

	response, err := httpClient.Get(getRequest)
	if err != nil {
		return nil, &Error{
			Error: fmt.Errorf("%s: %w", "error during request", err),
		}
	}

	defer func() {
		err := response.Body.Close()
		if err != nil && retErr == nil {
			retErr = &Error{
				Error: fmt.Errorf("%s: %w", "error closing response body", err),
			}
		}
	}()

	return parseResponse(response)
}

// parseResponse parses the HTTP response
func parseResponse(response *http.Response) (*Response, *Error) {
	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, &Error{
			Error: fmt.Errorf("%s: %w", "couldn't read response body", err),
		}
	}

	if response.StatusCode > 299 {
		var errResponse Error
		if err := json.Unmarshal(bytes, &errResponse); err != nil {
			return nil, &Error{
				Error:    fmt.Errorf("%s: %w", "couldn't unmarshal resource error response", err),
				Response: bytes,
			}
		}

		if errResponse.Reason == "" {
			errResponse.Error = fmt.Errorf("unexpected json response body")
			errResponse.Response = bytes
		}
		return nil, &errResponse
	}

	var result Response
	if err := json.Unmarshal(bytes, &result); err != nil {
		return nil, &Error{
			Error:    fmt.Errorf("%s: %w", "couldn't unmarshal resource response", err),
			Response: bytes,
		}
	}

	return &result, nil
}

// prepareRequest prepares the HTTP request
func prepareRequest(request Request, endpointURL string) (*http.Request, error) {
	getURL := fmt.Sprintf(endpointURL, request.BaseURL)
	query := &url.Values{}
	query.Add("size", fmt.Sprintf("%d", request.Size))
	query.Add("page", fmt.Sprintf("%d", request.Page))

	var objectType = "subscription"

	if endpointURL == cluster.GetClusterURL() {
		objectType = "cluster"
	}

	applyPreFilters(query, request.Filter, objectType)

	getRequest, err := http.NewRequest("GET", getURL, nil)
	if err != nil {
		return nil, err
	}
	getRequest.URL.RawQuery = query.Encode()
	getRequest.Header.Add("Authorization", fmt.Sprintf("Bearer %s", request.Token))
	getRequest = getRequest.WithContext(context.Background())
	return getRequest, nil
}

// GetHTTPClient returns the HTTP client
func GetHTTPClient() HTTPRequester {
	return httpClient
}

// GetProvider returns the resource loader
func GetProvider() ResourceLoader {
	return provider
}

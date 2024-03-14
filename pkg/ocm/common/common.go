// Copyright Contributors to the Open Cluster Management project

package common

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func (c *RestClient) Get(request *http.Request) (*http.Response, error) {
	client := http.Client{}
	return client.Do(request)
}

func (p *Provider) GetData(request Request, urlPattern string) (interface{}, error) {
	getRequest, err := PrepareRequest(request, urlPattern)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "error forming request", err)
	}

	response, err := p.GetInterface.Get(getRequest)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "error during request", err)
	}

	defer func() {
		_ = response.Body.Close()
	}()

	return ParseResponse(response)
}

func ParseResponse(response *http.Response) (*SubscriptionResponse, *SubscriptionError) {
	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, &SubscriptionError{
			Error: fmt.Errorf("%s: %w", "couldn't read response body", err),
		}
	}

	if response.StatusCode > 299 {
		var errResponse SubscriptionError
		if err := json.Unmarshal(bytes, &errResponse); err != nil {
			return nil, &SubscriptionError{
				Error:    fmt.Errorf("%s: %w", "couldn't unmarshal subscription error response", err),
				Response: bytes,
			}
		}

		if errResponse.Reason == "" {
			errResponse.Error = fmt.Errorf("unexpected json response body")
			errResponse.Response = bytes
		}
		return nil, &errResponse
	}

	var result SubscriptionResponse
	if err := json.Unmarshal(bytes, &result); err != nil {
		return nil, &SubscriptionError{
			Error:    fmt.Errorf("%s: %w", "couldn't unmarshal subscription response", err),
			Response: bytes,
		}
	}

	return &result, nil
}

func PrepareRequest(request Request, urlPattern string) (*http.Request, error) {
	getURL := fmt.Sprintf(urlPattern, request.BaseURL)
	query := &url.Values{}
	query.Add("size", fmt.Sprintf("%d", request.Size))
	query.Add("page", fmt.Sprintf("%d", request.Page))

	ApplyPreFilters(query, request.Filter)

	getRequest, err := http.NewRequest("GET", getURL, nil)
	if err != nil {
		return nil, err
	}
	getRequest.URL.RawQuery = query.Encode()
	getRequest.Header.Add("Authorization", fmt.Sprintf("Bearer %s", request.Token))
	getRequest = getRequest.WithContext(context.Background())
	return getRequest, nil
}

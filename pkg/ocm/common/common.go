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

func (p *Provider) GetResources(request Request, endpointURL string) (retRes *Response, retErr *ErrorResponse) {
	getRequest, err := PrepareRequest(request, endpointURL)
	if err != nil {
		return nil, &ErrorResponse{
			Error: fmt.Errorf("%s: %w", "error forming request", err),
		}
	}

	response, err := httpClient.Get(getRequest)
	if err != nil {
		return nil, &ErrorResponse{
			Error: fmt.Errorf("%s: %w", "error during request", err),
		}
	}

	defer func() {
		err := response.Body.Close()
		if err != nil && retErr == nil {
			retErr = &ErrorResponse{
				Error: fmt.Errorf("%s: %w", "error closing response body", err),
			}
		}
	}()

	retRes, retErr = ParseResponse(response)
	return
}

func ParseResponse(response *http.Response) (*Response, *ErrorResponse) {
	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, &ErrorResponse{
			Error: fmt.Errorf("%s: %w", "couldn't read response body", err),
		}
	}

	if response.StatusCode > 299 {
		var errResponse ErrorResponse
		if err := json.Unmarshal(bytes, &errResponse); err != nil {
			return nil, &ErrorResponse{
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
		return nil, &ErrorResponse{
			Error:    fmt.Errorf("%s: %w", "couldn't unmarshal resource response", err),
			Response: bytes,
		}
	}

	return &result, nil
}

// PrepareRequest prepares the HTTP request
func PrepareRequest(request Request, endpointURL string) (*http.Request, error) {
	getURL := fmt.Sprintf(endpointURL, request.BaseURL)
	query := &url.Values{}
	query.Add("size", fmt.Sprintf("%d", request.Size))
	query.Add("page", fmt.Sprintf("%d", request.Page))
	// applyPreFilters(query, request.Filter)

	getRequest, err := http.NewRequest("GET", getURL, nil)
	if err != nil {
		return nil, err
	}
	getRequest.URL.RawQuery = query.Encode()
	getRequest.Header.Add("Authorization", fmt.Sprintf("Bearer %s", request.Token))
	getRequest = getRequest.WithContext(context.Background())
	return getRequest, nil
}

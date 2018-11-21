package giniapi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
)

// MakeAPIRequest is a wrapper around http.NewRequest to create http
// request and inject required headers, set timeout, ...
func (api *APIClient) makeAPIRequest(ctx context.Context, verb, url string, body io.Reader, headers map[string]string, userIdentifier string) (*http.Response, error) {
	req, err := http.NewRequest(verb, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %s", err)
	}

	if _, ok := headers["Accept"]; !ok {
		req.Header.Add("Accept", fmt.Sprintf("application/vnd.gini.%s+json", api.Config.APIVersion))
	}

	req.Header.Add("User-Agent", fmt.Sprintf("gini-api-go/%s", VERSION))

	if reflect.TypeOf(api.Config.Authentication).Name() == "BasicAuth" {
		if userIdentifier == "" {
			return nil, fmt.Errorf("userIdentifier required (Authentication=BasicAuth)")
		}
		req.Header.Add("X-User-Identifier", userIdentifier)
	}

	// Append additional headers
	for h, v := range headers {
		req.Header.Add(h, v)
	}

	// Add context to request
	req = req.WithContext(ctx)

	resp, err := api.HTTPClient.Do(req)

	return resp, err
}

// apiResponse combines a HTTP response, error object and additional data
// into a ApiResponse object
func apiResponse(message, docId string, response *http.Response, error error) APIResponse {
	r := APIResponse{
		Message: message,
		DocumentId: docId,
		Error: error,
		HttpResponse: response,
	}

	if response != nil {
		r.RequestId = response.Header.Get("X-Request-Id")
	}

	return r
}

func encodeURLParams(baseURL string, queryParams map[string]interface{}) string {
	u, _ := url.Parse(baseURL)

	params := url.Values{}

	for key, value := range queryParams {
		switch value := value.(type) {
		case string:
			params.Add(key, value)
		case int:
			params.Add(key, strconv.Itoa(value))
		}
	}

	u.RawQuery = params.Encode()
	return u.String()
}

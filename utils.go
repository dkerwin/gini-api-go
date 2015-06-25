package giniapi

import (
	"fmt"
	"io"
	"net/http"
)

// Wrapper around http.NewRequest to inject headers etc.
func (api *APIClient) MakeAPIRequest(verb string, url string, body io.Reader, headers map[string]string, userIdentifier string) (*http.Response, error) {
	req, err := http.NewRequest(verb, url, body)
	if err != nil {
		return nil, fmt.Errorf("Failed to create HTTP request: %s", err)
	}
	req.Header.Add("Accept", fmt.Sprintf("application/vnd.gini.%s+json", api.Config.APIVersion))
	req.Header.Add("User-Agent", fmt.Sprintf("gini-api-go/%s", VERSION))

	if api.Config.Authentication == "basicAuth" {
		if userIdentifier == "" {
			return nil, fmt.Errorf("userIdentifier required (Authentication=basicAuth)")
		}
		req.Header.Add("X-User-Identifier", userIdentifier)
	}

	// Append additional headers
	for h, v := range headers {
		req.Header.Add(h, v)
	}

	resp, err := api.HTTPClient.Do(req)
	return resp, err
}

package giniapi

import (
	"errors"
	"fmt"
	"io"
	"net/http"
)

// Wrapper around http.NewRequest to inject headers etc.
func (api *APIClient) MakeAPIRequest(verb string, url string, body io.Reader, headers map[string]string, userIdentifier string) (*http.Response, error) {
	req, err := http.NewRequest(verb, url, body)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Accept", fmt.Sprintf("application/vnd.gini.%s+json", api.Config.APIVersion))
	req.Header.Add("User-Agent", fmt.Sprintf("gini-api-go/%s", VERSION))

	if api.Config.Authentication == "enterprise" {
		if userIdentifier != "" {
			return nil, errors.New("userIdentifier required (Authentication=enterprise)")
		} else {
			req.Header.Add("X-User-Identifier", userIdentifier)
		}
	}

	// Append additional headers
	for h, v := range headers {
		req.Header.Add(h, v)
	}

	// log.Printf("My request: %#v\n", req)

	resp, err := api.HTTPClient.Do(req)

	return resp, err
}

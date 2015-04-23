package giniapi

import (
	"io"
	"net/http"
)

// Wrapper around http.NewRequest to inject headers etc.
func (api *APIClient) MakeAPIRequest(verb string, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(verb, url, body)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Accept", "application/vnd.gini.v1+json")
	req.Header.Add("User-Agent", "gini-api-go/"+VERSION)

	resp, err := api.HTTPClient.Do(req)

	return resp, err
}

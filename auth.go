package giniapi

import (
	"fmt"
	"golang.org/x/oauth2"
	"net/http"
)

// Create custom http.Client for oauth2 or enterprise auth
func NewHttpClient(config *Config) (*http.Client, error) {
	if config.Authentication == "oauth2" {
		// Setup oauth2
		conf := &oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			Scopes:       []string{"write"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  config.Endpoints.UserCenter + "/oauth/authorize",
				TokenURL: config.Endpoints.UserCenter + "/oauth/token",
			},
		}

		if config.AuthCode != "" {
			token, err := conf.Exchange(oauth2.NoContext, config.AuthCode)
			if err != nil {
				return nil, fmt.Errorf("Failed to exchange auth code: %s", err)
			}
			client := conf.Client(oauth2.NoContext, token)
			return client, nil

		} else if config.Username != "" && config.Password != "" {
			token, err := conf.PasswordCredentialsToken(oauth2.NoContext, config.Username, config.Password)
			if err != nil {
				return nil, err
			}
			client := conf.Client(oauth2.NoContext, token)
			return client, nil
		} else {
			return nil, fmt.Errorf("Not enough parameters for oauth2")
		}
	} else if config.Authentication == "basicAuth" {
		client := &http.Client{Transport: BasicAuthTransport{Config: config}}
		return client, nil
	}
	return &http.Client{}, nil
}

// Custom net/http transport to add Authorization headers
// for Gini API's basic auth system
type BasicAuthTransport struct {
	Transport http.RoundTripper
	Config    *Config
}

// RoundTrip to add basic auth header to all requests
func (bat BasicAuthTransport) RoundTrip(r *http.Request) (res *http.Response, err error) {
	r.SetBasicAuth(bat.Config.ClientID, bat.Config.ClientSecret)

	t := bat.Transport
	if t == nil {
		t = http.DefaultTransport
	}

	res, err = t.RoundTrip(r)
	return
}

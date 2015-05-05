package giniapi

import (
	"fmt"
	"golang.org/x/oauth2"
	"log"
	"net/http"
)

// Create custom http.Client for oauth2 or enterprise auth
func NewHttpClient(config *Config) (client *http.Client) {
	if config.Authentication == "oauth2" {
		if config.AuthCode != "" {
			fmt.Println("To be implemented...")
		} else if config.Username != "" && config.Password != "" {
			// Create Oauth2 client
			conf := &oauth2.Config{
				ClientID:     config.ClientID,
				ClientSecret: config.ClientSecret,
				Scopes:       []string{"write"},
				Endpoint: oauth2.Endpoint{
					AuthURL:  "https://user.gini.net/oauth/authorize",
					TokenURL: "https://user.gini.net/oauth/token",
				},
			}

			token, err := conf.PasswordCredentialsToken(oauth2.NoContext, config.Username, config.Password)
			if err != nil {
				fmt.Println("Password exchange failed: ", err)
			}

			client := conf.Client(oauth2.NoContext, token)
			return client
		} else {
			log.Fatal("Not enough parameters for oauth2")
		}
	} else if config.Authentication == "enterprise" {
		client := &http.Client{Transport: EnterpriseTransport{Config: config}}
		return client
	}
	return &http.Client{}
}

// Custom net/http transport to add basic auth headers
// for Gini API's enterprise system
type EnterpriseTransport struct {
	Transport http.RoundTripper
	Config    *Config
}

func (et EnterpriseTransport) RoundTrip(r *http.Request) (res *http.Response, err error) {
	r.SetBasicAuth(et.Config.ClientID, et.Config.ClientSecret)

	t := et.Transport
	if t == nil {
		t = http.DefaultTransport
	}

	res, err = t.RoundTrip(r)
	return
}

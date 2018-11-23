package giniapi

import (
	"testing"
)

func Test_newHTTPClient(t *testing.T) {
	// Basic config
	config := Config{
		ClientID:     "testclient",
		ClientSecret: "secret",
		Endpoints: Endpoints{
			API:        testHTTPServer.URL,
			UserCenter: testHTTPServer.URL,
		},
	}

	// basicAuth
	config.Authentication = UseBasicAuth
	if client, resp := newHTTPClient(&config); client == nil || resp.Error != nil {
		t.Errorf("Failed to create http client: %s", resp.Error)
	}

	// oauth2
	config.Authentication = UseOauth2

	// AuthCode
	config.AuthCode = "123456"
	if client, resp := newHTTPClient(&config); client == nil || resp.Error != nil {
		t.Errorf("Failed to exchange auth code: %s", resp.Error)
	}

	// Username + Password
	config.AuthCode = ""
	config.Username = "user1"
	config.Password = "secret"
	if client, resp := newHTTPClient(&config); client == nil || resp.Error != nil {
		t.Errorf("Failed to exchange username and password: %s", resp.Error)
	}

	// missing auth_code and user credentials
	config.AuthCode = ""
	config.Username = ""
	config.Password = ""
	if client, resp := newHTTPClient(&config); client != nil || resp.Error == nil {
		t.Errorf("Invalid oauth2 auth parameters shoulfd raise err: %s", resp.Error)
	}
}

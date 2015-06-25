package giniapi

import (
	"testing"
)

func Test_NewHttpClient(t *testing.T) {
	// Basic config
	config := Config{
		ClientID:     "testclient",
		ClientSecret: "secret",
		Endpoints: Endpoints{
			API:        testHTTPServer.URL,
			UserCenter: testHTTPServer.URL,
		},
	}

	// invalid
	config.Authentication = "unknown"
	if client, err := NewHttpClient(&config); client != nil || err == nil {
		t.Errorf("Unknown Authentication should not return http client: %s", err)
	}

	// basicAuth
	config.Authentication = "basicAuth"
	if client, err := NewHttpClient(&config); client == nil || err != nil {
		t.Errorf("Failed to create http client: %s", err)
	}

	// oauth2
	config.Authentication = "oauth2"

	// AuthCode
	config.AuthCode = "123456"
	if client, err := NewHttpClient(&config); client == nil || err != nil {
		t.Errorf("Failed to exchange auth code: %s", err)
	}

	// Username + Password
	config.AuthCode = ""
	config.Username = "user1"
	config.Password = "secret"
	if client, err := NewHttpClient(&config); client == nil || err != nil {
		t.Errorf("Failed to exchange username and password: %s", err)
	}

	// missing auth_code and user credentials
	config.AuthCode = ""
	config.Username = ""
	config.Password = ""
	if client, err := NewHttpClient(&config); client != nil || err == nil {
		t.Errorf("Invalid oauth2 auth parameters shoulfd raise err: %s", err)
	}
}

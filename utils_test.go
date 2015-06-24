package giniapi

import (
	// "fmt"
	"io/ioutil"
	"testing"
)

func Test_MakeAPIRequest(t *testing.T) {
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
	config.Authentication = "basicAuth"
	api, err := NewClient(&config)
	if err != nil {
		t.Errorf("Failed to setup NewClient: %s", err)
	}

	// Fail without userIdentifier
	if response, err := api.MakeAPIRequest("GET", testHTTPServer.URL+"/test/http/basicAuth", nil, nil, ""); response != nil || err == nil {
		t.Errorf("Missing userIdentifier should raise err")
	}

	// Succeed with userIdentifier
	response, err := api.MakeAPIRequest("GET", testHTTPServer.URL+"/test/http/basicAuth", nil, nil, "user123")
	if response == nil || err != nil {
		t.Errorf("HTTP call with supplied userIdentifier failed: %s", err)
	}

	body, _ := ioutil.ReadAll(response.Body)
	if response.StatusCode != 200 || string(body) != "test completed" {
		t.Errorf("Body (%s) or statusCode(%d) mismatch", string(body), response.StatusCode)
	}

	// oauth2
	config.Authentication = "oauth2"
	config.AuthCode = "123456"

	api, err = NewClient(&config)
	if err != nil {
		t.Errorf("Failed to setup NewClient: %s", err)
	}

	// Make oauth2 call
	if response, err := api.MakeAPIRequest("GET", testHTTPServer.URL+"/test/http/oauth2", nil, nil, ""); response == nil || err != nil {
		t.Errorf("Call failed: %#v", err)
	}
}

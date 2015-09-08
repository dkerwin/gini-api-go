package giniapi

import (
	"testing"
	"time"
)

func Test_UploadOptionsTimeout(t *testing.T) {
	u := UploadOptions{}
	assertEqual(t, u.Timeout(), 30*time.Second, "")

	u.PollTimeout = 1 * time.Second
	assertEqual(t, u.Timeout(), 1*time.Second, "")
}

func Test_ConfigVerify(t *testing.T) {
	c := Config{}

	// Empty config fails
	assertNotEqual(t, c.Verify(), nil, "")

	// Minimal Oauth2 config
	c.ClientID = "client"
	c.ClientSecret = "secret"
	c.Authentication = UseOauth2

	// Fail without auth_code || username & password
	assertNotEqual(t, c.Verify(), nil, "")

	c.Username = "user1"
	assertNotEqual(t, c.Verify(), nil, "")

	c.AuthCode = "12345"
	assertEqual(t, c.Verify(), nil, "")

	c.Password = "secret"
	assertEqual(t, c.Verify(), nil, "")

	// Verify defaults
	c.Verify()

	assertEqual(t, c.APIVersion, "v1", "")
	assertEqual(t, c.Endpoints.API, "https://api.gini.net", "")
	assertEqual(t, c.Endpoints.UserCenter, "https://user.gini.net", "")
}

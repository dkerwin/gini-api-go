// Copyright 2015-2018 The gini-api-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package giniapi interacts with Gini's API service to make sense of unstructured
documents. Please visit http://developer.gini.net/gini-api/html/index.html
for more details about the Gini API and it's capabilities.

API features

Supported API calls include:

	- Upload documents (native, scanned, text)
	- List a users documents
	- Search documents
	- Get extractions (incubator is supported)
	- Download rendered pages, processed document and layout XML
	- Submit feedback on extractions
	- Submit error reports

Contributing

It's awesome that you consider contributing to gini-api-go. Here are the 5 easy steps you should follow:

	- Fork repository on Github
	- Create a topic/feature branch
	- Write code AND tests
	- Update documentation if necessary
	- Open a pull request

*/
package giniapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"time"
)

const (
	// VERSION is the API client version
	VERSION                   = "1.0.0"
	ErrConfigInvalid          = "failed to initialize config object"
	ErrMissingCredentials     = "username or password cannot be empty in Oauth2 flow"
	ErrOauthAuthCodeExchange  = "failed to exchange oauth2 auth code"
	ErrOauthCredentials       = "failed to obtain token with username/password"
	ErrOauthParametersMissing = "oauth2 authentication requires AuthCode or Username + Password"
	ErrUploadFailed           = "failed to upload document"
	ErrDocumentGet            = "failed to GET document object"
	ErrDocumentParse          = "failed to parse document json"
	ErrDocumentRead           = "failed to read document body"
	ErrDocumentList           = "failed to get document list"
	ErrDocumentTimeout        = "failed to process document in time"
	ErrDocumentProcessing     = "failed to process document"
	ErrDocumentDelete         = "failed to delete document"
	ErrDocumentLayout         = "failed to retrieve layout"
	ErrDocumentExtractions    = "failed to retrieve extractions"
	ErrDocumentProcessed      = "failed to retrieve processed document"
	ErrDocumentFeedback       = "failed to submit feedback"
	ErrHTTPPostFailed         = "failed to complete POST request"
	ErrHTTPGetFailed          = "failed to complete GET request"
	ErrHTTPDeleteFailed       = "failed to complete DELETE request"
	ErrHTTPPutFailed          = "failed to complete PUT request"
)

// Config to setup Gini API connection
type Config struct {
	// ClientID is the application's ID.
	ClientID string
	// ClientSecret is the application's secret.
	ClientSecret string
	// Username for oauth2 password grant
	Username string
	// Password for oauth2 pssword grant
	Password string
	// Auth_code to exchange for oauth2 token
	AuthCode string
	// Scopes to use (leave empty for all assigned scopes)
	Scopes []string
	// API & Usercenter endpoints
	Endpoints
	// APIVersion to use (v1)
	APIVersion string `default:"v1"`
	// Authentication to use
	// oauth2: auth_code || password credentials
	// basicAuth: basic auth + user identifier
	Authentication APIAuthScheme
}

func (c *Config) Verify() error {
	if c.ClientID == "" || c.ClientSecret == "" {
		return errors.New(ErrConfigInvalid)
	}

	if reflect.TypeOf(c.Authentication).Name() == "Oauth2" {
		if c.AuthCode == "" && (c.Username == "" || c.Password == "") {
			return errors.New(ErrMissingCredentials)
		}
	}

	cType := reflect.TypeOf(*c)

	// Fix potentially missing APIVersion with default
	if c.APIVersion == "" {
		f, _ := cType.FieldByName("APIVersion")
		c.APIVersion = f.Tag.Get("default")
	}

	// Fix potential missing Endpoints with defaults
	cType = reflect.TypeOf(c.Endpoints)

	if c.Endpoints.API == "" {
		f, _ := cType.FieldByName("API")
		c.Endpoints.API = f.Tag.Get("default")
	}
	if c.Endpoints.UserCenter == "" {
		f, _ := cType.FieldByName("UserCenter")
		c.Endpoints.UserCenter = f.Tag.Get("default")
	}

	return nil
}

// APIResponse will transport about the request back to the caller
type APIResponse struct {
	// Error: stores error object encountered on the way
	Error error
	// Message: error message with more context
	Message string
	// DocumentId: internal Id of the document. Can be empty
	DocumentId string
	// RequestId: request Id returned for the request to the gini API
	RequestId string
	// HttpResponse: full response object
	HttpResponse *http.Response
}

// Endpoints to access API and Usercenter
type Endpoints struct {
	API        string `default:"https://api.gini.net"`
	UserCenter string `default:"https://user.gini.net"`
}

// UploadOptions specify parameters to the Upload function
type UploadOptions struct {
	FileName       string
	DocType        string
	UserIdentifier string
}

// ListOptions specify parameters to the List function
type ListOptions struct {
	Limit          int
	Offset         int
	UserIdentifier string
}

// APIClient is the main interface for the user
type APIClient struct {
	// Config
	Config

	// Http client
	HTTPClient *http.Client
}

// NewClient validates your Config parameters and returns a APIClient object
// with a matching http client included.
func NewClient(config *Config) (*APIClient, error) {
	if err := config.Verify(); err != nil {
		return nil, err
	}

	// Get http client based on the selected authentication scheme
	client, resp := newHTTPClient(config)
	if resp.Error != nil {
		return nil, resp.Error
	}

	return &APIClient{
		Config:     *config,
		HTTPClient: client,
	}, nil

}

// Upload a document from a given io.Reader object (document). Additional options can be
// passed with a instance of UploadOptions. FileName and DocType are optional and can be empty.
// UserIdentifier is required if Authentication method is "basic_auth".
// Upload time is measured and stored in Timing struct (part of Document).
func (api *APIClient) Upload(ctx context.Context, document io.Reader, options UploadOptions) (*Document, APIResponse) {
	start := time.Now()

	resp, err := api.makeAPIRequest(ctx, "POST", fmt.Sprintf("%s/documents", api.Config.Endpoints.API), document, nil, options.UserIdentifier)

	if err != nil {
		return nil, apiResponse(ErrHTTPPostFailed, "", resp, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, apiResponse(ErrUploadFailed, "", resp, errors.New(ErrUploadFailed))
	}

	uploadDuration := time.Since(start)

	// Fetch the document
	doc, response := api.Get(ctx, resp.Header.Get("Location"), options.UserIdentifier)

	if err != nil {
		return nil, response
	}

	// Add upload timer to document
	doc.Timing.Upload = uploadDuration

	return doc, apiResponse("document upload completed", doc.ID, resp, err)
}

// Get Document struct from URL
func (api *APIClient) Get(ctx context.Context, url, userIdentifier string) (*Document, APIResponse) {
	resp, err := api.makeAPIRequest(ctx, "GET", url, nil, nil, userIdentifier)

	if err != nil {
		return nil, apiResponse(ErrHTTPGetFailed, "", resp, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, apiResponse(ErrDocumentGet, "", resp, errors.New(ErrDocumentGet))
	}

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, apiResponse(ErrDocumentRead, "", resp, err)
	}

	var doc Document
	if err := json.Unmarshal(contents, &doc); err != nil {
		return nil, apiResponse(ErrDocumentParse, "", resp, err)
	}

	// Add client and owner to doc object
	doc.client = api
	doc.Owner = userIdentifier

	return &doc, apiResponse("document fetch completed", doc.ID, resp, err)
}

// List returns DocumentSet
func (api *APIClient) List(ctx context.Context, options ListOptions) (*DocumentSet, APIResponse) {
	params := map[string]interface{}{
		"limit":  options.Limit,
		"offset": options.Offset,
	}

	u := encodeURLParams(fmt.Sprintf("%s/documents", api.Config.Endpoints.API), params)

	resp, err := api.makeAPIRequest(ctx, "GET", u, nil, nil, options.UserIdentifier)

	if err != nil {
		return nil, apiResponse(ErrHTTPGetFailed, "", resp, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, apiResponse(ErrDocumentList, "", resp, errors.New(ErrDocumentList))
	}

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, apiResponse(ErrDocumentRead, "", resp, err)
	}

	var docs DocumentSet
	if err := json.Unmarshal(contents, &docs); err != nil {
		return nil, apiResponse(ErrDocumentParse, "", resp, err)
	}

	// Extra round: Ingesting *APIClient into each and every doc
	for _, d := range docs.Documents {
		d.client = api
		d.Owner = options.UserIdentifier
	}

	return &docs, apiResponse("document list completed", "", resp, err)
}

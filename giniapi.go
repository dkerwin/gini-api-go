// Copyright 2015 The giniapi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package giniapit interacts with Gini's API service to make sense of unstructured
documents. Please visit http://developer.gini.net/gini-api/html/index.html
for more details about the Gini API.

API features

Suppoted API calls include:

	- Upload documents (native, scanned, text)
	- List a users documents
	- Search documents
	- Get extractions (incubator is supported)
	- Download rendered pages, processed document and layout XML
	- Submit feedback on extractions
	- Submit error reports

Contributing

It's awesome that you consider contributing to gini-api-go. Here's how it's done:

	- Fork repository on Github
	- Create a topic/feature branch
	- Write code AND tests
	- Update documentation if necessary
	- Open a pull request

*/
package giniapi

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"time"
)

const (
	VERSION string = "0.1.0"
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
	Authentication string `default:"oauth2"`
}

// Endpoints to access API and Usercenter
type Endpoints struct {
	API        string `default:"https://api.gini.net"`
	UserCenter string `default:"https://user.gini.net"`
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
	cType := reflect.TypeOf(*config)

	// Fix potential missing APIVersion with default
	if config.APIVersion == "" {
		f, _ := cType.FieldByName("APIVersion")
		config.APIVersion = f.Tag.Get("default")
	}

	// Fix potential missing Authentication with default
	if config.Authentication == "" {
		f, _ := cType.FieldByName("Authentication")
		config.APIVersion = f.Tag.Get("default")
	}

	// Fix potential missing Endpoints with defaults
	cType = reflect.TypeOf(config.Endpoints)

	if config.Endpoints.API == "" {
		f, _ := cType.FieldByName("API")
		config.Endpoints.API = f.Tag.Get("default")
	}
	if config.Endpoints.UserCenter == "" {
		f, _ := cType.FieldByName("UserCenter")
		config.Endpoints.UserCenter = f.Tag.Get("default")
	}

	// Get http client based on the selected Authentication
	client, err := NewHTTPClient(config)
	if err != nil {
		return nil, err
	}

	return &APIClient{
		Config:     *config,
		HTTPClient: client,
	}, nil

}

// Upload a document from a given io.Reader (bodyBuf). fileName and docType are not mandatory
// and can be empty. userIdentifier is required when Authentication method is "basic_auth".
// Upload time is measured and stored in Timing struct (part of Document).
func (api *APIClient) Upload(bodyBuf io.Reader, fileName, docType, userIdentifier string, pollTimeoutSec int32) (*Document, error) {
	start := time.Now()
	resp, err := api.MakeAPIRequest("POST", fmt.Sprintf("%s/documents", api.Config.Endpoints.API), bodyBuf, nil, userIdentifier)
	if err != nil {
		return nil, fmt.Errorf("Document upload failed: %s", err)
	}
	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("Upload failed with HTTP status code %d", resp.StatusCode)
	}
	uploadDuration := time.Since(start)

	doc, _ := api.Get(resp.Header["Location"][0], userIdentifier)
	if err != nil {
		return nil, err
	}
	doc.Timing.Upload = uploadDuration

	// Poll for completion or failure with timeout
	err = doc.Poll(time.Duration(pollTimeoutSec) * time.Second)

	return doc, err
}

// Get Document struct from URL
func (api *APIClient) Get(url, userIdentifier string) (*Document, error) {
	resp, err := api.MakeAPIRequest("GET", url, nil, nil, userIdentifier)
	if err != nil {
		return nil, fmt.Errorf("Failed to get document %s: %s", url, err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Failed to get document %s: HTTP status code %s", url, resp.StatusCode)
	}

	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read document body: %s", err)
	}

	var doc Document
	err = json.Unmarshal(contents, &doc)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse document json: %s", err)
	}

	// Add client and owner to doc object
	doc.client = api
	doc.Owner = userIdentifier

	return &doc, nil
}

// ListDocuments returns DocumentSet
func (api *APIClient) List(p *ListParams) DocumentSet {
	u := fmt.Sprintf("%s/documents?limit=%d&offset=%d",
		api.Config.Endpoints.API,
		p.Limit,
		p.Offset)

	resp, err := api.MakeAPIRequest("GET", u, nil, nil, "")
	if resp.StatusCode != http.StatusOK {
		log.Fatal(resp.Status)
	}
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var docs DocumentSet
	err = json.Unmarshal(contents, &docs)
	if err != nil {
		log.Fatal(err)
	}

	// Extra round: Ingesting *APIClient into each and every doc
	for _, d := range docs.Documents {
		d.client = api
	}

	return docs
}

// ListDocuments returns DocumentSet
func (api *APIClient) Search(p *SearchParams) DocumentSet {
	u := fmt.Sprintf("%s/search?q=%s&type=%slimit=%d&next=%d",
		api.Config.Endpoints.API,
		p.Query,
		p.Type,
		p.Limit,
		p.Offset)

	resp, err := api.MakeAPIRequest("GET", u, nil, nil, "")
	if resp.StatusCode != http.StatusOK {
		log.Fatal(resp.Status)
	}
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var docs DocumentSet
	err = json.Unmarshal(contents, &docs)
	if err != nil {
		log.Fatal(err)
	}

	// Extra round: Ingesting *APIClient into each and every doc
	for _, d := range docs.Documents {
		d.client = api
	}

	return docs
}

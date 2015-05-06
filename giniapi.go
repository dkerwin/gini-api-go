// Copyright 2015 The gini-api Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package gini-api provides support for making
// requests to the Gini API (https://api.gini.net)
package giniapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"time"
)

const (
	VERSION string = "0.0.1"
)

// New API Client
func NewClient(config *Config) (*APIClient, error) {
	typ := reflect.TypeOf(*config)

	// Fix potential missing APIVersion with default
	if config.APIVersion == "" {
		f, _ := typ.FieldByName("APIVersion")
		config.APIVersion = f.Tag.Get("default")
	}

	// Fix potential missing Authentication with default
	if config.Authentication == "" {
		f, _ := typ.FieldByName("Authentication")
		config.APIVersion = f.Tag.Get("default")
	}

	// Fix potential missing Endpoints with defaults
	typ = reflect.TypeOf(config.Endpoints)

	if config.Endpoints.API == "" {
		f, _ := typ.FieldByName("API")
		config.Endpoints.API = f.Tag.Get("default")
	}
	if config.Endpoints.UserCenter == "" {
		f, _ := typ.FieldByName("UserCenter")
		config.Endpoints.UserCenter = f.Tag.Get("default")
	}

	client, err := NewHttpClient(config)
	if err != nil {
		return nil, err
	}

	return &APIClient{
		Config:     *config,
		HTTPClient: client,
	}, nil

}

// Upload document (read from local file)
func (api *APIClient) Upload(bodyBuf io.Reader, doctype string, userIdentifier string) (*Document, error) {
	start := time.Now()
	resp, err := api.MakeAPIRequest("POST", "https://api.gini.net/documents", bodyBuf, userIdentifier, nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusCreated {
		return nil, errors.New(fmt.Sprintf("Invalid HTTP status code: %s", resp.StatusCode))
	}
	uploadDuration := time.Since(start)

	doc := api.Get(resp.Header["Location"][0])
	doc.Timing.Upload = uploadDuration

	doc.Poll(10)

	return &doc, nil
}

// Get Document struct from URL
func (api *APIClient) Get(url string) Document {
	resp, err := api.MakeAPIRequest("GET", url, nil, "", nil)
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

	var doc Document
	err = json.Unmarshal(contents, &doc)
	if err != nil {
		log.Fatal(err)
	}

	doc.Client = api

	return doc
}

// ListDocuments returns DocumentSet
func (api *APIClient) List(p *ListParams) DocumentSet {
	u := fmt.Sprintf("%s/documents?limit=%d&offset=%d",
		api.Config.Endpoints.API,
		p.Limit,
		p.Offset)

	resp, err := api.MakeAPIRequest("GET", u, nil, "", nil)
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
		d.Client = api
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

	resp, err := api.MakeAPIRequest("GET", u, nil, "", nil)
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
		d.Client = api
	}

	return docs
}

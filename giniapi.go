// Copyright 2015 The gini-api Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package gini-api provides support for making
// requests to the Gini API (https://api.gini.net)
package giniapi

import (
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
)

const (
	VERSION string = "0.0.1"
)

// Upload non-form
func (api *APIClient) Upload(filename string) Document {
	bodyBuf, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
	}

	resp, err := api.MakeAPIRequest("POST", "https://api.gini.net/documents", bodyBuf)

	if resp.StatusCode != http.StatusCreated {
		log.Fatal(resp.Status)
	}
	if err != nil {
		log.Fatal(err)
	}

	doc := api.Get(resp.Header["Location"][0])
	doc.Poll(10)

	return doc
}

// Get Document struct from URL
func (api *APIClient) Get(url string) Document {
	resp, err := api.MakeAPIRequest("GET", url, nil)
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

	resp, err := api.MakeAPIRequest("GET", u, nil)
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

	resp, err := api.MakeAPIRequest("GET", u, nil)
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

func NewAPIClient(config *Config) (*APIClient, error) {
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

			return &APIClient{
				Config:     *config,
				HTTPClient: client,
			}, nil
		} else {
			log.Fatal("Not enough parameters for oauth2")
		}
	} else if config.Authentication == "enterprise" {
		fmt.Println("Some work to do...")
	} else {
		fmt.Printf("config: %#v\n", config)

		return &APIClient{
			Config:     *config,
			HTTPClient: &http.Client{},
		}, nil
	}

	return nil, nil
}

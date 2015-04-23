package giniapi

import (
	"net/http"
)

// API configuration
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

	// API & Usercenter endpoints
	Endpoints

	// APIVersion to use (v1)
	APIVersion string `default:"v1"`

	// Authentication to use
	// oauth2: auth_code, password credentials
	// enterprise: basic auth + user identifier
	Authentication string `default:"oauth2"`
}

type Endpoints struct {
	API        string `default:"https://api.gini.net"`
	UserCenter string `default:"https://user.gini.net"`
}

// Client struct
type APIClient struct {
	// Config
	Config

	// Http client
	HTTPClient *http.Client
}

// APIError struct
type APIError struct {
	// StatusCode
	StatusCode int

	// APIMessage
	Message string

	// APIRequestId
	RequestId string

	// DocumentId
	DocumentId string
}

// Document struct
type Document struct {
	Client *APIClient
	Links  struct {
		Document    string `json:"document"`
		Extractions string `json:"extractions"`
		Layout      string `json:"layout"`
		Processed   string `json:"processed"`
	} `json:"_links"`
	CreationDate         int    `json:"creationDate"`
	ID                   string `json:"id"`
	Name                 string `json:"name"`
	Origin               string `json:"origin"`
	PageCount            int    `json:"pageCount"`
	Pages                []Page `json:"pages"`
	Progress             string `json:"progress"`
	SourceClassification string `json:"sourceClassification"`
}

// Page struct
type Page struct {
	Images     map[string]string `json:"images"`
	PageNumber int               `json:"pageNumber"`
}

// Box struct
type Box struct {
	Height float64 `json:"height"`
	Left   float64 `json:"left"`
	Page   int     `json:"page"`
	Top    float64 `json:"top"`
	Width  float64 `json:"width"`
}

// Extraction struct
type Extraction struct {
	Box        `json:"box"`
	Candidates string `json:"candidates"`
	Entity     string `json:"entity"`
	Value      string `json:"value"`
}

// Document extractions struct
type DocExtractions struct {
	Candidates  map[string]Extraction `json:"extractions"`
	Extractions map[string]Extraction `json:"extractions"`
}

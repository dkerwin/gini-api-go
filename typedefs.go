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

// DocumentSet list of documents
type DocumentSet struct {
	TotalCount int         `json:"totalCount"`
	Documents  []*Document `json:"documents"`
}

// Page struct
type Page struct {
	Images     map[string]string `json:"images"`
	PageNumber int               `json:"pageNumber"`
}

////////////////////////////////////////////////////
// Page Layout structs
////////////////////////////////////////////////////
type Layout struct {
	Pages []PageLayout
}

type PageLayout struct {
	Number    int
	SizeX     float64
	SizeY     float64
	TextZones []TextZone
	Regions   []Region
}

type TextZone struct {
	Paragraphs []Paragraph
}

type PageCoordinates struct {
	W float64
	H float64
	T float64
	L float64
}

type Paragraph struct {
	PageCoordinates
	Lines []Line
}

type Line struct {
	PageCoordinates
	Words []Word
}

type Word struct {
	PageCoordinates
	Fontsize   float64
	FontFamily string
	Bold       bool
	Text       string
}

type Region struct {
	PageCoordinates
	Type string
}

// ListParams
type ListParams struct {
	Limit  int
	Offset int
}

// SearchParams
type SearchParams struct {
	Query  string
	Type   string
	Limit  int
	Offset int
}

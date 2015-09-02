package giniapi

import (
	"fmt"
	"net/http"
)

var (
	ErrUploadFailed = "failed to upoad document"
	ErrDocumentGet  = "failed to GET document object"
	ErrPostFailed   = "failed to complete POST request"
)

// APIError provides additional error informations
type APIError struct {
	StatusCode int
	Message    string
	RequestID  string
	DocumentID string
}

// Error satisifes the Error interface
func (e *APIError) Error() string {
	return fmt.Sprintf("%s", e.Message)
}

// NewHttpError is a wrapper to simplify the error creation
func NewHttpError(message, docId string, err error, response *http.Response) *APIError {
	ae := APIError{
		Message:    message,
		DocumentID: docId,
	}

	// Sanity check for response pointer
	if response != nil {
		ae.StatusCode = response.StatusCode
		ae.RequestID = response.Header.Get("X-Request-Id")
	}

	return &ae
}

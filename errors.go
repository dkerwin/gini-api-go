package giniapi

import (
	"fmt"
	"net/http"
)

var (
	ErrUploadFailed   = "failed to upoad document"
	ErrDocumentGet    = "failed to GET document object"
	ErrDocumentParse  = "failed to parse document json"
	ErrDocumentRead   = "failed to read document body"
	ErrDocumentList   = "failed to get document list"
	ErrDocumentSearch = "failed to complete your search"

	ErrHTTPPostFailed   = "failed to complete POST request"
	ErrHTTPGetFailed    = "failed to complete GET request"
	ErrHTTPDeleteFailed = "failed to complete GET request"
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
	return fmt.Sprintf("%s (HTTP status: %d, RequestID: %s, DocumentID: %s)",
		e.Message, e.StatusCode, e.RequestID, e.DocumentID)
}

// NewHttpError is a wrapper to simplify the error creation
func newHTTPError(message, docID string, err error, response *http.Response) *APIError {
	ae := APIError{
		Message:    message,
		DocumentID: docID,
	}

	// Sanity check for response pointer
	if response != nil {
		ae.StatusCode = response.StatusCode
		ae.RequestID = response.Header.Get("X-Request-Id")
	}

	return &ae
}

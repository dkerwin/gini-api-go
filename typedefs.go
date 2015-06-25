package giniapi

// APIError struct
type APIError struct {
	// StatusCode
	StatusCode int

	// APIMessage
	Message string

	// APIRequestId
	RequestID string

	// DocumentId
	DocumentID string
}

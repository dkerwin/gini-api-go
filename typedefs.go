package giniapi

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

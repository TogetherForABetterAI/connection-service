package models

// APIError represents an error response in the API, according to RFC 7807.
type APIError struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail"`
	Instance string `json:"instance"`
}

// ServiceError represents an error from an external service call
// It preserves the HTTP status code and response body for proper propagation
type ServiceError struct {
	StatusCode   int
	ResponseBody string
	Message      string
}

func (e *ServiceError) Error() string {
	return e.Message
}

// NewServiceError creates a new ServiceError with the given parameters
func NewServiceError(statusCode int, responseBody, message string) *ServiceError {
	return &ServiceError{
		StatusCode:   statusCode,
		ResponseBody: responseBody,
		Message:      message,
	}
}

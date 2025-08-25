package models


// APIError represents an error response in the API, according to RFC 7807.
type APIError struct {
    Type     string `json:"type"`	
    Title    string `json:"title"`		
    Status   int    `json:"status"`
    Detail   string `json:"detail"`
    Instance string `json:"instance"`
}

// ConnectRequest represents the body of a request to create a snap.
type ConnectRequest struct {
    Token     string `json:"token"`
	InputsFmt string `json:"inputs_fmt"`
	OutputsFmt string `json:"outputs_fmt"`
}

type ConnectResponse struct {
    Status  string `json:"status"`
    Message string `json:"message"`
}
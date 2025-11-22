package models

// ConnectRequest represents the body of a request to create a connection.
type ConnectRequest struct {
	ClientId string `json:"client_id"`
	Token    string `json:"token"`
}

// ConnectResponse represents the response after a successful connection.
type ConnectResponse struct {
	Status      string               `json:"status"`
	Message     string               `json:"message"`
	Credentials *RabbitMQCredentials `json:"credentials,omitempty"`
}

// RabbitMQCredentials contains the RabbitMQ connection details for a client
type RabbitMQCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     int32  `json:"port"`
}

// ConnectNotification represents a notification sent when a client connects
type ConnectNotification struct {
	ClientId      string `json:"client_id"`
	SessionId     string `json:"session_id"`
	InputsFormat  string `json:"inputs_format"`
	OutputsFormat string `json:"outputs_format"`
	ModelType     string `json:"model_type"`
}

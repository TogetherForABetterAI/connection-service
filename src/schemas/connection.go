package schemas

// ConnectRequest represents the body of a request to create a connection.
type ConnectRequest struct {
	UserID string `json:"user_id"`
	Token  string `json:"token"`
}

// ConnectResponse represents the response after a successful connection.
type ConnectResponse struct {
	Status        string               `json:"status"`
	Message       string               `json:"message"`
	Credentials   *RabbitMQCredentials `json:"credentials,omitempty"`
	InputsFormat  string               `json:"inputs_format"`
	OutputsFormat string               `json:"outputs_format"`
	ModelType     string               `json:"model_type"`
}

// RabbitMQCredentials contains the RabbitMQ connection details for a client
type RabbitMQCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     int32  `json:"port"`
}

// NotifyNewConnection represents a notification sent when a client connects
type NotifyNewConnection struct {
	UserID        string `json:"user_id"`
	SessionId     string `json:"session_id"`
	Email         string `json:"email"`
	InputsFormat  string `json:"inputs_format"`
	OutputsFormat string `json:"outputs_format"`
	ModelType     string `json:"model_type"`
}

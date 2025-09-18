package service

import (
	"auth-gateway/src/rabbitmq"
	"encoding/json"
	"fmt"
)

type ConnectionService struct {
	Publisher rabbitmq.Publisher
}

func NewConnectionService(publisher rabbitmq.Publisher) *ConnectionService {
	return &ConnectionService{Publisher: publisher}
}

type ConnectNotification struct {
	ClientId      string `json:"client_id"`
	InputsFormat  string `json:"inputs_format"`
	OutputsFormat string `json:"outputs_format"`
}

func (s *ConnectionService) NotifyNewConnection(clientId, inputsFormat, outputsFormat string) error {
	notification := ConnectNotification{
		ClientId:      clientId,
		InputsFormat:  inputsFormat,
		OutputsFormat: outputsFormat,
	}
	body, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}
	return s.Publisher.Publish("new_connections", body)
}

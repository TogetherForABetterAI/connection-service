package service

import (
	"connection-service/src/config"
	"connection-service/src/middleware"
	"connection-service/src/models"
	"encoding/json"
	"fmt"
)

type ConnectionService struct {
	Publisher middleware.Publisher
}

func NewConnectionService(publisher middleware.Publisher) *ConnectionService {
	return &ConnectionService{Publisher: publisher}
}

func (s *ConnectionService) NotifyNewConnection(clientId, inputsFormat, outputsFormat, modelType string) error {

	exchangeName := config.CONNECTION_EXCHANGE

	notification := models.ConnectNotification{
		ClientId:      clientId,
		InputsFormat:  inputsFormat,
		OutputsFormat: outputsFormat,
		ModelType:     modelType,
	}
	body, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}
	return s.Publisher.Publish(exchangeName, body)
}

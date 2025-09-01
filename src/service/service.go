package service

import (
	"github.com/sirupsen/logrus"

)

type Service struct {
    Logger *logrus.Logger
}

// NewService creates a new instance of Service.
func NewService(logger *logrus.Logger) *Service {
    return &Service{
        Logger: logger,
    }
}


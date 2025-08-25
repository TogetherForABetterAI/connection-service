package logger

import (
    "github.com/sirupsen/logrus"
)

// Logger is a global variable that represents the logger instance.
var Logger *logrus.Logger

// Init initializes the logger by creating a new instance of logrus.Logger
func Init() {
    Logger = logrus.New()
    Logger.SetFormatter(&logrus.TextFormatter{
        FullTimestamp: true,
    })
}
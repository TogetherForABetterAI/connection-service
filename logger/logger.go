package logger

import (
	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

func Init() {
	Logger = logrus.New()
	Logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
}

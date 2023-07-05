package log

import (
	"go.uber.org/zap"
)

func InitLogger() *zap.Logger {
	logger, _ := zap.NewProduction()
	return logger
}

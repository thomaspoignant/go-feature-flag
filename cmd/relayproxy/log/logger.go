package log

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	ZapLogger *zap.Logger
	Atom      zap.AtomicLevel
}

func InitLogger() *Logger {
	logger := &Logger{Atom: zap.NewAtomicLevel()}

        prodConfig := zap.NewProductionConfig()
        prodConfig.Level = logger.Atom
	logger.ZapLogger = prodConfig.Build()

	return logger
}

package log

import (
	"go.uber.org/zap"
)

type Logger struct {
	ZapLogger *zap.Logger
	Atom      zap.AtomicLevel
}

func InitLogger() *Logger {
	logger := &Logger{Atom: zap.NewAtomicLevel()}

	prodConfig := zap.NewProductionConfig()
	prodConfig.Level = logger.Atom
	zapLogger, _ := prodConfig.Build()
	logger.ZapLogger = zapLogger

	return logger
}

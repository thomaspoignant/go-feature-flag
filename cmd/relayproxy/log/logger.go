package log

import (
	// register logfmt encoder
	_ "github.com/jsternberg/zap-logfmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	ZapLogger *zap.Logger
	config    zap.Config
}

func InitLogger() *Logger {
	zapConfig := zap.NewProductionConfig()
	zapLogger, _ := zapConfig.Build()

	logger := Logger{
		ZapLogger: zapLogger,
		config:    zapConfig,
	}

	return &logger
}

// Update the logger with a new format and level. Replaces the underlying
// zap.Logger with a new one. Not safe for concurrent use.
func (l *Logger) Update(format string, level zapcore.Level) {
	if format == "" {
		format = "json"
	}

	l.config.Level.SetLevel(level)
	l.config.Encoding = format

	zapLogger, _ := l.config.Build()

	l.ZapLogger = zapLogger
}

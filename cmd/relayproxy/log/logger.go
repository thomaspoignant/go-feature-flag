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

	encoderCfg := zap.NewProductionEncoderConfig()
	logger.ZapLogger = zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.Lock(os.Stdout),
		logger.Atom,
	))

	return logger
}

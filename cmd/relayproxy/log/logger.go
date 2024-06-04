package log

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitLogger(a *zap.AtomicLevel) *zap.Logger {
	var atom zap.AtomicLevel
	if a == nil {
		a := zap.NewAtomicLevel()
		a.SetLevel(zap.InfoLevel)
	}

	encoderCfg := zap.NewProductionEncoderConfig()

	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.Lock(os.Stdout),
		atom,
	))

	return logger
}

package fflog

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"strings"
	"time"
)

// LogDateFormat is the default log format
const LogDateFormat = time.RFC3339

// FFLogger is the internal logger struct for GO Feature Flag
type FFLogger struct {
	LeveledLogger *slog.Logger
	LegacyLogger  *log.Logger
}

// Error is the function to use to log error
func (f *FFLogger) Error(msg string, keysAndValues ...any) {
	if f != nil && f.LeveledLogger != nil {
		f.LeveledLogger.ErrorContext(context.Background(), msg, keysAndValues...)
		return
	}
	f.legacyLog("ERROR", msg, keysAndValues...)
}

// Info is the function to use to log info
func (f *FFLogger) Info(msg string, keysAndValues ...any) {
	if f != nil && f.LeveledLogger != nil {
		f.LeveledLogger.InfoContext(context.Background(), msg, keysAndValues...)
		return
	}
	f.legacyLog("INFO", msg, keysAndValues...)
}

// Debug is the function to use to log debug
func (f *FFLogger) Debug(msg string, keysAndValues ...any) {
	if f != nil && f.LeveledLogger != nil {
		f.LeveledLogger.DebugContext(context.Background(), msg, keysAndValues...)
		return
	}
	f.legacyLog("DEBUG", msg, keysAndValues...)
}

// Warn is the function to use to log warn
func (f *FFLogger) Warn(msg string, keysAndValues ...any) {
	if f != nil && f.LeveledLogger != nil {
		f.LeveledLogger.WarnContext(context.Background(), msg, keysAndValues...)
		return
	}
	f.legacyLog("WARN", msg, keysAndValues...)
}

func (f *FFLogger) legacyLog(level string, msg string, keysAndValues ...any) {
	if f != nil && f.LegacyLogger != nil {
		if len(keysAndValues) == 0 {
			f.LegacyLogger.Printf("%s %s %s", time.Now().Format("2006/01/02 15:04:05"), level, msg)
			return
		}

		attrs := make([]string, 0)
		for _, attr := range keysAndValues {
			attrs = append(attrs, fmt.Sprintf("%v", attr))
		}
		f.LegacyLogger.Printf(
			"%s %s %s %v",
			time.Now().Format("2006/01/02 15:04:05"),
			level,
			msg,
			strings.Join(attrs, " "),
		)
	}
}

// GetLogLogger is returning a classic logger from a slog logger
func (f *FFLogger) GetLogLogger(level slog.Level) *log.Logger {
	if f.LeveledLogger != nil {
		return slog.NewLogLogger(f.LeveledLogger.Handler(), level)
	}
	return f.LegacyLogger
}

// ConvertToFFLogger is converting a classic logger to our internal logger.
func ConvertToFFLogger(logger *log.Logger) *FFLogger {
	return &FFLogger{
		LegacyLogger: logger,
	}
}

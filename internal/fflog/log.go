package fflog

import (
	"log"
	"time"
)

const LogDateFormat = time.RFC3339

// Logger is an internal logger that use the logger we give in the configs
type Logger struct {
	Logger *log.Logger
}

// Printf is printing the log in the logger if present.
func (l *Logger) Printf(format string, v ...interface{}) {
	if l.Logger != nil {
		l.Logger.Printf(format, v...)
	}
}

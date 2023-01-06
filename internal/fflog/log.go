package fflog

import (
	"log"
	"time"
)

const LogDateFormat = time.RFC3339

func Printf(logger *log.Logger, format string, v ...interface{}) {
	if logger != nil {
		date := time.Now().Format(LogDateFormat)
		v = append([]interface{}{date}, v...)
		logger.Printf("[%v] "+format, v...)
	}
}

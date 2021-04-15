package fflog

import (
	"log"
	"time"
)

func Printf(logger *log.Logger, format string, v ...interface{}) {
	if logger != nil {
		date := time.Now().Format(time.RFC3339)
		v = append([]interface{}{date}, v...)
		logger.Printf("[%v] "+format, v...)
	}
}

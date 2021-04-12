package fflog

import "log"

func Printf(logger *log.Logger, format string, v ...interface{}) {
	if logger != nil {
		logger.Printf(format, v...)
	}
}

package api

import (
	"sync"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// bodyDumpHandler logs request bodies with truncation for large payloads.
// Bodies larger than maxBodyLogSize are truncated and suffixed with a marker.
func bodyDumpHandler(logger *zap.Logger) func(echo.Context, []byte, []byte) {
	maxBodyLogSize := 8192 // 8KiB
	truncatedBodySuffix := []byte("... truncated ...")
	truncatedSize := maxBodyLogSize + len(truncatedBodySuffix)

	// use a sync.Pool to reuse the same buffer for large requests - this
	// reduces allocations and can significantly improve performance.
	bufferPool := &sync.Pool{
		New: func() any {
			buf := make([]byte, truncatedSize)
			copy(buf[maxBodyLogSize:], truncatedBodySuffix)
			return &buf
		},
	}

	return func(_ echo.Context, reqBody []byte, _ []byte) {
		if len(reqBody) > maxBodyLogSize {
			bufPtr := bufferPool.Get().(*[]byte)
			truncated := *bufPtr
			copy(truncated[:maxBodyLogSize], reqBody[:maxBodyLogSize])
			logger.Debug("Request info", zap.ByteString("request_body", truncated))
			bufferPool.Put(bufPtr)
		} else {
			logger.Debug("Request info", zap.ByteString("request_body", reqBody))
		}
	}
}

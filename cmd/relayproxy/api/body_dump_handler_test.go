package api

import (
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func Test_bodyDumpHandler_smallBody(t *testing.T) {
	core, recorded := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	body := []byte("small body")
	handler := bodyDumpHandler(logger)
	handler(echo.Context(nil), body, nil)

	logs := recorded.All()
	assert.Len(t, logs, 1)
	assert.Equal(t, "Request info", logs[0].Message)

	loggedBody := logs[0].Context[0].Interface.([]byte)
	assert.Equal(t, body, loggedBody)
}

func Test_bodyDumpHandler_largeBody(t *testing.T) {
	core, recorded := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	body := []byte(strings.Repeat("x", 20*1024)) // 20KB
	handler := bodyDumpHandler(logger)
	handler(echo.Context(nil), body, nil)

	logs := recorded.All()
	assert.Len(t, logs, 1)

	loggedBody := logs[0].Context[0].Interface.([]byte)
	assert.True(t, len(loggedBody) < len(body), "logged body should be truncated")
	assert.True(t, strings.HasSuffix(string(loggedBody), "... truncated ..."))
}

func Benchmark_bodyDumpHandler(b *testing.B) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)
	handler := bodyDumpHandler(logger)

	b.Run("small body", func(b *testing.B) {
		body := []byte(strings.Repeat("x", 100))
		b.ReportAllocs()
		b.ResetTimer()

		for b.Loop() {
			handler(echo.Context(nil), body, nil)
		}
	})

	b.Run("large body", func(b *testing.B) {
		body := []byte(strings.Repeat("x", 1024*1024)) // 1MB
		b.ReportAllocs()
		b.ResetTimer()

		for b.Loop() {
			handler(echo.Context(nil), body, nil)
		}
	})
}

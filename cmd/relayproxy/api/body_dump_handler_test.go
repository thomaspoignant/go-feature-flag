package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func Test_bodyDumpSkipper(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		level zapcore.Level
		path  string
		want  bool
	}{
		{name: "debug on regular endpoint", level: zapcore.DebugLevel, path: "/v1/allflags", want: false},
		{name: "debug on swagger", level: zapcore.DebugLevel, path: "/swagger/index.html", want: true},
		{name: "debug on validate endpoint", level: zapcore.DebugLevel, path: "/admin/v1/flags/validate", want: true},
		{name: "debug on other admin endpoint", level: zapcore.DebugLevel, path: "/admin/v1/retriever/refresh", want: false},
		{name: "info level", level: zapcore.InfoLevel, path: "/v1/allflags", want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			core, _ := observer.New(tt.level)
			skipper := bodyDumpSkipper(zap.New(core))
			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, tt.path, nil)
			c := echo.New().NewContext(req, httptest.NewRecorder())
			assert.Equal(t, tt.want, skipper(c))
		})
	}
}

func Test_bodyDumpHandler_smallBody(t *testing.T) {
	core, recorded := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	body := []byte("small body")
	handler := bodyDumpHandler(logger)
	handler(nil, body, nil, nil)

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
	handler(nil, body, nil, nil)

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
			handler(nil, body, nil, nil)
		}
	})

	b.Run("large body", func(b *testing.B) {
		body := []byte(strings.Repeat("x", 1024*1024)) // 1MB
		b.ReportAllocs()
		b.ResetTimer()

		for b.Loop() {
			handler(nil, body, nil, nil)
		}
	})
}

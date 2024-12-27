package log_test

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/log"
	"go.uber.org/zap"
)

// temporarily replace stderr so we can capture it
// close the closer before reading from the reader
// not safe for concurrent use!
func replaceStderr(t *testing.T) (io.Reader, io.Closer) {
	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stderr = w

	t.Cleanup(func() { os.Stderr = oldStderr })

	return r, w
}

func TestInitLogger(t *testing.T) {
	r, w := replaceStderr(t)

	logger := log.InitLogger()
	logger.ZapLogger.Info("test message")

	w.Close()

	b, err := io.ReadAll(r)
	require.NoError(t, err)

	assert.Contains(t, string(b), `"level":"info"`)
	assert.Contains(t, string(b), `"msg":"test message"`)
}

func TestLoggerUpdate(t *testing.T) {
	t.Run("update to logfmt/warn", func(t *testing.T) {
		r, w := replaceStderr(t)

		logger := log.InitLogger()
		logger.Update("logfmt", zap.WarnLevel)
		logger.ZapLogger.Info("hidden")
		logger.ZapLogger.Warn("danger, Will Robinson!")

		w.Close()

		b, err := io.ReadAll(r)
		require.NoError(t, err)

		assert.NotContains(t, string(b), `{`)
		assert.NotContains(t, string(b), `info`)

		assert.Contains(t, string(b), `level=warn`)
		assert.Contains(t, string(b), `msg="danger, Will Robinson!"`)
	})

	t.Run("json by default", func(t *testing.T) {
		r, w := replaceStderr(t)

		logger := log.InitLogger()
		logger.Update("", zap.DebugLevel)
		logger.ZapLogger.Info("hello world")
		logger.ZapLogger.Debug("debugging")

		w.Close()

		b, err := io.ReadAll(r)
		require.NoError(t, err)

		assert.Contains(t, string(b), `"level":"debug"`)
		assert.Contains(t, string(b), `"msg":"debugging"`)
	})
}

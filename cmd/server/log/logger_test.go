package log_test

import (
	"github.com/thomaspoignant/go-feature-flag/cmd/server/log"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

// Test that InitLogger returns a valid *zap.Logger
func TestInitLogger(t *testing.T) {
	logger := log.InitLogger()
	assert.NotNil(t, logger, "Expected logger to not be nil")
}

// Test that the logger outputs the expected message
func TestLoggerOutput(t *testing.T) {
	// Create an observer to capture logs
	core, observedLogs := observer.New(zap.InfoLevel)

	// Create a new logger with the observer
	logger := zap.New(core)

	// Set the global logger to use the new logger
	zap.ReplaceGlobals(logger)

	// Initialize the logger from the package
	log.InitLogger()

	// Log a message
	zap.L().Info("test message")

	// Ensure that the message was logged as expected
	logs := observedLogs.All()
	assert.Equal(t, 1, len(logs), "Expected 1 log message, got %d", len(logs))
	assert.Equal(t, "test message", logs[0].Message, "Unexpected log message")
}

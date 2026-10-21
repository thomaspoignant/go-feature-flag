package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service/stream"
	"go.uber.org/zap"
)

type recordingAPIServer struct{ calls *[]string }

func (s recordingAPIServer) Stop(_ context.Context) { *s.calls = append(*s.calls, "api.Stop") }

type recordingFlagsetManager struct {
	service.FlagsetManager
	calls *[]string
}

func (m recordingFlagsetManager) Close() { *m.calls = append(*m.calls, "flagsets.Close") }

func TestShutdown(t *testing.T) {
	var calls []string
	shutdown(
		zap.NewNop(),
		recordingAPIServer{calls: &calls},
		stream.NewWebsocketService(),
		stream.NewSSEService(),
		recordingFlagsetManager{calls: &calls},
		nil,
	)
	// The flagsets must be closed (flushing the exporters) after the server stopped accepting requests.
	assert.Equal(t, []string{"api.Stop", "flagsets.Close"}, calls)
}

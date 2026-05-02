package service

import (
	"encoding/json"
	"net/http"

	"github.com/r3labs/sse/v2"
	"github.com/thomaspoignant/go-feature-flag/notifier"
)

// SSEService is the service interface that handles SSE connections.
// It dispatches flag change events scoped to specific flagset names so that
// only clients connected with an API key belonging to a given flagset receive
// the corresponding events.
type SSEService interface {
	// BroadcastFlagChanges sends the diff to all clients subscribed to the
	// given flagset stream.
	BroadcastFlagChanges(flagsetName string, diff notifier.DiffCache)
	// ServeHTTP handles incoming SSE client connections. The request must
	// carry a "stream" query parameter set to the target flagset name.
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	// Close shuts down the SSE server and disconnects all clients.
	Close()
}

// NewSSEService creates a new SSEService backed by r3labs/sse.
func NewSSEService() SSEService {
	server := sse.New()
	server.AutoReplay = false
	server.AutoStream = true
	return &sseServiceImpl{server: server}
}

type sseServiceImpl struct {
	server *sse.Server
}

func (s *sseServiceImpl) BroadcastFlagChanges(flagsetName string, diff notifier.DiffCache) {
	data, err := json.Marshal(diff)
	if err != nil {
		return
	}
	s.server.Publish(flagsetName, &sse.Event{
		Data: data,
	})
}

func (s *sseServiceImpl) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.server.ServeHTTP(w, r)
}

func (s *sseServiceImpl) Close() {
	s.server.Close()
}

package stream

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/r3labs/sse/v2"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/model"
	"github.com/thomaspoignant/go-feature-flag/notifier"
)

// SSEService is the service interface that handles SSE connections.
// It dispatches OFREP flag change events (OpenFeature ADR-0008) scoped to specific
// flagset names so that only clients connected with an API key belonging to a given
// flagset receive the corresponding events.
type SSEService interface {
	// BroadcastFlagChanges sends an OFREP refetchEvaluation event to all clients
	// subscribed to the given flagset stream.
	BroadcastFlagChanges(flagsetName string, diff notifier.DiffCache) error
	// ServeHTTP handles incoming SSE client connections. The request must
	// carry a "stream" query parameter set to the target flagset name.
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	// SetOnSubscribe registers a callback invoked when a client subscribes to
	// any stream. The streamID matches the flagset name used in BroadcastFlagChanges.
	SetOnSubscribe(fn func(streamID string))
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

// BroadcastFlagChanges only notifies that the flags changed, the diff is not sent:
// providers re-fetch their evaluations when they receive the event.
func (s *sseServiceImpl) BroadcastFlagChanges(flagsetName string, _ notifier.DiffCache) error {
	now := time.Now()
	data, err := json.Marshal(model.OFREPSSEEvent{
		Type:         model.OFREPSSEEventTypeRefetchEvaluation,
		LastModified: now.Unix(),
	})
	if err != nil {
		return fmt.Errorf("sse: failed to marshal OFREP event for stream %q: %w", flagsetName, err)
	}
	s.server.Publish(flagsetName, &sse.Event{
		ID:    []byte(strconv.FormatInt(now.UnixNano(), 10)),
		Event: []byte("message"),
		Data:  data,
	})
	return nil
}

func (s *sseServiceImpl) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.server.ServeHTTP(w, r)
}

func (s *sseServiceImpl) SetOnSubscribe(fn func(streamID string)) {
	s.server.OnSubscribe = func(streamID string, _ *sse.Subscriber) {
		fn(streamID)
	}
}

func (s *sseServiceImpl) Close() {
	s.server.Close()
}

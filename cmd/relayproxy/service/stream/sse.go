package stream

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
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
func NewSSEService(opts ...Option) SSEService {
	server := sse.New()
	server.AutoReplay = false
	server.AutoStream = true
	s := &sseServiceImpl{
		server:        server,
		activeStreams: map[string]int{},
		done:          make(chan struct{}),
	}
	if interval := newOptions(opts...).sseHeartbeatInterval; interval > 0 {
		go s.sendHeartbeats(interval)
	}
	return s
}

type sseServiceImpl struct {
	server *sse.Server
	// activeStreams counts the connected clients per stream, used to send the heartbeats.
	activeStreams   map[string]int
	muActiveStreams sync.Mutex
	done            chan struct{}
	closeOnce       sync.Once
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
	streamID := r.URL.Query().Get("stream")
	s.trackStream(streamID, 1)
	defer s.trackStream(streamID, -1)

	if flusher, ok := w.(http.Flusher); ok {
		w = &connectedCommentWriter{ResponseWriter: w, flusher: flusher}
	}
	s.server.ServeHTTP(w, r)
}

func (s *sseServiceImpl) SetOnSubscribe(fn func(streamID string)) {
	s.server.OnSubscribe = func(streamID string, _ *sse.Subscriber) {
		fn(streamID)
	}
}

func (s *sseServiceImpl) Close() {
	s.closeOnce.Do(func() { close(s.done) })
	s.server.Close()
}

func (s *sseServiceImpl) trackStream(streamID string, delta int) {
	s.muActiveStreams.Lock()
	defer s.muActiveStreams.Unlock()
	s.activeStreams[streamID] += delta
	if s.activeStreams[streamID] <= 0 {
		delete(s.activeStreams, streamID)
	}
}

// sendHeartbeats periodically sends an SSE comment on every stream with connected clients,
// so that proxies and load balancers do not close the connections while no flag changes.
func (s *sseServiceImpl) sendHeartbeats(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-s.done:
			return
		case <-ticker.C:
			s.muActiveStreams.Lock()
			streamIDs := make([]string, 0, len(s.activeStreams))
			for streamID := range s.activeStreams {
				streamIDs = append(streamIDs, streamID)
			}
			s.muActiveStreams.Unlock()
			for _, streamID := range streamIDs {
				s.server.TryPublish(streamID, &sse.Event{Comment: []byte("heartbeat")})
			}
		}
	}
}

// connectedCommentWriter sends an SSE comment right after the response headers.
// Some browsers (e.g. Firefox) fire the EventSource "open" event only once the first
// bytes of the body are received: without it, clients stay "connecting" until the first event.
type connectedCommentWriter struct {
	http.ResponseWriter
	flusher http.Flusher
	status  int
	sent    bool
}

func (w *connectedCommentWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *connectedCommentWriter) Flush() {
	if !w.sent && w.status == http.StatusOK {
		w.sent = true
		_, _ = io.WriteString(w.ResponseWriter, ": connected\n\n")
	}
	w.flusher.Flush()
}

func (w *connectedCommentWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

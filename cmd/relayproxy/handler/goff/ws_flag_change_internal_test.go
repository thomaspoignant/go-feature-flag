package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service/stream"
	"go.uber.org/zap"
)

// newWSConnPair stands up a throwaway websocket server and returns the
// server-side and client-side *websocket.Conn so the unexported write paths
// (threadSafeConn / pingPongLoop) can be tested directly.
func newWSConnPair(t *testing.T) (server, client *websocket.Conn) {
	t.Helper()
	upgrader := websocket.Upgrader{}
	serverConnCh := make(chan *websocket.Conn, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		serverConnCh <- c
	}))
	t.Cleanup(srv.Close)

	url := "ws" + strings.TrimPrefix(srv.URL, "http")
	client, _, err := websocket.DefaultDialer.Dial(url, nil)
	require.NoError(t, err)
	server = <-serverConnCh
	t.Cleanup(func() {
		_ = client.Close()
		_ = server.Close()
	})
	return server, client
}

func Test_threadSafeConn_ping(t *testing.T) {
	server, client := newWSConnPair(t)

	pinged := make(chan struct{}, 1)
	client.SetPingHandler(func(string) error {
		select {
		case pinged <- struct{}{}:
		default:
		}
		return nil
	})
	// A reader is required for the client to process incoming control frames.
	go func() {
		for {
			if _, _, err := client.ReadMessage(); err != nil {
				return
			}
		}
	}()

	tc := &threadSafeConn{conn: server}

	// Happy path: the ping reaches the client.
	require.NoError(t, tc.ping())
	select {
	case <-pinged:
	case <-time.After(2 * time.Second):
		t.Fatal("client never received the ping")
	}

	// Error path: writing to a closed connection returns an error.
	_ = server.Close()
	assert.Error(t, tc.ping())
}

func Test_pingPongLoop_pings_then_stops_on_context(t *testing.T) {
	server, client := newWSConnPair(t)

	pinged := make(chan struct{}, 8)
	client.SetPingHandler(func(string) error {
		select {
		case pinged <- struct{}{}:
		default:
		}
		return nil
	})
	go func() {
		for {
			if _, _, err := client.ReadMessage(); err != nil {
				return
			}
		}
	}()

	f := NewWsFlagChange(stream.NewWebsocketService(), zap.NewNop())
	f.pingInterval = 10 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		f.pingPongLoop(ctx, &threadSafeConn{conn: server})
		close(done)
	}()

	select {
	case <-pinged:
	case <-time.After(2 * time.Second):
		t.Fatal("ping loop never sent a ping")
	}

	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("ping loop did not stop when the context was cancelled")
	}
}

func Test_pingPongLoop_stops_on_write_error(t *testing.T) {
	server, client := newWSConnPair(t)
	// Close both ends so the next write fails immediately.
	_ = client.Close()
	_ = server.Close()

	f := NewWsFlagChange(stream.NewWebsocketService(), zap.NewNop())
	f.pingInterval = 5 * time.Millisecond

	done := make(chan struct{})
	go func() {
		f.pingPongLoop(context.Background(), &threadSafeConn{conn: server})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("ping loop did not stop after a write error")
	}
}

// Test_pingPongLoop_zero_interval_uses_default ensures a zero-valued interval
// falls back to the default rather than panicking time.NewTicker.
func Test_pingPongLoop_zero_interval_uses_default(t *testing.T) {
	server, _ := newWSConnPair(t)

	f := NewWsFlagChange(stream.NewWebsocketService(), zap.NewNop())
	f.pingInterval = 0 // force the fallback branch

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		f.pingPongLoop(ctx, &threadSafeConn{conn: server})
		close(done)
	}()

	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("ping loop did not start/stop with a zero interval")
	}
}

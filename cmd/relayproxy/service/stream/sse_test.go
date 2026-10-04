package stream_test

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/model"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service/stream"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/modules/core/testutils/testconvert"
	"github.com/thomaspoignant/go-feature-flag/notifier"
)

func TestSSEService_BroadcastAndReceive(t *testing.T) {
	tests := []struct {
		name             string
		subscribeFlagset string
		broadcastFlagset string
		diff             notifier.DiffCache
		expectReceive    bool
	}{
		{
			name:             "client receives event from its own stream",
			subscribeFlagset: "flagsetA",
			broadcastFlagset: "flagsetA",
			diff: notifier.DiffCache{
				Added: map[string]flag.Flag{
					"flag-1": &flag.InternalFlag{
						Variations: &map[string]*any{
							"A": testconvert.Interface(true),
						},
						DefaultRule: &flag.Rule{VariationResult: testconvert.String("A")},
					},
				},
			},
			expectReceive: true,
		},
		{
			name:             "client does not receive event from a different stream",
			subscribeFlagset: "flagsetA",
			broadcastFlagset: "flagsetB",
			diff: notifier.DiffCache{
				Deleted: map[string]flag.Flag{
					"flag-2": &flag.InternalFlag{},
				},
			},
			expectReceive: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			sseService := stream.NewSSEService()
			defer sseService.Close()

			srv := httptest.NewServer(http.HandlerFunc(sseService.ServeHTTP))
			defer srv.Close()

			req, err := http.NewRequestWithContext(ctx, http.MethodGet,
				srv.URL+"?stream="+tt.subscribeFlagset, nil)
			require.NoError(t, err)

			subscribed := make(chan struct{}, 1)
			sseService.SetOnSubscribe(func(_ string) {
				select {
				case subscribed <- struct{}{}:
				default:
				}
			})

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, http.StatusOK, resp.StatusCode)

			select {
			case <-subscribed:
			case <-ctx.Done():
				t.Fatal("timed out waiting for SSE client to subscribe")
			}
			before := time.Now()
			require.NoError(t, sseService.BroadcastFlagChanges(tt.broadcastFlagset, tt.diff))
			if !tt.expectReceive {
				// Broadcast the correct stream so the client unblocks after.
				before = time.Now()
				require.NoError(t, sseService.BroadcastFlagChanges(tt.subscribeFlagset, notifier.DiffCache{}))
			}

			fields := map[string]string{}
			scanner := bufio.NewScanner(resp.Body)
			for scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(line, ":") {
					continue // SSE comment (connection comment or heartbeat), ignored by clients
				}
				if line == "" && len(fields) > 0 {
					break
				}
				if key, value, ok := strings.Cut(line, ": "); ok {
					fields[key] = value
				}
			}

			// The event id is a timestamp, it tells us which broadcast was received.
			id, err := strconv.ParseInt(fields["id"], 10, 64)
			require.NoError(t, err)
			assert.GreaterOrEqual(t, id, before.UnixNano(),
				"client should only receive the events of its own stream")
			assert.Equal(t, "message", fields["event"])
			var got model.OFREPSSEEvent
			require.NoError(t, json.Unmarshal([]byte(fields["data"]), &got))
			assert.Equal(t, model.OFREPSSEEvent{
				Type:         model.OFREPSSEEventTypeRefetchEvaluation,
				LastModified: got.LastModified,
			}, got)
			assert.GreaterOrEqual(t, got.LastModified, before.Unix())
			assert.NotContains(t, fields["data"], "flag-", "the flag diff must not be sent")
		})
	}
}

func TestSSEService_Close(t *testing.T) {
	sseService := stream.NewSSEService()
	sseService.Close()
}

func TestSSEService_Comments(t *testing.T) {
	tests := []struct {
		name              string
		query             string
		heartbeatInterval time.Duration
		wantStatus        int
		wantConnected     bool
		wantHeartbeat     bool
	}{
		{
			name:              "connected comment is the first bytes of the body",
			query:             "?stream=flagsetA",
			heartbeatInterval: 0,
			wantStatus:        http.StatusOK,
			wantConnected:     true,
			wantHeartbeat:     false,
		},
		{
			name:              "heartbeat comments are sent periodically",
			query:             "?stream=flagsetA",
			heartbeatInterval: 50 * time.Millisecond,
			wantStatus:        http.StatusOK,
			wantConnected:     true,
			wantHeartbeat:     true,
		},
		{
			name:              "negative interval disables the heartbeat",
			query:             "?stream=flagsetA",
			heartbeatInterval: -1,
			wantStatus:        http.StatusOK,
			wantConnected:     true,
			wantHeartbeat:     false,
		},
		{
			name:              "no connected comment on an error response",
			query:             "",
			heartbeatInterval: 50 * time.Millisecond,
			wantStatus:        http.StatusInternalServerError,
			wantConnected:     false,
			wantHeartbeat:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The body is read until this timeout, it leaves time for several heartbeats.
			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()
			sseService := stream.NewSSEService(stream.WithSSEHeartbeatInterval(tt.heartbeatInterval))
			defer sseService.Close()

			srv := httptest.NewServer(http.HandlerFunc(sseService.ServeHTTP))
			defer srv.Close()
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+tt.query, nil)
			require.NoError(t, err)
			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			require.Equal(t, tt.wantStatus, resp.StatusCode)

			lines := readLines(resp.Body)
			if tt.wantConnected {
				// No event is broadcast: the first bytes of the body must be the connection comment,
				// otherwise some browsers (e.g. Firefox) never fire the EventSource "open" event.
				require.GreaterOrEqual(t, len(lines), 2)
				assert.Equal(t, ": connected\n", lines[0])
				assert.Equal(t, "\n", lines[1], "the comment must be terminated by a blank line")
			} else {
				assert.NotContains(t, lines, ": connected\n")
			}
			heartbeats := countLine(lines, ": heartbeat\n")
			if tt.wantHeartbeat {
				assert.GreaterOrEqual(t, heartbeats, 2, "should receive heartbeats before the timeout")
			} else {
				assert.Zero(t, heartbeats)
			}
		})
	}
}

// readLines reads the body line by line until it ends or the request times out.
func readLines(body io.Reader) []string {
	var lines []string
	reader := bufio.NewReader(body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return lines
		}
		lines = append(lines, line)
	}
}

func countLine(lines []string, want string) int {
	count := 0
	for _, line := range lines {
		if line == want {
			count++
		}
	}
	return count
}

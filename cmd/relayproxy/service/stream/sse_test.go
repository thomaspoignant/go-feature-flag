package stream_test

import (
	"bufio"
	"context"
	"encoding/json"
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

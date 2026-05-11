package stream_test

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			require.NoError(t, sseService.BroadcastFlagChanges(tt.broadcastFlagset, tt.diff))

			if tt.expectReceive {
				scanner := bufio.NewScanner(resp.Body)
				var received string
				for scanner.Scan() {
					if data, ok := strings.CutPrefix(scanner.Text(), "data: "); ok {
						received = data
						break
					}
				}
				require.NotEmpty(t, received)
				expected, err := json.Marshal(tt.diff)
				require.NoError(t, err)
				assert.JSONEq(t, string(expected), received)
			} else {
				// Broadcast the correct stream so the client unblocks after.
				afterDiff := notifier.DiffCache{
					Added: map[string]flag.Flag{"marker": &flag.InternalFlag{}},
				}
				require.NoError(t, sseService.BroadcastFlagChanges(tt.subscribeFlagset, afterDiff))

				scanner := bufio.NewScanner(resp.Body)
				var received string
				for scanner.Scan() {
					if data, ok := strings.CutPrefix(scanner.Text(), "data: "); ok {
						received = data
						break
					}
				}
				require.NotEmpty(t, received)
				assert.Contains(t, received, "marker",
					"client should only receive the marker event, not the broadcast to a different flagset")
			}
		})
	}
}

func TestSSEService_Close(t *testing.T) {
	sseService := stream.NewSSEService()
	sseService.Close()
}

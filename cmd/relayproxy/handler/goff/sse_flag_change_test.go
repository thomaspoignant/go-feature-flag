package controller_test

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	controller "github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/handler/goff"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/model"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service/stream"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/modules/core/testutils/testconvert"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"go.uber.org/zap"
)

type mockFlagsetManagerSSE struct {
	flagsetName string
	isDefault   bool
	err         error
}

func (m *mockFlagsetManagerSSE) FlagSetName(_ string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.flagsetName, nil
}

func (m *mockFlagsetManagerSSE) IsDefaultFlagSet() bool { return m.isDefault }

func (m *mockFlagsetManagerSSE) FlagSet(_ string) (*ffclient.GoFeatureFlag, error) { return nil, nil }
func (m *mockFlagsetManagerSSE) AllFlagSets() (map[string]*ffclient.GoFeatureFlag, error) {
	return nil, nil
}
func (m *mockFlagsetManagerSSE) Default() *ffclient.GoFeatureFlag { return nil }
func (m *mockFlagsetManagerSSE) Close()                           {}
func (m *mockFlagsetManagerSSE) OnConfigChange(_ *config.Config)  {}

func Test_SSE_FlagChange(t *testing.T) {
	tests := []struct {
		name       string
		flagChange notifier.DiffCache
	}{
		{
			name: "update single flag",
			flagChange: notifier.DiffCache{
				Updated: map[string]notifier.DiffUpdated{
					"my-flag": {
						Before: &flag.InternalFlag{
							Variations: &map[string]*any{
								"A": testconvert.Interface(true),
								"B": testconvert.Interface(false),
							},
							DefaultRule: &flag.Rule{VariationResult: testconvert.String("A")},
						},
						After: &flag.InternalFlag{
							Variations: &map[string]*any{
								"A": testconvert.Interface(true),
								"B": testconvert.Interface(false),
							},
							DefaultRule: &flag.Rule{VariationResult: testconvert.String("B")},
						},
					},
				},
			},
		},
		{
			name: "add and delete flags at the same time",
			flagChange: notifier.DiffCache{
				Deleted: map[string]flag.Flag{
					"flag-1": &flag.InternalFlag{
						Variations: &map[string]*any{
							"A": testconvert.Interface(true),
						},
						DefaultRule: &flag.Rule{VariationResult: testconvert.String("A")},
					},
				},
				Added: map[string]flag.Flag{
					"flag-2": &flag.InternalFlag{
						Variations: &map[string]*any{
							"B": testconvert.Interface(false),
						},
						DefaultRule: &flag.Rule{VariationResult: testconvert.String("B")},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			sseService := stream.NewSSEService()
			defer sseService.Close()

			subscribed := make(chan struct{}, 1)
			sseService.SetOnSubscribe(func(_ string) {
				select {
				case subscribed <- struct{}{}:
				default:
				}
			})

			flagsetMgr := &mockFlagsetManagerSSE{
				flagsetName: "default",
				isDefault:   true,
			}

			ctrl := controller.NewSSEFlagChange(sseService, flagsetMgr, zap.NewNop())

			e := echo.New()
			e.GET("/stream/v1/sse/flag/change", ctrl.Handler)
			srv := httptest.NewServer(e)
			defer srv.Close()

			req, err := http.NewRequestWithContext(ctx, http.MethodGet,
				srv.URL+"/stream/v1/sse/flag/change", nil)
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Contains(t, resp.Header.Get("Content-Type"), "text/event-stream")

			select {
			case <-subscribed:
			case <-ctx.Done():
				t.Fatal("timed out waiting for SSE client to subscribe")
			}
			before := time.Now().Unix()
			require.NoError(t, sseService.BroadcastFlagChanges("default", tt.flagChange))

			event := readSSEEvent(t, resp)
			assert.NotEmpty(t, event["id"], "events should have an id")
			assert.Equal(t, "message", event["event"])
			var got model.OFREPSSEEvent
			require.NoError(t, json.Unmarshal([]byte(event["data"]), &got))
			assert.Equal(t, model.OFREPSSEEventTypeRefetchEvaluation, got.Type)
			assert.GreaterOrEqual(t, got.LastModified, before)
		})
	}
}

func Test_SSE_FlagChange_FlagsetScoping(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sseService := stream.NewSSEService()
	defer sseService.Close()

	subscribed := make(chan struct{}, 1)
	sseService.SetOnSubscribe(func(_ string) {
		select {
		case subscribed <- struct{}{}:
		default:
		}
	})

	flagsetMgr := &mockFlagsetManagerSSE{
		flagsetName: "flagsetA",
		isDefault:   false,
	}

	ctrl := controller.NewSSEFlagChange(sseService, flagsetMgr, zap.NewNop())

	e := echo.New()
	e.GET("/stream/v1/sse/flag/change", ctrl.Handler)
	srv := httptest.NewServer(e)
	defer srv.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		srv.URL+"/stream/v1/sse/flag/change?apiKey=key-a", nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	select {
	case <-subscribed:
	case <-ctx.Done():
		t.Fatal("timed out waiting for SSE client to subscribe")
	}

	// Broadcast to flagsetB -- the client on flagsetA should NOT receive it.
	require.NoError(t, sseService.BroadcastFlagChanges("flagsetB", notifier.DiffCache{
		Added: map[string]flag.Flag{"wrong-flag": &flag.InternalFlag{}},
	}))

	// Broadcast to flagsetA -- the client should receive this one.
	// The event id is a timestamp, so it tells us which broadcast was received.
	afterWrongBroadcast := time.Now().UnixNano()
	require.NoError(t, sseService.BroadcastFlagChanges("flagsetA", notifier.DiffCache{
		Added: map[string]flag.Flag{"right-flag": &flag.InternalFlag{}},
	}))

	event := readSSEEvent(t, resp)
	id, err := strconv.ParseInt(event["id"], 10, 64)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, id, afterWrongBroadcast,
		"client should only receive the event of its own flagset")
	assert.Contains(t, event["data"], model.OFREPSSEEventTypeRefetchEvaluation)
}

// readSSEEvent reads the next SSE event of the response and returns its fields.
func readSSEEvent(t *testing.T, resp *http.Response) map[string]string {
	t.Helper()
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
	require.NotEmpty(t, fields["data"], "should have received an SSE data line")
	return fields
}

func Test_SSE_FlagChange_Errors(t *testing.T) {
	tests := []struct {
		name           string
		apiKey         string
		flagsetMgr     *mockFlagsetManagerSSE
		expectedStatus int
	}{
		{
			name:   "empty apiKey in flagset mode returns 400",
			apiKey: "",
			flagsetMgr: &mockFlagsetManagerSSE{
				isDefault: false,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "FlagSetName error returns 400",
			apiKey: "unknown-key",
			flagsetMgr: &mockFlagsetManagerSSE{
				isDefault: false,
				err:       fmt.Errorf("unknown api key"),
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sseService := stream.NewSSEService()
			defer sseService.Close()

			ctrl := controller.NewSSEFlagChange(sseService, tt.flagsetMgr, zap.NewNop())

			e := echo.New()
			e.GET("/stream/v1/sse/flag/change", ctrl.Handler)
			srv := httptest.NewServer(e)
			defer srv.Close()

			url := srv.URL + "/stream/v1/sse/flag/change"
			if tt.apiKey != "" {
				url += "?apiKey=" + tt.apiKey
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

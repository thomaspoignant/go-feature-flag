package controller_test

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	controller "github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/handler/goff"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
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

			sseService := service.NewSSEService()
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
			require.NoError(t, sseService.BroadcastFlagChanges("default", tt.flagChange))

			scanner := bufio.NewScanner(resp.Body)
			var received string
			for scanner.Scan() {
				if data, ok := strings.CutPrefix(scanner.Text(), "data: "); ok {
					received = data
					break
				}
			}
			require.NotEmpty(t, received, "should have received an SSE data line")

			expected, err := json.Marshal(tt.flagChange)
			require.NoError(t, err)
			assert.JSONEq(t, string(expected), received)
		})
	}
}

func Test_SSE_FlagChange_FlagsetScoping(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sseService := service.NewSSEService()
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
	diff := notifier.DiffCache{
		Added: map[string]flag.Flag{
			"right-flag": &flag.InternalFlag{
				Variations: &map[string]*any{
					"A": testconvert.Interface(true),
				},
				DefaultRule: &flag.Rule{VariationResult: testconvert.String("A")},
			},
		},
	}
	require.NoError(t, sseService.BroadcastFlagChanges("flagsetA", diff))

	scanner := bufio.NewScanner(resp.Body)
	var received string
	for scanner.Scan() {
		if data, ok := strings.CutPrefix(scanner.Text(), "data: "); ok {
			received = data
			break
		}
	}
	require.NotEmpty(t, received, "should have received an SSE data line")
	assert.Contains(t, received, "right-flag")
	assert.NotContains(t, received, "wrong-flag")
}

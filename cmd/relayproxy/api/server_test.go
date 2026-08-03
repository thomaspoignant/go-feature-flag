package api_test

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	promdto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/api"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/proxynotifier"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service/stream"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/log"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/retrieverconf"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"github.com/thomaspoignant/go-feature-flag/testutils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/protobuf/encoding/protodelim"
)

// newTestLogger returns a logger whose Fatal() terminates the calling goroutine instead of the
// whole process. server.go reports a failed bind with zapLog.Fatal, and the default zap
// behaviour is os.Exit(1): fired from the background goroutine that runs StartWithContext, it
// takes the test binary down and every remaining test in the package is silently skipped.
// With this hook the server goroutine dies, the Fatal is still logged, and the test fails on
// its own readiness assertion with a message that points at the right place.
func newTestLogger(t *testing.T) *log.Logger {
	t.Helper()
	l := log.InitLogger()
	l.ZapLogger = l.ZapLogger.WithOptions(zap.WithFatalHook(zapcore.WriteThenGoexit))
	t.Cleanup(func() { _ = l.ZapLogger.Sync() })
	return l
}

// waitForServer blocks until the server accepts connections on baseURL. Binding a listener is
// asynchronous, so a fixed sleep is a coin flip: 10ms is below the Windows timer granularity and
// a loaded runner needs far more than that.
func waitForServer(t *testing.T, baseURL string) {
	t.Helper()
	require.Eventually(t, func() bool {
		resp, err := http.Get(baseURL + "/health")
		if err != nil {
			return false
		}
		_ = resp.Body.Close()
		return true
	}, 10*time.Second, 20*time.Millisecond, "server never became reachable on %s", baseURL)
}

func Test_Starting_RelayProxy_with_monitoring_on_same_port(t *testing.T) {
	port := testutils.GetFreePort(t)
	proxyConf := &config.Config{
		CommonFlagSet: config.CommonFlagSet{
			Retrievers: &[]retrieverconf.RetrieverConf{
				{
					Kind: "file",
					Path: "../../../testdata/flag-config.yaml",
				},
			},
		},
		Server: config.Server{
			Mode: config.ServerModeHTTP,
			Port: port,
		},
	}
	log := newTestLogger(t)

	metricsV2, err := metric.NewMetrics()
	if err != nil {
		log.ZapLogger.Error("impossible to initialize prometheus metrics", zap.Error(err))
	}
	wsService := stream.NewWebsocketService()
	defer wsService.Close() // close all the open connections
	prometheusNotifier := metric.NewPrometheusNotifier(metricsV2)
	proxyNotifier := proxynotifier.NewNotifierWebsocket(wsService)
	flagsetManager, err := service.NewFlagsetManager(proxyConf, log.ZapLogger, []notifier.Notifier{
		prometheusNotifier,
		proxyNotifier,
	}, nil)
	require.NoError(t, err)

	services := service.Services{
		MonitoringService: service.NewMonitoring(flagsetManager),
		WebsocketService:  wsService,
		FlagsetManager:    flagsetManager,
		Metrics:           metricsV2,
	}

	s := api.New(proxyConf, services, log.ZapLogger)
	go func() { s.StartWithContext(context.Background()) }()
	defer s.Stop(context.Background())

	baseURL := fmt.Sprintf("http://localhost:%d", port)
	waitForServer(t, baseURL)

	response, err := http.Get(baseURL + "/health")
	require.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	assert.Equal(t, http.StatusOK, response.StatusCode)

	responseM, err := http.Get(baseURL + "/metrics")
	require.NoError(t, err)
	defer func() { _ = responseM.Body.Close() }()
	assert.Equal(t, http.StatusOK, responseM.StatusCode)

	responseI, err := http.Get(baseURL + "/info")
	require.NoError(t, err)
	defer func() { _ = responseI.Body.Close() }()
	assert.Equal(t, http.StatusOK, responseI.StatusCode)
}

func Test_Starting_RelayProxy_with_monitoring_on_different_port(t *testing.T) {
	port := testutils.GetFreePort(t)
	monitoringPort := testutils.GetFreePort(t)
	proxyConf := &config.Config{
		CommonFlagSet: config.CommonFlagSet{
			Retrievers: &[]retrieverconf.RetrieverConf{
				{
					Kind: "file",
					Path: "../../../testdata/flag-config.yaml",
				},
			},
		},
		Server: config.Server{
			Mode:           config.ServerModeHTTP,
			Port:           port,
			MonitoringPort: monitoringPort,
		},
	}
	log := newTestLogger(t)

	metricsV2, err := metric.NewMetrics()
	if err != nil {
		log.ZapLogger.Error("impossible to initialize prometheus metrics", zap.Error(err))
	}
	wsService := stream.NewWebsocketService()
	defer wsService.Close() // close all the open connections
	prometheusNotifier := metric.NewPrometheusNotifier(metricsV2)
	proxyNotifier := proxynotifier.NewNotifierWebsocket(wsService)
	flagsetManager, err := service.NewFlagsetManager(proxyConf, log.ZapLogger, []notifier.Notifier{
		prometheusNotifier,
		proxyNotifier,
	}, nil)
	require.NoError(t, err)

	services := service.Services{
		MonitoringService: service.NewMonitoring(flagsetManager),
		WebsocketService:  wsService,
		FlagsetManager:    flagsetManager,
		Metrics:           metricsV2,
	}

	s := api.New(proxyConf, services, log.ZapLogger)
	go func() { s.StartWithContext(context.Background()) }()
	defer s.Stop(context.Background())

	baseURL := fmt.Sprintf("http://localhost:%d", port)
	monitoringURL := fmt.Sprintf("http://localhost:%d", monitoringPort)
	waitForServer(t, baseURL)
	waitForServer(t, monitoringURL)

	response, err := http.Get(baseURL + "/health")
	require.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	assert.Equal(t, http.StatusNotFound, response.StatusCode)

	responseM, err := http.Get(baseURL + "/metrics")
	require.NoError(t, err)
	defer func() { _ = responseM.Body.Close() }()
	assert.Equal(t, http.StatusNotFound, responseM.StatusCode)

	responseI, err := http.Get(baseURL + "/info")
	require.NoError(t, err)
	defer func() { _ = responseI.Body.Close() }()
	assert.Equal(t, http.StatusNotFound, responseI.StatusCode)

	responseH1, err := http.Get(monitoringURL + "/health")
	require.NoError(t, err)
	defer func() { _ = responseH1.Body.Close() }()
	assert.Equal(t, http.StatusOK, responseH1.StatusCode)

	responseM1, err := http.Get(monitoringURL + "/metrics")
	require.NoError(t, err)
	defer func() { _ = responseM1.Body.Close() }()
	assert.Equal(t, http.StatusOK, responseM1.StatusCode)

	responseI1, err := http.Get(monitoringURL + "/info")
	require.NoError(t, err)
	defer func() { _ = responseI1.Body.Close() }()
	assert.Equal(t, http.StatusOK, responseI1.StatusCode)
}

func Test_CheckOFREPAPIExists(t *testing.T) {
	port := testutils.GetFreePort(t)
	proxyConf := &config.Config{
		CommonFlagSet: config.CommonFlagSet{
			Retrievers: &[]retrieverconf.RetrieverConf{
				{
					Kind: "file",
					Path: "../../../testdata/flag-config.yaml",
				},
			},
		},
		Server: config.Server{
			Mode: config.ServerModeHTTP,
			Port: port,
		},
		AuthorizedKeys: config.APIKeys{
			Admin:      nil,
			Evaluation: []string{"test"},
		},
	}
	log := newTestLogger(t)

	metricsV2, err := metric.NewMetrics()
	if err != nil {
		log.ZapLogger.Error("impossible to initialize prometheus metrics", zap.Error(err))
	}
	wsService := stream.NewWebsocketService()
	defer wsService.Close() // close all the open connections
	prometheusNotifier := metric.NewPrometheusNotifier(metricsV2)
	proxyNotifier := proxynotifier.NewNotifierWebsocket(wsService)
	flagsetManager, err := service.NewFlagsetManager(proxyConf, log.ZapLogger, []notifier.Notifier{
		prometheusNotifier,
		proxyNotifier,
	}, nil)
	require.NoError(t, err)

	services := service.Services{
		MonitoringService: service.NewMonitoring(flagsetManager),
		WebsocketService:  wsService,
		FlagsetManager:    flagsetManager,
		Metrics:           metricsV2,
	}

	s := api.New(proxyConf, services, log.ZapLogger)
	go func() { s.StartWithContext(context.Background()) }()
	defer s.Stop(context.Background())

	baseURL := fmt.Sprintf("http://localhost:%d", port)
	waitForServer(t, baseURL)

	req, err := http.NewRequest("POST",
		baseURL+"/ofrep/v1/evaluate/flags",
		strings.NewReader(`{ "context":{"targetingKey":"some-key"}}`))
	require.NoError(t, err)
	req.Header.Add("Authorization", "Bearer test")
	req.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	assert.Equal(t, http.StatusOK, response.StatusCode)

	req, err = http.NewRequest("POST",
		baseURL+"/ofrep/v1/evaluate/flags/some-key",
		strings.NewReader(`{ "context":{"targetingKey":"some-key"}}`))
	require.NoError(t, err)
	req.Header.Add("Authorization", "Bearer test")
	req.Header.Add("Content-Type", "application/json")
	response, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	assert.Equal(t, http.StatusNotFound, response.StatusCode)

	req, err = http.NewRequest("POST",
		baseURL+"/ofrep/v1/evaluate/flags/test-flag",
		strings.NewReader(`{ "context":{"targetingKey":"some-key"}}`))
	require.NoError(t, err)
	req.Header.Add("Authorization", "Bearer test")
	req.Header.Add("Content-Type", "application/json")
	response, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	assert.Equal(t, http.StatusOK, response.StatusCode)
}

func Test_Middleware_VersionHeader_Enabled_Default(t *testing.T) {
	port := testutils.GetFreePort(t)
	proxyConf := &config.Config{
		CommonFlagSet: config.CommonFlagSet{
			Retrievers: &[]retrieverconf.RetrieverConf{
				{
					Kind: "file",
					Path: "../../../testdata/flag-config.yaml",
				},
			},
		},
		Server: config.Server{
			Mode: config.ServerModeHTTP,
			Port: port,
		},
	}
	log := newTestLogger(t)

	metricsV2, err := metric.NewMetrics()
	require.NoError(t, err)
	wsService := stream.NewWebsocketService()
	defer wsService.Close()
	flagsetManager, err := service.NewFlagsetManager(proxyConf, log.ZapLogger, nil, nil)
	require.NoError(t, err)

	services := service.Services{
		MonitoringService: service.NewMonitoring(flagsetManager),
		WebsocketService:  wsService,
		FlagsetManager:    flagsetManager,
		Metrics:           metricsV2,
	}

	s := api.New(proxyConf, services, log.ZapLogger)
	go func() { s.StartWithContext(context.Background()) }()
	defer s.Stop(context.Background())

	baseURL := fmt.Sprintf("http://localhost:%d", port)
	waitForServer(t, baseURL)

	response, err := http.Get(baseURL + "/health")
	require.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, proxyConf.Version, response.Header.Get("X-GOFEATUREFLAG-VERSION"))
}

func Test_VersionHeader_Disabled(t *testing.T) {
	port := testutils.GetFreePort(t)
	proxyConf := &config.Config{
		CommonFlagSet: config.CommonFlagSet{
			Retrievers: &[]retrieverconf.RetrieverConf{
				{
					Kind: "file",
					Path: "../../../testdata/flag-config.yaml",
				},
			},
		},
		Server: config.Server{
			Mode: config.ServerModeHTTP,
			Port: port,
		},
		DisableVersionHeader: true,
	}
	log := newTestLogger(t)

	metricsV2, err := metric.NewMetrics()
	require.NoError(t, err)
	wsService := stream.NewWebsocketService()
	defer wsService.Close()
	flagsetManager, err := service.NewFlagsetManager(proxyConf, log.ZapLogger, nil, nil)
	require.NoError(t, err)

	services := service.Services{
		MonitoringService: service.NewMonitoring(flagsetManager),
		WebsocketService:  wsService,
		FlagsetManager:    flagsetManager,
		Metrics:           metricsV2,
	}

	s := api.New(proxyConf, services, log.ZapLogger)
	go func() { s.StartWithContext(context.Background()) }()
	defer s.Stop(context.Background())

	baseURL := fmt.Sprintf("http://localhost:%d", port)
	waitForServer(t, baseURL)

	response, err := http.Get(baseURL + "/health")
	require.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Empty(t, response.Header.Get("X-GOFEATUREFLAG-VERSION"))
}

func Test_AuthenticationMiddleware(t *testing.T) {
	t.Run("Non Admin Endpoint", func(t *testing.T) {
		tests := []authMiddlewareTestCase{
			{
				name:          "Authentication disabled",
				configAPIKeys: nil,
				want:          http.StatusOK,
			},
			{
				name:          "Evaluation key provided",
				configAPIKeys: &config.APIKeys{Evaluation: []string{"test"}},
				want:          http.StatusUnauthorized,
			},
			{
				name:          "Admin key provided, no evaluation key provided",
				configAPIKeys: &config.APIKeys{Admin: []string{"test"}},
				want:          http.StatusOK,
			},
			{
				name:          "Evaluation and Admin key provided",
				configAPIKeys: &config.APIKeys{Evaluation: []string{"test"}, Admin: []string{"test"}},
				want:          http.StatusUnauthorized,
			},
		}

		runAuthMiddlewareTests(
			t,
			tests,
			func(_ *testing.T, _ authMiddlewareTestCase, baseURL string) (*http.Response, error) {
				return http.Post(baseURL+"/ofrep/v1/evaluate/flags/test-flag", "application/json",
					strings.NewReader(`{"context":{"targetingKey":"some-key"}}`),
				)
			})
	})

	t.Run("Admin Endpoint", func(t *testing.T) {
		tests := []authMiddlewareTestCase{
			{
				name:          "Authentication disabled",
				configAPIKeys: nil,
				want:          http.StatusUnauthorized,
			},
			{
				name:          "Evaluation key provided",
				configAPIKeys: &config.APIKeys{Evaluation: []string{"test"}},
				want:          http.StatusUnauthorized,
			},
			{
				name:          "Admin key provided, no evaluation key provided",
				configAPIKeys: &config.APIKeys{Admin: []string{"test"}},
				want:          http.StatusOK,
			},
			{
				name:          "Evaluation and Admin key provided",
				configAPIKeys: &config.APIKeys{Evaluation: []string{"test"}, Admin: []string{"test"}},
				want:          http.StatusOK,
			},
		}

		runAuthMiddlewareTests(
			t,
			tests,
			func(t *testing.T, tt authMiddlewareTestCase, baseURL string) (*http.Response, error) {
				request, err := http.NewRequest("POST", baseURL+"/admin/v1/retriever/refresh", nil)
				require.NoError(t, err)
				request.Header.Add("Content-Type", "application/json")
				if tt.configAPIKeys != nil && len(tt.configAPIKeys.Admin) > 0 {
					request.Header.Add("Authorization", "Bearer "+tt.configAPIKeys.Admin[0])
				}
				return http.DefaultClient.Do(request)
			})
	})
}

type authMiddlewareTestCase struct {
	name          string
	configAPIKeys *config.APIKeys
	want          int // http status code
}

// runAuthMiddlewareTests starts a relay proxy for each test case, issues the
// request returned by doRequest, and asserts the resulting HTTP status code.
// It holds the shared table-test setup so Test_AuthenticationMiddleware stays flat.
func runAuthMiddlewareTests(
	t *testing.T,
	tests []authMiddlewareTestCase,
	doRequest func(t *testing.T, tt authMiddlewareTestCase, baseURL string) (*http.Response, error),
) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port := testutils.GetFreePort(t)
			proxyConf := &config.Config{
				CommonFlagSet: config.CommonFlagSet{
					Retrievers: &[]retrieverconf.RetrieverConf{
						{
							Kind: "file",
							Path: "../../../testdata/flag-config.yaml",
						},
					},
				},
				Server: config.Server{
					Mode: config.ServerModeHTTP,
					Port: port,
				},
				DisableVersionHeader: true,
			}
			if tt.configAPIKeys != nil {
				proxyConf.AuthorizedKeys = *tt.configAPIKeys
			}
			proxyConf.ForceReloadAPIKeys()

			log := newTestLogger(t)

			metricsV2, err := metric.NewMetrics()
			require.NoError(t, err)
			wsService := stream.NewWebsocketService()
			defer wsService.Close()
			flagsetManager, err := service.NewFlagsetManager(proxyConf, log.ZapLogger, nil, nil)
			require.NoError(t, err)

			services := service.Services{
				MonitoringService: service.NewMonitoring(flagsetManager),
				WebsocketService:  wsService,
				FlagsetManager:    flagsetManager,
				Metrics:           metricsV2,
			}

			s := api.New(proxyConf, services, log.ZapLogger)
			go func() { s.StartWithContext(context.Background()) }()
			defer s.Stop(context.Background())

			baseURL := fmt.Sprintf("http://localhost:%d", port)
			waitForServer(t, baseURL)

			response, err := doRequest(t, tt, baseURL)
			require.NoError(t, err)
			defer func() { _ = response.Body.Close() }()
			assert.Equal(t, tt.want, response.StatusCode)
		})
	}
}

// skipUnixSocketOnWindows skips the unix socket tests on Windows. Binding an AF_UNIX listener
// there intermittently fails on the CI runners with WSAENETDOWN ("a socket operation encountered
// a dead network"), and unix socket mode is a Unix-only deployment target anyway.
func skipUnixSocketOnWindows(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("AF_UNIX listeners are unreliable on Windows CI (WSAENETDOWN); " +
			"unix socket mode is a Unix-only deployment target")
	}
}

// Helper function to create an HTTP client that can connect via Unix socket
func newUnixSocketHTTPClient(socketPath string) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		},
	}
}

func Test_Starting_RelayProxy_UnixSocket(t *testing.T) {
	skipUnixSocketOnWindows(t)
	// Create a temporary directory for the socket
	tempDir, err := os.MkdirTemp("", "goff-test-socket-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	socketPath := filepath.Join(tempDir, "goff-test.sock")

	proxyConf := &config.Config{
		CommonFlagSet: config.CommonFlagSet{
			Retrievers: &[]retrieverconf.RetrieverConf{
				{
					Kind: "file",
					Path: "../../../testdata/flag-config.yaml",
				},
			},
		},
		Server: config.Server{
			Mode:           config.ServerModeUnixSocket,
			UnixSocketPath: socketPath,
		},
	}
	log := newTestLogger(t)

	metricsV2, err := metric.NewMetrics()
	if err != nil {
		log.ZapLogger.Error("impossible to initialize prometheus metrics", zap.Error(err))
	}
	wsService := stream.NewWebsocketService()
	defer wsService.Close()
	prometheusNotifier := metric.NewPrometheusNotifier(metricsV2)
	proxyNotifier := proxynotifier.NewNotifierWebsocket(wsService)
	flagsetManager, err := service.NewFlagsetManager(proxyConf, log.ZapLogger, []notifier.Notifier{
		prometheusNotifier,
		proxyNotifier,
	}, nil)
	require.NoError(t, err)

	services := service.Services{
		MonitoringService: service.NewMonitoring(flagsetManager),
		WebsocketService:  wsService,
		FlagsetManager:    flagsetManager,
		Metrics:           metricsV2,
	}

	s := api.New(proxyConf, services, log.ZapLogger)
	go func() { s.StartWithContext(context.Background()) }()
	defer s.Stop(context.Background())

	// Wait for the socket to be created
	require.Eventually(t, func() bool {
		_, err := os.Stat(socketPath)
		return err == nil
	}, 1*time.Second, 10*time.Millisecond, "unix socket file was not created in time")

	// Verify socket file exists
	_, err = os.Stat(socketPath)
	require.NoError(t, err, "Unix socket file should exist")

	// Create a Unix socket HTTP client
	client := newUnixSocketHTTPClient(socketPath)

	// Test health endpoint
	response, err := client.Get("http://unix/health")
	require.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	assert.Equal(t, http.StatusOK, response.StatusCode)

	// Test metrics endpoint
	responseM, err := client.Get("http://unix/metrics")
	require.NoError(t, err)
	defer func() { _ = responseM.Body.Close() }()
	assert.Equal(t, http.StatusOK, responseM.StatusCode)

	// Test info endpoint
	responseI, err := client.Get("http://unix/info")
	require.NoError(t, err)
	defer func() { _ = responseI.Body.Close() }()
	assert.Equal(t, http.StatusOK, responseI.StatusCode)
}

func Test_Starting_RelayProxy_UnixSocket_MonitoringPort(t *testing.T) {
	skipUnixSocketOnWindows(t)
	monitoringPort := testutils.GetFreePort(t)
	// Create a temporary directory for the socket
	tempDir, err := os.MkdirTemp("", "goff-test-socket-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	socketPath := filepath.Join(tempDir, "goff-test.sock")

	proxyConf := &config.Config{
		CommonFlagSet: config.CommonFlagSet{
			Retrievers: &[]retrieverconf.RetrieverConf{
				{
					Kind: "file",
					Path: "../../../testdata/flag-config.yaml",
				},
			},
		},
		Server: config.Server{
			Mode:           config.ServerModeUnixSocket,
			UnixSocketPath: socketPath,
			MonitoringPort: monitoringPort,
		},
	}
	log := newTestLogger(t)

	metricsV2, err := metric.NewMetrics()
	if err != nil {
		log.ZapLogger.Error("impossible to initialize prometheus metrics", zap.Error(err))
	}
	wsService := stream.NewWebsocketService()
	defer wsService.Close()
	prometheusNotifier := metric.NewPrometheusNotifier(metricsV2)
	proxyNotifier := proxynotifier.NewNotifierWebsocket(wsService)
	flagsetManager, err := service.NewFlagsetManager(proxyConf, log.ZapLogger, []notifier.Notifier{
		prometheusNotifier,
		proxyNotifier,
	}, nil)
	require.NoError(t, err)

	services := service.Services{
		MonitoringService: service.NewMonitoring(flagsetManager),
		WebsocketService:  wsService,
		FlagsetManager:    flagsetManager,
		Metrics:           metricsV2,
	}

	s := api.New(proxyConf, services, log.ZapLogger)
	go func() { s.StartWithContext(context.Background()) }()
	defer s.Stop(context.Background())

	// Wait for the socket to be created
	require.Eventually(t, func() bool {
		_, err := os.Stat(socketPath)
		return err == nil
	}, 1*time.Second, 10*time.Millisecond, "unix socket file was not created in time")

	// Verify socket file exists
	_, err = os.Stat(socketPath)
	require.NoError(t, err, "Unix socket file should exist")

	monitoringURL := fmt.Sprintf("http://localhost:%d", monitoringPort)
	waitForServer(t, monitoringURL)

	// Test health endpoint
	responseH1, err := http.Get(monitoringURL + "/health")
	require.NoError(t, err)
	defer responseH1.Body.Close()
	assert.Equal(t, http.StatusOK, responseH1.StatusCode)

	// Test metrics endpoint
	responseM1, err := http.Get(monitoringURL + "/metrics")
	require.NoError(t, err)
	defer responseM1.Body.Close()
	assert.Equal(t, http.StatusOK, responseM1.StatusCode)

	// Test info endpoint
	responseI, err := http.Get(monitoringURL + "/info")
	require.NoError(t, err)
	defer responseI.Body.Close()
	assert.Equal(t, http.StatusOK, responseI.StatusCode)
}

func Test_Starting_RelayProxy_UnixSocket_OFREP_API(t *testing.T) {
	skipUnixSocketOnWindows(t)
	// Create a temporary directory for the socket
	tempDir, err := os.MkdirTemp("", "goff-test-socket-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	socketPath := filepath.Join(tempDir, "goff-test-ofrep.sock")

	proxyConf := &config.Config{
		CommonFlagSet: config.CommonFlagSet{
			Retrievers: &[]retrieverconf.RetrieverConf{
				{
					Kind: "file",
					Path: "../../../testdata/flag-config.yaml",
				},
			},
		},
		Server: config.Server{
			Mode:           config.ServerModeUnixSocket,
			UnixSocketPath: socketPath,
		},
		AuthorizedKeys: config.APIKeys{
			Admin:      nil,
			Evaluation: []string{"test"},
		},
	}
	log := newTestLogger(t)

	metricsV2, err := metric.NewMetrics()
	if err != nil {
		log.ZapLogger.Error("impossible to initialize prometheus metrics", zap.Error(err))
	}
	wsService := stream.NewWebsocketService()
	defer wsService.Close()
	prometheusNotifier := metric.NewPrometheusNotifier(metricsV2)
	proxyNotifier := proxynotifier.NewNotifierWebsocket(wsService)
	flagsetManager, err := service.NewFlagsetManager(proxyConf, log.ZapLogger, []notifier.Notifier{
		prometheusNotifier,
		proxyNotifier,
	}, nil)
	require.NoError(t, err)

	services := service.Services{
		MonitoringService: service.NewMonitoring(flagsetManager),
		WebsocketService:  wsService,
		FlagsetManager:    flagsetManager,
		Metrics:           metricsV2,
	}

	s := api.New(proxyConf, services, log.ZapLogger)
	go func() { s.StartWithContext(context.Background()) }()
	defer s.Stop(context.Background())

	// Wait for the socket to be created
	require.Eventually(t, func() bool {
		_, err := os.Stat(socketPath)
		return err == nil
	}, 1*time.Second, 10*time.Millisecond, "unix socket file was not created in time")

	// Verify socket file exists
	_, err = os.Stat(socketPath)
	require.NoError(t, err, "Unix socket file should exist")

	// Create a Unix socket HTTP client
	client := newUnixSocketHTTPClient(socketPath)

	// Test OFREP evaluate all flags endpoint
	req, err := http.NewRequest("POST",
		"http://unix/ofrep/v1/evaluate/flags",
		strings.NewReader(`{ "context":{"targetingKey":"some-key"}}`))
	require.NoError(t, err)
	req.Header.Add("Authorization", "Bearer test")
	req.Header.Add("Content-Type", "application/json")
	response, err := client.Do(req)
	require.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	assert.Equal(t, http.StatusOK, response.StatusCode)

	// Test OFREP evaluate specific flag endpoint (non-existent flag)
	req, err = http.NewRequest("POST",
		"http://unix/ofrep/v1/evaluate/flags/some-key",
		strings.NewReader(`{ "context":{"targetingKey":"some-key"}}`))
	require.NoError(t, err)
	req.Header.Add("Authorization", "Bearer test")
	req.Header.Add("Content-Type", "application/json")
	response, err = client.Do(req)
	require.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	assert.Equal(t, http.StatusNotFound, response.StatusCode)

	// Test OFREP evaluate specific flag endpoint (existing flag)
	req, err = http.NewRequest("POST",
		"http://unix/ofrep/v1/evaluate/flags/test-flag",
		strings.NewReader(`{ "context":{"targetingKey":"some-key"}}`))
	require.NoError(t, err)
	req.Header.Add("Authorization", "Bearer test")
	req.Header.Add("Content-Type", "application/json")
	response, err = client.Do(req)
	require.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	assert.Equal(t, http.StatusOK, response.StatusCode)
}

func TestStartingRelayProxyUnixSocketAuthentication(t *testing.T) {
	skipUnixSocketOnWindows(t)
	tests := []struct {
		name          string
		configAPIKeys *config.APIKeys
		endpoint      string
		method        string
		body          string
		authHeader    string
		want          int // http status code
	}{
		{
			name:          "Authentication disabled - health endpoint",
			configAPIKeys: nil,
			endpoint:      "http://unix/health",
			method:        "GET",
			want:          http.StatusOK,
		},
		{
			name:          "Evaluation endpoint - with valid key",
			configAPIKeys: &config.APIKeys{Evaluation: []string{"test-key"}},
			endpoint:      "http://unix/ofrep/v1/evaluate/flags/test-flag",
			method:        "POST",
			body:          `{"context":{"targetingKey":"some-key"}}`,
			authHeader:    "Bearer test-key",
			want:          http.StatusOK,
		},
		{
			name:          "Evaluation endpoint - without key (should fail)",
			configAPIKeys: &config.APIKeys{Evaluation: []string{"test-key"}},
			endpoint:      "http://unix/ofrep/v1/evaluate/flags/test-flag",
			method:        "POST",
			body:          `{"context":{"targetingKey":"some-key"}}`,
			authHeader:    "",
			want:          http.StatusUnauthorized,
		},
		{
			name:          "Admin endpoint - with valid admin key",
			configAPIKeys: &config.APIKeys{Admin: []string{"admin-key"}},
			endpoint:      "http://unix/admin/v1/retriever/refresh",
			method:        "POST",
			authHeader:    "Bearer admin-key",
			want:          http.StatusOK,
		},
		{
			name:          "Admin endpoint - without admin key (should fail)",
			configAPIKeys: &config.APIKeys{Admin: []string{"admin-key"}},
			endpoint:      "http://unix/admin/v1/retriever/refresh",
			method:        "POST",
			authHeader:    "",
			want:          http.StatusUnauthorized,
		},
	}

	for index, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for the socket
			tempDir, err := os.MkdirTemp("", "")
			require.NoError(t, err)
			defer os.RemoveAll(tempDir)

			socketPath := filepath.Join(tempDir, fmt.Sprintf("goff-test-%d.sock", index))

			proxyConf := &config.Config{
				CommonFlagSet: config.CommonFlagSet{
					Retrievers: &[]retrieverconf.RetrieverConf{
						{
							Kind: "file",
							Path: "../../../testdata/flag-config.yaml",
						},
					},
				},
				Server: config.Server{
					Mode:           config.ServerModeUnixSocket,
					UnixSocketPath: socketPath,
				},
			}
			if tt.configAPIKeys != nil {
				proxyConf.AuthorizedKeys = *tt.configAPIKeys
			}
			proxyConf.ForceReloadAPIKeys()

			log := newTestLogger(t)

			metricsV2, err := metric.NewMetrics()
			require.NoError(t, err)
			wsService := stream.NewWebsocketService()
			defer wsService.Close()
			flagsetManager, err := service.NewFlagsetManager(proxyConf, log.ZapLogger, nil, nil)
			require.NoError(t, err)

			services := service.Services{
				MonitoringService: service.NewMonitoring(flagsetManager),
				WebsocketService:  wsService,
				FlagsetManager:    flagsetManager,
				Metrics:           metricsV2,
			}

			s := api.New(proxyConf, services, log.ZapLogger)
			go func() { s.StartWithContext(context.Background()) }()
			defer s.Stop(context.Background())

			// Wait for the socket to be created
			require.Eventually(t, func() bool {
				_, err := os.Stat(socketPath)
				return err == nil
			}, 1*time.Second, 10*time.Millisecond, "unix socket file was not created in time")

			// Create a Unix socket HTTP client
			client := newUnixSocketHTTPClient(socketPath)

			// Create and execute request
			var req *http.Request
			if tt.body != "" {
				req, err = http.NewRequest(tt.method, tt.endpoint, strings.NewReader(tt.body))
			} else {
				req, err = http.NewRequest(tt.method, tt.endpoint, nil)
			}
			require.NoError(t, err)

			if tt.authHeader != "" {
				req.Header.Add("Authorization", tt.authHeader)
			}
			if tt.body != "" {
				req.Header.Add("Content-Type", "application/json")
			}

			response, err := client.Do(req)
			require.NoError(t, err)
			defer func() { _ = response.Body.Close() }()
			assert.Equal(t, tt.want, response.StatusCode)
		})
	}
}

func TestStartingRelayProxyUnixSocketVersionHeader(t *testing.T) {
	skipUnixSocketOnWindows(t)
	tests := []struct {
		name                 string
		disableVersionHeader bool
		wantVersionHeader    bool
	}{
		{
			name:                 "Version header enabled by default",
			disableVersionHeader: false,
			wantVersionHeader:    true,
		},
		{
			name:                 "Version header disabled",
			disableVersionHeader: true,
			wantVersionHeader:    false,
		},
	}

	for index, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for the socket
			tempDir, err := os.MkdirTemp("", "")
			require.NoError(t, err)
			defer os.RemoveAll(tempDir)

			socketPath := filepath.Join(tempDir, fmt.Sprintf("goff-test-version-%d.sock", index))

			proxyConf := &config.Config{
				CommonFlagSet: config.CommonFlagSet{
					Retrievers: &[]retrieverconf.RetrieverConf{
						{
							Kind: "file",
							Path: "../../../testdata/flag-config.yaml",
						},
					},
				},
				Server: config.Server{
					Mode:           config.ServerModeUnixSocket,
					UnixSocketPath: socketPath,
				},
				DisableVersionHeader: tt.disableVersionHeader,
				Version:              "test-version-1.0.0",
			}

			log := newTestLogger(t)

			metricsV2, err := metric.NewMetrics()
			require.NoError(t, err)
			wsService := stream.NewWebsocketService()
			defer wsService.Close()
			flagsetManager, err := service.NewFlagsetManager(proxyConf, log.ZapLogger, nil, nil)
			require.NoError(t, err)

			services := service.Services{
				MonitoringService: service.NewMonitoring(flagsetManager),
				WebsocketService:  wsService,
				FlagsetManager:    flagsetManager,
				Metrics:           metricsV2,
			}

			s := api.New(proxyConf, services, log.ZapLogger)
			go func() { s.StartWithContext(context.Background()) }()
			defer s.Stop(context.Background())

			// Wait for the socket to be created
			require.Eventually(t, func() bool {
				_, err := os.Stat(socketPath)
				return err == nil
			}, 1*time.Second, 10*time.Millisecond, "unix socket file was not created in time")

			// Create a Unix socket HTTP client
			client := newUnixSocketHTTPClient(socketPath)

			response, err := client.Get("http://unix/health")
			require.NoError(t, err)
			defer func() { _ = response.Body.Close() }()
			assert.Equal(t, http.StatusOK, response.StatusCode)

			if tt.wantVersionHeader {
				assert.Equal(t, "test-version-1.0.0", response.Header.Get("X-GOFEATUREFLAG-VERSION"))
			} else {
				assert.Empty(t, response.Header.Get("X-GOFEATUREFLAG-VERSION"))
			}
		})
	}
}

func Test_NativeHistograms_Enabled(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)

	port := testutils.GetFreePort(t)
	proxyConf := &config.Config{
		CommonFlagSet: config.CommonFlagSet{
			Retrievers: &[]retrieverconf.RetrieverConf{
				{
					Kind: "file",
					Path: "../../../testdata/flag-config.yaml",
				},
			},
		},
		Server: config.Server{
			Mode: config.ServerModeHTTP,
			Port: port,
		},
	}

	log := newTestLogger(t)

	metricsV2, err := metric.NewMetrics()
	require.NoError(t, err)

	wsService := stream.NewWebsocketService()
	defer wsService.Close()

	prometheusNotifier := metric.NewPrometheusNotifier(metricsV2)
	proxyNotifier := proxynotifier.NewNotifierWebsocket(wsService)

	flagsetManager, err := service.NewFlagsetManager(proxyConf, log.ZapLogger, []notifier.Notifier{
		prometheusNotifier,
		proxyNotifier,
	}, nil)
	require.NoError(t, err)

	services := service.Services{
		MonitoringService: service.NewMonitoring(flagsetManager),
		WebsocketService:  wsService,
		FlagsetManager:    flagsetManager,
		Metrics:           metricsV2,
	}

	s := api.New(proxyConf, services, log.ZapLogger)
	go func() { s.StartWithContext(ctx) }()
	defer s.Stop(ctx)

	baseURL := fmt.Sprintf("http://localhost:%d", port)
	waitForServer(t, baseURL)

	// Make some requests to generate histogram data
	healthResp, err := http.Get(baseURL + "/health")
	require.NoError(t, err)
	_ = healthResp.Body.Close()

	infoResp, err := http.Get(baseURL + "/info")
	require.NoError(t, err)
	_ = infoResp.Body.Close()

	// Scrape /metrics with protobuf Accept header
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/metrics", nil)
	require.NoError(t, err)
	req.Header.Set("Accept", "application/vnd.google.protobuf;proto=io.prometheus.client.MetricFamily;encoding=delimited")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Decode delimited protobuf stream
	var histogramFamilies []*promdto.MetricFamily
	reader := bufio.NewReader(resp.Body)
	for {
		var mf promdto.MetricFamily
		err := protodelim.UnmarshalFrom(reader, &mf)
		if errors.Is(err, io.EOF) {
			break
		}
		require.NoError(t, err)

		// Only check histogram metric families from our subsystem
		if mf.GetType() == promdto.MetricType_HISTOGRAM {
			if strings.HasPrefix(mf.GetName(), "gofeatureflag_") {
				histogramFamilies = append(histogramFamilies, &mf)
			}
		}
	}

	// Assert we found at least one histogram metric family
	require.NotEmpty(t, histogramFamilies)

	// Check each histogram has both classic buckets and native histogram fields
	for _, mf := range histogramFamilies {
		for _, metric := range mf.GetMetric() {
			h := metric.GetHistogram()
			if h == nil {
				continue
			}

			// Assert classic histogram buckets are still present
			assert.NotEmpty(t, h.GetBucket())

			// Assert native histogram data is present
			// Check for native histogram fields: PositiveSpan, NegativeSpan, or Schema
			hasNativeData := len(h.GetPositiveSpan()) > 0 || len(h.GetNegativeSpan()) > 0 || h.Schema != nil
			assert.True(t, hasNativeData)
		}
	}
}

func Test_PortFreedAfterShutdown(t *testing.T) {
	// Pick a free port.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	proxyConf := &config.Config{
		CommonFlagSet: config.CommonFlagSet{
			Retrievers: &[]retrieverconf.RetrieverConf{
				{
					Kind: "file",
					Path: "../../../testdata/flag-config.yaml",
				},
			},
		},
		Server: config.Server{
			Mode: config.ServerModeHTTP,
			Port: port,
		},
	}

	l := newTestLogger(t)

	wsService := stream.NewWebsocketService()

	flagsetManager, err := service.NewFlagsetManager(proxyConf, l.ZapLogger, nil, nil)
	require.NoError(t, err)

	services := service.Services{
		MonitoringService: service.NewMonitoring(flagsetManager),
		WebsocketService:  wsService,
		FlagsetManager:    flagsetManager,
	}

	s := api.New(proxyConf, services, l.ZapLogger)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		s.StartWithContext(ctx)
		close(done)
	}()

	// Wait for the server to be ready.
	require.Eventually(t, func() bool {
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/health", port))
		if err != nil {
			return false
		}
		_ = resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}, 5*time.Second, 50*time.Millisecond)

	// Cancel the context to simulate Ctrl+C.
	cancel()
	wsService.Close()

	// Wait for StartWithContext to return.
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("server did not stop within 10 seconds")
	}

	// Verify the port is free by binding to it again.
	ln2, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	require.NoError(t, err, "port %d should be free after shutdown", port)
	ln2.Close()
}

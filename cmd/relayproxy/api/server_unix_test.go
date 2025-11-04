package api_test

import (
	"context"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/api"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/log"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/retrieverconf"
	"github.com/thomaspoignant/go-feature-flag/notifier"
)

func Test_UnixSocket_BasicEndpoints(t *testing.T) {
	tests := []struct {
		name         string
		socketPath   string
		testEndpoint string
		wantStatus   int
	}{
		{
			name:         "Unix socket with health endpoint",
			socketPath:   "/tmp/goff-test-health.sock",
			testEndpoint: "/health",
			wantStatus:   http.StatusOK,
		},
		{
			name:         "Unix socket with info endpoint",
			socketPath:   "/tmp/goff-test-info.sock",
			testEndpoint: "/info",
			wantStatus:   http.StatusOK,
		},
		{
			name:         "Unix socket with metrics endpoint",
			socketPath:   "/tmp/goff-test-metrics.sock",
			testEndpoint: "/metrics",
			wantStatus:   http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing socket
			_ = os.Remove(tt.socketPath)
			defer os.Remove(tt.socketPath)

			proxyConf := &config.Config{
				CommonFlagSet: config.CommonFlagSet{
					Retrievers: &[]retrieverconf.RetrieverConf{
						{
							Kind: "file",
							Path: "../../../testdata/flag-config.yaml",
						},
					},
				},
				UnixSocket: tt.socketPath,
			}

			logger := log.InitLogger()
			defer func() { _ = logger.ZapLogger.Sync() }()

			metricsV2, err := metric.NewMetrics()
			require.NoError(t, err)

			wsService := service.NewWebsocketService()
			defer wsService.Close()

			prometheusNotifier := metric.NewPrometheusNotifier(metricsV2)
			proxyNotifier := service.NewNotifierWebsocket(wsService)
			flagsetManager, err := service.NewFlagsetManager(proxyConf, logger.ZapLogger, []notifier.Notifier{
				prometheusNotifier,
				proxyNotifier,
			})
			require.NoError(t, err)

			services := service.Services{
				MonitoringService: service.NewMonitoring(flagsetManager),
				WebsocketService:  wsService,
				FlagsetManager:    flagsetManager,
				Metrics:           metricsV2,
			}

			s := api.New(proxyConf, services, logger.ZapLogger)
			go func() { s.Start() }()
			defer s.Stop(context.Background())

			// Wait for socket to be created
			time.Sleep(100 * time.Millisecond)

			// Verify socket file exists
			_, err = os.Stat(tt.socketPath)
			assert.NoError(t, err, "Socket file should exist")

			// Create HTTP client that uses Unix socket
			client := &http.Client{
				Transport: &http.Transport{
					DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
						return net.Dial("unix", tt.socketPath)
					},
				},
			}

			// Test the endpoint
			resp, err := client.Get("http://unix" + tt.testEndpoint)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func Test_UnixSocket_WithTCPPort(t *testing.T) {
	socketPath := "/tmp/goff-test-both.sock"
	_ = os.Remove(socketPath)
	defer os.Remove(socketPath)

	proxyConf := &config.Config{
		CommonFlagSet: config.CommonFlagSet{
			Retrievers: &[]retrieverconf.RetrieverConf{
				{
					Kind: "file",
					Path: "../../../testdata/flag-config.yaml",
				},
			},
		},
		ListenPort: 11030,
		UnixSocket: socketPath,
	}

	logger := log.InitLogger()
	defer func() { _ = logger.ZapLogger.Sync() }()

	metricsV2, err := metric.NewMetrics()
	require.NoError(t, err)

	wsService := service.NewWebsocketService()
	defer wsService.Close()

	prometheusNotifier := metric.NewPrometheusNotifier(metricsV2)
	proxyNotifier := service.NewNotifierWebsocket(wsService)
	flagsetManager, err := service.NewFlagsetManager(proxyConf, logger.ZapLogger, []notifier.Notifier{
		prometheusNotifier,
		proxyNotifier,
	})
	require.NoError(t, err)

	services := service.Services{
		MonitoringService: service.NewMonitoring(flagsetManager),
		WebsocketService:  wsService,
		FlagsetManager:    flagsetManager,
		Metrics:           metricsV2,
	}

	s := api.New(proxyConf, services, logger.ZapLogger)
	go func() { s.Start() }()
	defer s.Stop(context.Background())

	time.Sleep(100 * time.Millisecond)

	// Test TCP endpoint
	respTCP, err := http.Get("http://localhost:11030/health")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, respTCP.StatusCode)

	// Test Unix socket endpoint
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		},
	}

	respUnix, err := client.Get("http://unix/health")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, respUnix.StatusCode)
}

func Test_UnixSocket_CleanupOnStop(t *testing.T) {
	socketPath := "/tmp/goff-test-cleanup.sock"
	_ = os.Remove(socketPath)

	proxyConf := &config.Config{
		CommonFlagSet: config.CommonFlagSet{
			Retrievers: &[]retrieverconf.RetrieverConf{
				{
					Kind: "file",
					Path: "../../../testdata/flag-config.yaml",
				},
			},
		},
		UnixSocket: socketPath,
	}

	logger := log.InitLogger()
	defer func() { _ = logger.ZapLogger.Sync() }()

	metricsV2, err := metric.NewMetrics()
	require.NoError(t, err)

	wsService := service.NewWebsocketService()
	defer wsService.Close()

	prometheusNotifier := metric.NewPrometheusNotifier(metricsV2)
	proxyNotifier := service.NewNotifierWebsocket(wsService)
	flagsetManager, err := service.NewFlagsetManager(proxyConf, logger.ZapLogger, []notifier.Notifier{
		prometheusNotifier,
		proxyNotifier,
	})
	require.NoError(t, err)

	services := service.Services{
		MonitoringService: service.NewMonitoring(flagsetManager),
		WebsocketService:  wsService,
		FlagsetManager:    flagsetManager,
		Metrics:           metricsV2,
	}

	s := api.New(proxyConf, services, logger.ZapLogger)
	go func() { s.Start() }()

	time.Sleep(100 * time.Millisecond)

	// Verify socket exists
	_, err = os.Stat(socketPath)
	assert.NoError(t, err, "Socket file should exist before stop")

	// Stop the server
	s.Stop(context.Background())

	// Verify socket is removed
	_, err = os.Stat(socketPath)
	assert.True(t, os.IsNotExist(err), "Socket file should be removed after stop")
}

func Test_UnixSocket_InSubdirectory(t *testing.T) {
	tmpDir := t.TempDir()
	socketPath := filepath.Join(tmpDir, "subdir", "goff.sock")

	// Create subdirectory
	err := os.MkdirAll(filepath.Dir(socketPath), 0755)
	require.NoError(t, err)

	defer os.Remove(socketPath)

	proxyConf := &config.Config{
		CommonFlagSet: config.CommonFlagSet{
			Retrievers: &[]retrieverconf.RetrieverConf{
				{
					Kind: "file",
					Path: "../../../testdata/flag-config.yaml",
				},
			},
		},
		UnixSocket: socketPath,
	}

	logger := log.InitLogger()
	defer func() { _ = logger.ZapLogger.Sync() }()

	metricsV2, err := metric.NewMetrics()
	require.NoError(t, err)

	wsService := service.NewWebsocketService()
	defer wsService.Close()

	prometheusNotifier := metric.NewPrometheusNotifier(metricsV2)
	proxyNotifier := service.NewNotifierWebsocket(wsService)
	flagsetManager, err := service.NewFlagsetManager(proxyConf, logger.ZapLogger, []notifier.Notifier{
		prometheusNotifier,
		proxyNotifier,
	})
	require.NoError(t, err)

	services := service.Services{
		MonitoringService: service.NewMonitoring(flagsetManager),
		WebsocketService:  wsService,
		FlagsetManager:    flagsetManager,
		Metrics:           metricsV2,
	}

	s := api.New(proxyConf, services, logger.ZapLogger)
	go func() { s.Start() }()
	defer s.Stop(context.Background())

	time.Sleep(100 * time.Millisecond)

	// Verify socket file exists
	_, err = os.Stat(socketPath)
	assert.NoError(t, err, "Socket file should exist in subdirectory")

	// Test connection
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		},
	}

	resp, err := client.Get("http://unix/health")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func Test_UnixSocket_OFREPEndpoint(t *testing.T) {
	socketPath := "/tmp/goff-test-ofrep.sock"
	_ = os.Remove(socketPath)
	defer os.Remove(socketPath)

	proxyConf := &config.Config{
		CommonFlagSet: config.CommonFlagSet{
			Retrievers: &[]retrieverconf.RetrieverConf{
				{
					Kind: "file",
					Path: "../../../testdata/flag-config.yaml",
				},
			},
		},
		UnixSocket: socketPath,
		AuthorizedKeys: config.APIKeys{
			Evaluation: []string{"test-key"},
		},
	}

	logger := log.InitLogger()
	defer func() { _ = logger.ZapLogger.Sync() }()

	metricsV2, err := metric.NewMetrics()
	require.NoError(t, err)

	wsService := service.NewWebsocketService()
	defer wsService.Close()

	prometheusNotifier := metric.NewPrometheusNotifier(metricsV2)
	proxyNotifier := service.NewNotifierWebsocket(wsService)
	flagsetManager, err := service.NewFlagsetManager(proxyConf, logger.ZapLogger, []notifier.Notifier{
		prometheusNotifier,
		proxyNotifier,
	})
	require.NoError(t, err)

	services := service.Services{
		MonitoringService: service.NewMonitoring(flagsetManager),
		WebsocketService:  wsService,
		FlagsetManager:    flagsetManager,
		Metrics:           metricsV2,
	}

	s := api.New(proxyConf, services, logger.ZapLogger)
	go func() { s.Start() }()
	defer s.Stop(context.Background())

	time.Sleep(100 * time.Millisecond)

	// Create HTTP client that uses Unix socket
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		},
	}

	// Test OFREP endpoint
	req, err := http.NewRequest("POST",
		"http://unix/ofrep/v1/evaluate/flags/test-flag",
		strings.NewReader(`{"context":{"targetingKey":"some-key"}}`))
	require.NoError(t, err)
	req.Header.Add("Authorization", "Bearer test-key")
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

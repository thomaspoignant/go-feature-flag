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
	"strings"
	"testing"
	"time"

	promdto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/api"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/log"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/retrieverconf"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protodelim"
)

func Test_Starting_RelayProxy_with_monitoring_on_same_port(t *testing.T) {
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
			Port: 11024,
		},
	}
	log := log.InitLogger()
	defer func() { _ = log.ZapLogger.Sync() }()

	metricsV2, err := metric.NewMetrics()
	if err != nil {
		log.ZapLogger.Error("impossible to initialize prometheus metrics", zap.Error(err))
	}
	wsService := service.NewWebsocketService()
	defer wsService.Close() // close all the open connections
	prometheusNotifier := metric.NewPrometheusNotifier(metricsV2)
	proxyNotifier := service.NewNotifierWebsocket(wsService)
	flagsetManager, err := service.NewFlagsetManager(proxyConf, log.ZapLogger, []notifier.Notifier{
		prometheusNotifier,
		proxyNotifier,
	})
	assert.NoError(t, err)

	services := service.Services{
		MonitoringService: service.NewMonitoring(flagsetManager),
		WebsocketService:  wsService,
		FlagsetManager:    flagsetManager,
		Metrics:           metricsV2,
	}

	s := api.New(proxyConf, services, log.ZapLogger)
	go func() { s.StartWithContext(context.Background()) }()
	defer s.Stop(context.Background())

	time.Sleep(10 * time.Millisecond)

	response, err := http.Get("http://localhost:11024/health")
	assert.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	assert.Equal(t, http.StatusOK, response.StatusCode)

	responseM, err := http.Get("http://localhost:11024/metrics")
	assert.NoError(t, err)
	defer func() { _ = responseM.Body.Close() }()
	assert.Equal(t, http.StatusOK, responseM.StatusCode)

	responseI, err := http.Get("http://localhost:11024/info")
	assert.NoError(t, err)
	defer func() { _ = responseI.Body.Close() }()
	assert.Equal(t, http.StatusOK, responseI.StatusCode)
}

func Test_Starting_RelayProxy_with_monitoring_on_different_port(t *testing.T) {
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
			Port:           11024,
			MonitoringPort: 11025,
		},
	}
	log := log.InitLogger()
	defer func() { _ = log.ZapLogger.Sync() }()

	metricsV2, err := metric.NewMetrics()
	if err != nil {
		log.ZapLogger.Error("impossible to initialize prometheus metrics", zap.Error(err))
	}
	wsService := service.NewWebsocketService()
	defer wsService.Close() // close all the open connections
	prometheusNotifier := metric.NewPrometheusNotifier(metricsV2)
	proxyNotifier := service.NewNotifierWebsocket(wsService)
	flagsetManager, err := service.NewFlagsetManager(proxyConf, log.ZapLogger, []notifier.Notifier{
		prometheusNotifier,
		proxyNotifier,
	})
	assert.NoError(t, err)

	services := service.Services{
		MonitoringService: service.NewMonitoring(flagsetManager),
		WebsocketService:  wsService,
		FlagsetManager:    flagsetManager,
		Metrics:           metricsV2,
	}

	s := api.New(proxyConf, services, log.ZapLogger)
	go func() { s.StartWithContext(context.Background()) }()
	defer s.Stop(context.Background())

	time.Sleep(10 * time.Millisecond)

	response, err := http.Get("http://localhost:11024/health")
	assert.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	assert.Equal(t, http.StatusNotFound, response.StatusCode)

	responseM, err := http.Get("http://localhost:11024/metrics")
	assert.NoError(t, err)
	defer func() { _ = responseM.Body.Close() }()
	assert.Equal(t, http.StatusNotFound, responseM.StatusCode)

	responseI, err := http.Get("http://localhost:11024/info")
	assert.NoError(t, err)
	defer func() { _ = responseI.Body.Close() }()
	assert.Equal(t, http.StatusNotFound, responseI.StatusCode)

	responseH1, err := http.Get("http://localhost:11025/health")
	assert.NoError(t, err)
	defer func() { _ = responseH1.Body.Close() }()
	assert.Equal(t, http.StatusOK, responseH1.StatusCode)

	responseM1, err := http.Get("http://localhost:11025/metrics")
	assert.NoError(t, err)
	defer func() { _ = responseM1.Body.Close() }()
	assert.Equal(t, http.StatusOK, responseM1.StatusCode)

	responseI1, err := http.Get("http://localhost:11025/info")
	assert.NoError(t, err)
	defer func() { _ = responseI1.Body.Close() }()
	assert.Equal(t, http.StatusOK, responseI1.StatusCode)
}

func Test_CheckOFREPAPIExists(t *testing.T) {
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
			Port: 11024,
		},
		AuthorizedKeys: config.APIKeys{
			Admin:      nil,
			Evaluation: []string{"test"},
		},
	}
	log := log.InitLogger()
	defer func() { _ = log.ZapLogger.Sync() }()

	metricsV2, err := metric.NewMetrics()
	if err != nil {
		log.ZapLogger.Error("impossible to initialize prometheus metrics", zap.Error(err))
	}
	wsService := service.NewWebsocketService()
	defer wsService.Close() // close all the open connections
	prometheusNotifier := metric.NewPrometheusNotifier(metricsV2)
	proxyNotifier := service.NewNotifierWebsocket(wsService)
	flagsetManager, err := service.NewFlagsetManager(proxyConf, log.ZapLogger, []notifier.Notifier{
		prometheusNotifier,
		proxyNotifier,
	})
	assert.NoError(t, err)

	services := service.Services{
		MonitoringService: service.NewMonitoring(flagsetManager),
		WebsocketService:  wsService,
		FlagsetManager:    flagsetManager,
		Metrics:           metricsV2,
	}

	s := api.New(proxyConf, services, log.ZapLogger)
	go func() { s.StartWithContext(context.Background()) }()
	defer s.Stop(context.Background())

	time.Sleep(10 * time.Millisecond)

	req, err := http.NewRequest("POST",
		"http://localhost:11024/ofrep/v1/evaluate/flags",
		strings.NewReader(`{ "context":{"targetingKey":"some-key"}}`))
	assert.NoError(t, err)
	req.Header.Add("Authorization", "Bearer test")
	req.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	assert.Equal(t, http.StatusOK, response.StatusCode)

	req, err = http.NewRequest("POST",
		"http://localhost:11024/ofrep/v1/evaluate/flags/some-key",
		strings.NewReader(`{ "context":{"targetingKey":"some-key"}}`))
	assert.NoError(t, err)
	req.Header.Add("Authorization", "Bearer test")
	req.Header.Add("Content-Type", "application/json")
	response, err = http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	assert.Equal(t, http.StatusNotFound, response.StatusCode)

	req, err = http.NewRequest("POST",
		"http://localhost:11024/ofrep/v1/evaluate/flags/test-flag",
		strings.NewReader(`{ "context":{"targetingKey":"some-key"}}`))
	assert.NoError(t, err)
	req.Header.Add("Authorization", "Bearer test")
	req.Header.Add("Content-Type", "application/json")
	response, err = http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	assert.Equal(t, http.StatusOK, response.StatusCode)
}

func Test_Middleware_VersionHeader_Enabled_Default(t *testing.T) {
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
			Port: 11024,
		},
	}
	log := log.InitLogger()
	defer func() { _ = log.ZapLogger.Sync() }()

	metricsV2, err := metric.NewMetrics()
	require.NoError(t, err)
	wsService := service.NewWebsocketService()
	defer wsService.Close()
	flagsetManager, err := service.NewFlagsetManager(proxyConf, log.ZapLogger, nil)
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

	time.Sleep(10 * time.Millisecond)

	response, err := http.Get("http://localhost:11024/health")
	assert.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, proxyConf.Version, response.Header.Get("X-GOFEATUREFLAG-VERSION"))
}

func Test_VersionHeader_Disabled(t *testing.T) {
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
			Port: 11024,
		},
		DisableVersionHeader: true,
	}
	log := log.InitLogger()
	defer func() { _ = log.ZapLogger.Sync() }()

	metricsV2, err := metric.NewMetrics()
	require.NoError(t, err)
	wsService := service.NewWebsocketService()
	defer wsService.Close()
	flagsetManager, err := service.NewFlagsetManager(proxyConf, log.ZapLogger, nil)
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

	time.Sleep(10 * time.Millisecond)

	response, err := http.Get("http://localhost:11024/health")
	assert.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Empty(t, response.Header.Get("X-GOFEATUREFLAG-VERSION"))
}

func Test_AuthenticationMiddleware(t *testing.T) {
	t.Run("Non Admin Endpoint", func(t *testing.T) {
		tests := []struct {
			name          string
			configAPIKeys *config.APIKeys
			want          int // http status code
		}{
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

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
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
						Port: 11024,
					},
					DisableVersionHeader: true,
				}
				if tt.configAPIKeys != nil {
					proxyConf.AuthorizedKeys = *tt.configAPIKeys
				}
				proxyConf.ForceReloadAPIKeys()

				log := log.InitLogger()
				defer func() { _ = log.ZapLogger.Sync() }()

				metricsV2, err := metric.NewMetrics()
				require.NoError(t, err)
				wsService := service.NewWebsocketService()
				defer wsService.Close()
				flagsetManager, err := service.NewFlagsetManager(proxyConf, log.ZapLogger, nil)
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
				time.Sleep(10 * time.Millisecond)

				response, err := http.Post("http://localhost:11024/ofrep/v1/evaluate/flags/test-flag", "application/json",
					strings.NewReader(`{"context":{"targetingKey":"some-key"}}`),
				)
				assert.NoError(t, err)
				defer func() { _ = response.Body.Close() }()
				assert.Equal(t, tt.want, response.StatusCode)
			})
		}
	})

	t.Run("Admin Endpoint", func(t *testing.T) {
		tests := []struct {
			name          string
			configAPIKeys *config.APIKeys
			want          int // http status code
		}{
			{
				name:          "Authentication disabled",
				configAPIKeys: nil,
				want:          http.StatusBadRequest,
			},
			{
				name:          "Evaluation key provided",
				configAPIKeys: &config.APIKeys{Evaluation: []string{"test"}},
				want:          http.StatusBadRequest,
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

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
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
						Port: 11024,
					},
					DisableVersionHeader: true,
				}
				if tt.configAPIKeys != nil {
					proxyConf.AuthorizedKeys = *tt.configAPIKeys
				}
				proxyConf.ForceReloadAPIKeys()

				log := log.InitLogger()
				defer func() { _ = log.ZapLogger.Sync() }()

				metricsV2, err := metric.NewMetrics()
				require.NoError(t, err)
				wsService := service.NewWebsocketService()
				defer wsService.Close()
				flagsetManager, err := service.NewFlagsetManager(proxyConf, log.ZapLogger, nil)
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
				time.Sleep(10 * time.Millisecond)

				request, err := http.NewRequest("POST", "http://localhost:11024/admin/v1/retriever/refresh", nil)
				assert.NoError(t, err)
				request.Header.Add("Content-Type", "application/json")
				if tt.configAPIKeys != nil && len(tt.configAPIKeys.Admin) > 0 {
					request.Header.Add("Authorization", "Bearer "+tt.configAPIKeys.Admin[0])
				}
				response, err := http.DefaultClient.Do(request)
				assert.NoError(t, err)
				defer func() { _ = response.Body.Close() }()
				assert.Equal(t, tt.want, response.StatusCode)
			})
		}
	})
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
	log := log.InitLogger()
	defer func() { _ = log.ZapLogger.Sync() }()

	metricsV2, err := metric.NewMetrics()
	if err != nil {
		log.ZapLogger.Error("impossible to initialize prometheus metrics", zap.Error(err))
	}
	wsService := service.NewWebsocketService()
	defer wsService.Close()
	prometheusNotifier := metric.NewPrometheusNotifier(metricsV2)
	proxyNotifier := service.NewNotifierWebsocket(wsService)
	flagsetManager, err := service.NewFlagsetManager(proxyConf, log.ZapLogger, []notifier.Notifier{
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
	assert.NoError(t, err, "Unix socket file should exist")

	// Create a Unix socket HTTP client
	client := newUnixSocketHTTPClient(socketPath)

	// Test health endpoint
	response, err := client.Get("http://unix/health")
	assert.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	assert.Equal(t, http.StatusOK, response.StatusCode)

	// Test metrics endpoint
	responseM, err := client.Get("http://unix/metrics")
	assert.NoError(t, err)
	defer func() { _ = responseM.Body.Close() }()
	assert.Equal(t, http.StatusOK, responseM.StatusCode)

	// Test info endpoint
	responseI, err := client.Get("http://unix/info")
	assert.NoError(t, err)
	defer func() { _ = responseI.Body.Close() }()
	assert.Equal(t, http.StatusOK, responseI.StatusCode)
}

func Test_Starting_RelayProxy_UnixSocket_MonitoringPort(t *testing.T) {
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
			MonitoringPort: 11026,
		},
	}
	log := log.InitLogger()
	defer func() { _ = log.ZapLogger.Sync() }()

	metricsV2, err := metric.NewMetrics()
	if err != nil {
		log.ZapLogger.Error("impossible to initialize prometheus metrics", zap.Error(err))
	}
	wsService := service.NewWebsocketService()
	defer wsService.Close()
	prometheusNotifier := metric.NewPrometheusNotifier(metricsV2)
	proxyNotifier := service.NewNotifierWebsocket(wsService)
	flagsetManager, err := service.NewFlagsetManager(proxyConf, log.ZapLogger, []notifier.Notifier{
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
	assert.NoError(t, err, "Unix socket file should exist")

	// Test health endpoint
	responseH1, err := http.Get("http://localhost:11026/health")
	assert.NoError(t, err)
	defer responseH1.Body.Close()
	assert.Equal(t, http.StatusOK, responseH1.StatusCode)

	// Test metrics endpoint
	responseM1, err := http.Get("http://localhost:11026/metrics")
	assert.NoError(t, err)
	defer responseM1.Body.Close()
	assert.Equal(t, http.StatusOK, responseM1.StatusCode)

	// Test info endpoint
	responseI, err := http.Get("http://localhost:11026/info")
	assert.NoError(t, err)
	defer responseI.Body.Close()
	assert.Equal(t, http.StatusOK, responseI.StatusCode)
}

func Test_Starting_RelayProxy_UnixSocket_OFREP_API(t *testing.T) {
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
	log := log.InitLogger()
	defer func() { _ = log.ZapLogger.Sync() }()

	metricsV2, err := metric.NewMetrics()
	if err != nil {
		log.ZapLogger.Error("impossible to initialize prometheus metrics", zap.Error(err))
	}
	wsService := service.NewWebsocketService()
	defer wsService.Close()
	prometheusNotifier := metric.NewPrometheusNotifier(metricsV2)
	proxyNotifier := service.NewNotifierWebsocket(wsService)
	flagsetManager, err := service.NewFlagsetManager(proxyConf, log.ZapLogger, []notifier.Notifier{
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
	assert.NoError(t, err, "Unix socket file should exist")

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
	assert.NoError(t, err)
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
	assert.NoError(t, err)
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
	assert.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	assert.Equal(t, http.StatusOK, response.StatusCode)
}

func TestStartingRelayProxyUnixSocketAuthentication(t *testing.T) {
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
			want:          http.StatusBadRequest, // Returns 400 when auth is required but not provided
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

			log := log.InitLogger()
			defer func() { _ = log.ZapLogger.Sync() }()

			metricsV2, err := metric.NewMetrics()
			require.NoError(t, err)
			wsService := service.NewWebsocketService()
			defer wsService.Close()
			flagsetManager, err := service.NewFlagsetManager(proxyConf, log.ZapLogger, nil)
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
			assert.NoError(t, err)
			defer func() { _ = response.Body.Close() }()
			assert.Equal(t, tt.want, response.StatusCode)
		})
	}
}

func TestStartingRelayProxyUnixSocketVersionHeader(t *testing.T) {
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

			log := log.InitLogger()
			defer func() { _ = log.ZapLogger.Sync() }()

			metricsV2, err := metric.NewMetrics()
			require.NoError(t, err)
			wsService := service.NewWebsocketService()
			defer wsService.Close()
			flagsetManager, err := service.NewFlagsetManager(proxyConf, log.ZapLogger, nil)
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
			assert.NoError(t, err)
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)

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
			Port: 11024,
		},
	}

	log := log.InitLogger()
	defer func() { _ = log.ZapLogger.Sync() }()

	metricsV2, err := metric.NewMetrics()
	require.NoError(t, err)

	wsService := service.NewWebsocketService()
	defer wsService.Close()

	prometheusNotifier := metric.NewPrometheusNotifier(metricsV2)
	proxyNotifier := service.NewNotifierWebsocket(wsService)

	flagsetManager, err := service.NewFlagsetManager(proxyConf, log.ZapLogger, []notifier.Notifier{
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

	s := api.New(proxyConf, services, log.ZapLogger)
	go func() { s.StartWithContext(ctx) }()
	defer s.Stop(ctx)

	time.Sleep(10 * time.Millisecond)

	// Make some requests to generate histogram data
	healthResp, err := http.Get("http://localhost:11024/health")
	require.NoError(t, err)
	_ = healthResp.Body.Close()

	infoResp, err := http.Get("http://localhost:11024/info")
	require.NoError(t, err)
	_ = infoResp.Body.Close()

	// Scrape /metrics with protobuf Accept header
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:11024/metrics", nil)
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

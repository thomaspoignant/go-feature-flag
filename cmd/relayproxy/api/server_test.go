package api_test

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/api"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/log"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"go.uber.org/zap"
)

func Test_Starting_RelayProxy_with_monitoring_on_same_port(t *testing.T) {
	proxyConf := &config.Config{
		CommonFlagSet: config.CommonFlagSet{
			Retrievers: &[]config.RetrieverConf{
				{
					Kind: "file",
					Path: "../../../testdata/flag-config.yaml",
				},
			},
		},
		ListenPort: 11024,
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
	if err != nil {
		panic(err)
	}

	services := service.Services{
		MonitoringService: service.NewMonitoring(flagsetManager),
		WebsocketService:  wsService,
		FlagsetManager:    flagsetManager,
		Metrics:           metricsV2,
	}

	s := api.New(proxyConf, services, log.ZapLogger)
	go func() { s.Start() }()
	defer s.Stop(context.Background())

	time.Sleep(10 * time.Millisecond)

	response, err := http.Get("http://localhost:11024/health")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)

	responseM, err := http.Get("http://localhost:11024/metrics")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, responseM.StatusCode)

	responseI, err := http.Get("http://localhost:11024/info")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, responseI.StatusCode)
}

func Test_Starting_RelayProxy_with_monitoring_on_different_port(t *testing.T) {
	proxyConf := &config.Config{
		CommonFlagSet: config.CommonFlagSet{
			Retrievers: &[]config.RetrieverConf{
				{
					Kind: "file",
					Path: "../../../testdata/flag-config.yaml",
				},
			},
		},
		ListenPort:     11024,
		MonitoringPort: 11025,
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
	if err != nil {
		panic(err)
	}

	services := service.Services{
		MonitoringService: service.NewMonitoring(flagsetManager),
		WebsocketService:  wsService,
		FlagsetManager:    flagsetManager,
		Metrics:           metricsV2,
	}

	s := api.New(proxyConf, services, log.ZapLogger)
	go func() { s.Start() }()
	defer s.Stop(context.Background())

	time.Sleep(10 * time.Millisecond)

	response, err := http.Get("http://localhost:11024/health")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, response.StatusCode)

	responseM, err := http.Get("http://localhost:11024/metrics")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, responseM.StatusCode)

	responseI, err := http.Get("http://localhost:11024/info")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, responseI.StatusCode)

	responseH1, err := http.Get("http://localhost:11025/health")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, responseH1.StatusCode)

	responseM1, err := http.Get("http://localhost:11025/metrics")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, responseM1.StatusCode)

	responseI1, err := http.Get("http://localhost:11025/info")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, responseI1.StatusCode)
}

func Test_CheckOFREPAPIExists(t *testing.T) {
	proxyConf := &config.Config{
		CommonFlagSet: config.CommonFlagSet{
			Retrievers: &[]config.RetrieverConf{
				{
					Kind: "file",
					Path: "../../../testdata/flag-config.yaml",
				},
			},
		},
		ListenPort: 11024,
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
	if err != nil {
		panic(err)
	}

	services := service.Services{
		MonitoringService: service.NewMonitoring(flagsetManager),
		WebsocketService:  wsService,
		FlagsetManager:    flagsetManager,
		Metrics:           metricsV2,
	}

	s := api.New(proxyConf, services, log.ZapLogger)
	go func() { s.Start() }()
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
	assert.Equal(t, http.StatusOK, response.StatusCode)

	req, err = http.NewRequest("POST",
		"http://localhost:11024/ofrep/v1/evaluate/flags/some-key",
		strings.NewReader(`{ "context":{"targetingKey":"some-key"}}`))
	assert.NoError(t, err)
	req.Header.Add("Authorization", "Bearer test")
	req.Header.Add("Content-Type", "application/json")
	response, err = http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, response.StatusCode)

	req, err = http.NewRequest("POST",
		"http://localhost:11024/ofrep/v1/evaluate/flags/test-flag",
		strings.NewReader(`{ "context":{"targetingKey":"some-key"}}`))
	assert.NoError(t, err)
	req.Header.Add("Authorization", "Bearer test")
	req.Header.Add("Content-Type", "application/json")
	response, err = http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)
}

func Test_Middleware_VersionHeader_Enabled_Default(t *testing.T) {
	proxyConf := &config.Config{
		CommonFlagSet: config.CommonFlagSet{
			Retrievers: &[]config.RetrieverConf{
				{
					Kind: "file",
					Path: "../../../testdata/flag-config.yaml",
				},
			},
		},
		ListenPort: 11024,
	}
	log := log.InitLogger()
	defer func() { _ = log.ZapLogger.Sync() }()

	metricsV2, _ := metric.NewMetrics()
	wsService := service.NewWebsocketService()
	defer wsService.Close()
	flagsetManager, _ := service.NewFlagsetManager(proxyConf, log.ZapLogger, nil)

	services := service.Services{
		MonitoringService: service.NewMonitoring(flagsetManager),
		WebsocketService:  wsService,
		FlagsetManager:    flagsetManager,
		Metrics:           metricsV2,
	}

	s := api.New(proxyConf, services, log.ZapLogger)
	go func() { s.Start() }()
	defer s.Stop(context.Background())

	time.Sleep(10 * time.Millisecond)

	response, err := http.Get("http://localhost:11024/health")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, proxyConf.Version, response.Header.Get("X-GOFEATUREFLAG-VERSION"))
}

func Test_VersionHeader_Disabled(t *testing.T) {
	proxyConf := &config.Config{
		CommonFlagSet: config.CommonFlagSet{
			Retrievers: &[]config.RetrieverConf{
				{
					Kind: "file",
					Path: "../../../testdata/flag-config.yaml",
				},
			},
		},
		ListenPort:           11024,
		DisableVersionHeader: true,
	}
	log := log.InitLogger()
	defer func() { _ = log.ZapLogger.Sync() }()

	metricsV2, _ := metric.NewMetrics()
	wsService := service.NewWebsocketService()
	defer wsService.Close()
	flagsetManager, _ := service.NewFlagsetManager(proxyConf, log.ZapLogger, nil)

	services := service.Services{
		MonitoringService: service.NewMonitoring(flagsetManager),
		WebsocketService:  wsService,
		FlagsetManager:    flagsetManager,
		Metrics:           metricsV2,
	}

	s := api.New(proxyConf, services, log.ZapLogger)
	go func() { s.Start() }()
	defer s.Stop(context.Background())

	time.Sleep(10 * time.Millisecond)

	response, err := http.Get("http://localhost:11024/health")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Empty(t, response.Header.Get("X-GOFEATUREFLAG-VERSION"))
}

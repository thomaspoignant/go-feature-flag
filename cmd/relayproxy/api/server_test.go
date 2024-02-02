package api_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/api"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/log"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"go.uber.org/zap"
	"net/http"
	"testing"
	"time"
)

func Test_Starting_RelayProxy_with_monitoring_on_same_port(t *testing.T) {
	proxyConf := &config.Config{
		Retriever: &config.RetrieverConf{
			Kind: "file",
			Path: "../../../testdata/flag-config.yaml",
		},
		ListenPort: 11024,
	}
	zapLog := log.InitLogger()
	defer func() { _ = zapLog.Sync() }()

	metricsV2, err := metric.NewMetrics()
	if err != nil {
		zapLog.Error("impossible to initialize prometheus metrics", zap.Error(err))
	}
	wsService := service.NewWebsocketService()
	defer wsService.Close() // close all the open connections
	prometheusNotifier := metric.NewPrometheusNotifier(metricsV2)
	proxyNotifier := service.NewNotifierWebsocket(wsService)
	goff, err := service.NewGoFeatureFlagClient(proxyConf, zapLog, []notifier.Notifier{
		prometheusNotifier,
		proxyNotifier,
	})
	if err != nil {
		panic(err)
	}

	services := service.Services{
		MonitoringService:    service.NewMonitoring(goff),
		WebsocketService:     wsService,
		GOFeatureFlagService: goff,
		Metrics:              metricsV2,
	}

	s := api.New(proxyConf, services, zapLog)
	go func() { s.Start() }()
	defer s.Stop()

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
		Retriever: &config.RetrieverConf{
			Kind: "file",
			Path: "../../../testdata/flag-config.yaml",
		},
		ListenPort:     11024,
		MonitoringPort: 11025,
	}
	zapLog := log.InitLogger()
	defer func() { _ = zapLog.Sync() }()

	metricsV2, err := metric.NewMetrics()
	if err != nil {
		zapLog.Error("impossible to initialize prometheus metrics", zap.Error(err))
	}
	wsService := service.NewWebsocketService()
	defer wsService.Close() // close all the open connections
	prometheusNotifier := metric.NewPrometheusNotifier(metricsV2)
	proxyNotifier := service.NewNotifierWebsocket(wsService)
	goff, err := service.NewGoFeatureFlagClient(proxyConf, zapLog, []notifier.Notifier{
		prometheusNotifier,
		proxyNotifier,
	})
	if err != nil {
		panic(err)
	}

	services := service.Services{
		MonitoringService:    service.NewMonitoring(goff),
		WebsocketService:     wsService,
		GOFeatureFlagService: goff,
		Metrics:              metricsV2,
	}

	s := api.New(proxyConf, services, zapLog)
	go func() { s.Start() }()
	defer s.Stop()

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

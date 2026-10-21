package main

import (
	"context"
	"crypto/fips140"
	"fmt"
	"os"
	"time"

	promversion "github.com/prometheus/common/version"
	"github.com/spf13/pflag"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/api"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/docs"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/proxynotifier"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	stream "github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service/stream"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/log"
	"github.com/thomaspoignant/go-feature-flag/internal/flaghistory"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"go.uber.org/zap"
)

// version is overridden by GoReleaser during the build.
var version = "localdev"

const banner = `█▀▀ █▀█   █▀▀ █▀▀ ▄▀█ ▀█▀ █ █ █▀█ █▀▀   █▀▀ █   ▄▀█ █▀▀
█▄█ █▄█   █▀  ██▄ █▀█  █  █▄█ █▀▄ ██▄   █▀  █▄▄ █▀█ █▄█

     █▀█ █▀▀ █   ▄▀█ █▄█   █▀█ █▀█ █▀█ ▀▄▀ █▄█
     █▀▄ ██▄ █▄▄ █▀█  █    █▀▀ █▀▄ █▄█ █ █  █ 

GO Feature Flag Relay Proxy - Version %s
_____________________________________________`

// @title GO Feature Flag relay proxy endpoints
// @description.markdown
// @contact.name GO feature flag relay proxy
// @contact.url https://gofeatureflag.org
// @contact.email contact@gofeatureflag.org
// @license.name MIT
// @license.url https://github.com/thomaspoignant/go-feature-flag/blob/main/LICENSE
// @x-logo {"url":"https://raw.githubusercontent.com/thomaspoignant/go-feature-flag/main/logo_128.png"}
// @BasePath /
// @securitydefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description Use configured APIKeys in yaml config as authorization keys, disabled when this yaml config is not set.
// @securitydefinitions.apikey XApiKeyAuth
// @in header
// @name X-API-Key
// @description Use configured APIKeys in yaml config as authorization keys via X-API-Key header,
// @description disabled when this yaml config is not set.
func main() {
	// Init pFlag for config file
	f := pflag.NewFlagSet("config", pflag.ContinueOnError)
	f.String("config", "", "Location of your config file")
	_ = f.Parse(os.Args[1:])

	// Init logger
	logger := log.InitLogger()
	defer func() { _ = logger.ZapLogger.Sync() }()

	// Loading the configuration in viper
	proxyConf, err := config.New(f, logger.ZapLogger, version)
	if err != nil {
		logger.ZapLogger.Fatal("error while reading configuration", zap.Error(err))
	}
	defer func() {
		if err := proxyConf.StopConfigChangeWatcher(); err != nil {
			logger.ZapLogger.Error("error while stopping the configuration watcher", zap.Error(err))
		}
	}()

	if err := proxyConf.IsValid(); err != nil {
		logger.ZapLogger.Fatal("configuration error", zap.Error(err))
	}

	if !proxyConf.HideBanner {
		fmt.Printf(banner+"\n", version)
	}

	// Update the logger's format and level from the config
	logger.Update(proxyConf.LogFormat, proxyConf.ZapLogLevel())

	// Surface whether the binary is running with FIPS 140-3 validated crypto.
	// True only for builds produced with GOFIPS140 (see `make build-relayproxy-fips`).
	logger.ZapLogger.Info("crypto mode", zap.Bool("fips140", fips140.Enabled()))

	// Init swagger
	docs.SwaggerInfo.Version = proxyConf.Version
	docs.SwaggerInfo.Host = fmt.Sprintf("%s:%d", proxyConf.SwaggerHost(), proxyConf.ServerPort(logger.ZapLogger))

	// Set the version for the prometheus version collector
	promversion.Version = version
	// Initialize metrics
	metricsV2, err := metric.NewMetrics(metric.MetricsOpts{
		EnableBulkMetricFlagNames: proxyConf.EnableBulkMetricFlagNames,
	})
	if err != nil {
		logger.ZapLogger.Error("impossible to initialize prometheus metrics", zap.Error(err))
	}

	// Init services
	wsService := stream.NewWebsocketService(stream.WithFlagDetails(!proxyConf.DisableFlagDetailsInStream))
	sseService := stream.NewSSEService()
	prometheusNotifier := metric.NewPrometheusNotifier(metricsV2)
	proxyNotifier := proxynotifier.NewNotifierWebsocket(wsService)

	flagHistory := newFlagHistory(proxyConf, logger.ZapLogger)
	flagsetManager, err := service.NewFlagsetManager(proxyConf, logger.ZapLogger, []notifier.Notifier{
		prometheusNotifier,
		proxyNotifier,
	}, sseService, service.WithFlagHistory(flagHistory))
	if err != nil {
		if flagHistory != nil {
			flagHistory.Close()
		}
		logger.ZapLogger.Error(
			"impossible to start GO Feature Flag, we are not able to initialize the retrieval of flags",
			zap.Error(err),
		)
		return
	}
	services := service.Services{
		MonitoringService: service.NewMonitoring(flagsetManager),
		WebsocketService:  wsService,
		SSEService:        sseService,
		FlagsetManager:    flagsetManager,
		Metrics:           metricsV2,
		FlagHistory:       flagHistory,
	}
	// Init API server
	apiServer := api.New(proxyConf, services, logger.ZapLogger)

	defer shutdown(logger.ZapLogger, &apiServer, wsService, sseService, flagsetManager, flagHistory)
	apiServer.StartWithContext(context.Background())
}

// shutdown stops the relay proxy: streaming clients disconnect first, then the HTTP server stops,
// and finally the flagsets are closed.
func shutdown(
	logger *zap.Logger,
	apiServer interface{ Stop(ctx context.Context) },
	wsService stream.WebsocketService,
	sseService stream.SSEService,
	flagsetManager service.FlagsetManager,
	flagHistory *flaghistory.History,
) {
	logger.Info("Stopping API server")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	wsService.Close()
	sseService.Close()
	apiServer.Stop(ctx)
	// Closing the flagsets flushes the exporters, once the server no longer accepts requests.
	flagsetManager.Close()
	// The history is closed last: closing the flagsets can still produce notifications.
	if flagHistory != nil {
		flagHistory.Close()
	}
}

// newFlagHistory starts the history of the flag modifications, nil when it is disabled.
// The relay proxy starts even if the database is down, the modifications are kept in memory
// and written as soon as it is back.
func newFlagHistory(proxyConf *config.Config, logger *zap.Logger) *flaghistory.History {
	c := proxyConf.FlagHistory
	if !c.IsEnabled() {
		if c.URI == "" && (c.Enabled == nil || *c.Enabled) {
			logger.Info("flag history is inactive: set flagHistory.uri (PostgreSQL) to record the flag modifications")
		}
		return nil
	}
	source, err := os.Hostname()
	if err != nil {
		source = "unknown"
	}
	h := flaghistory.New(flaghistory.Config{
		Connect: func(ctx context.Context) (flaghistory.Store, error) {
			return flaghistory.NewPostgresStore(ctx, c.URI, c.Table)
		},
		Retention: c.RetentionDuration(),
		Source:    source,
		QueueSize: c.QueueSize,
		Logger:    logger,
	})
	h.Start()
	return h
}

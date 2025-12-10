package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	promversion "github.com/prometheus/common/version"
	"github.com/spf13/pflag"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/api"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/docs"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/log"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
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

	if err := proxyConf.IsValid(); err != nil {
		logger.ZapLogger.Fatal("configuration error", zap.Error(err))
	}

	if !proxyConf.HideBanner {
		fmt.Printf(banner+"\n", version)
	}

	// Update the logger's format and level from the config
	logger.Update(proxyConf.LogFormat, proxyConf.ZapLogLevel())

	// Init swagger
	docs.SwaggerInfo.Version = proxyConf.Version
	docs.SwaggerInfo.Host = fmt.Sprintf("%s:%d", proxyConf.Host, proxyConf.GetServerPort(logger.ZapLogger))

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
	wsService := service.NewWebsocketService()
	defer wsService.Close() // close all the open connections
	prometheusNotifier := metric.NewPrometheusNotifier(metricsV2)
	proxyNotifier := service.NewNotifierWebsocket(wsService)

	flagsetManager, err := service.NewFlagsetManager(proxyConf, logger.ZapLogger, []notifier.Notifier{
		prometheusNotifier,
		proxyNotifier,
	})
	if err != nil {
		logger.ZapLogger.Error(
			"impossible to start GO Feature Flag, we are not able to initialize the retrieval of flags",
			zap.Error(err),
		)
		return
	}

	services := service.Services{
		MonitoringService: service.NewMonitoring(flagsetManager),
		WebsocketService:  wsService,
		FlagsetManager:    flagsetManager,
		Metrics:           metricsV2,
	}
	// Init API server
	apiServer := api.New(proxyConf, services, logger.ZapLogger)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		apiServer.Stop(ctx)
	}()

	// Start config file watcher if flagsets are configured
	if len(proxyConf.FlagSets) > 0 {
		configFilePath, err := config.GetConfigFilePath()
		if err == nil {
			startConfigWatcher(configFilePath, f, flagsetManager, logger.ZapLogger, version, []notifier.Notifier{
				prometheusNotifier,
				proxyNotifier,
			})
		} else {
			logger.ZapLogger.Warn("could not start config file watcher", zap.Error(err))
		}
	}

	apiServer.StartWithContext(context.Background())
}

// startConfigWatcher starts a file watcher that monitors the configuration file for changes
// and reloads flagsets when the file is modified.
func startConfigWatcher(
	configFilePath string,
	flagSet *pflag.FlagSet,
	flagsetManager service.FlagsetManager,
	logger *zap.Logger,
	version string,
	notifiers []notifier.Notifier,
) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Error("failed to create file watcher", zap.Error(err))
		return
	}
	defer func() {
		_ = watcher.Close()
	}()

	// Watch the directory containing the config file
	configDir := filepath.Dir(configFilePath)
	if err := watcher.Add(configDir); err != nil {
		logger.Error("failed to watch config directory", zap.String("dir", configDir), zap.Error(err))
		return
	}

	logger.Info("watching configuration file for changes", zap.String("file", configFilePath))

	// Use a debounce mechanism to avoid multiple reloads for rapid file changes
	var reloadTimer *time.Timer
	var mu sync.Mutex

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// Only process write and rename events for the config file
				if (event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Rename == fsnotify.Rename) &&
					event.Name == configFilePath {
					mu.Lock()
					if reloadTimer != nil {
						reloadTimer.Stop()
					}
					// Debounce: wait 500ms before reloading to handle rapid file changes
					reloadTimer = time.AfterFunc(500*time.Millisecond, func() {
						reloadFlagsets(configFilePath, flagSet, flagsetManager, logger, version, notifiers)
					})
					mu.Unlock()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logger.Error("file watcher error", zap.Error(err))
			}
		}
	}()
}

// reloadFlagsets reloads the configuration file and updates flagsets
func reloadFlagsets(
	configFilePath string,
	flagSet *pflag.FlagSet,
	flagsetManager service.FlagsetManager,
	logger *zap.Logger,
	version string,
	notifiers []notifier.Notifier,
) {
	logger.Info("configuration file changed, reloading flagsets", zap.String("file", configFilePath))

	// Reload configuration from file
	newConfig, err := config.ReloadFromFile(flagSet, logger, version)
	if err != nil {
		logger.Error("failed to reload configuration file", zap.Error(err))
		return
	}

	// Validate configuration
	if err := newConfig.IsValid(); err != nil {
		logger.Error("reloaded configuration is invalid", zap.Error(err))
		return
	}

	// Reload flagsets
	if err := flagsetManager.ReloadFlagsets(newConfig, logger, notifiers); err != nil {
		logger.Error("failed to reload flagsets", zap.Error(err))
		return
	}

	logger.Info("flagsets reloaded successfully")
}

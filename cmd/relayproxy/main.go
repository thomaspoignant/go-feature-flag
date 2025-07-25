package main

import (
	"context"
	"fmt"
	"os"
	"time"

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

// version, releaseDate are override by the makefile during the build.
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
	docs.SwaggerInfo.Host = fmt.Sprintf("%s:%d", proxyConf.Host, proxyConf.ListenPort)

	// Init services
	metricsV2, err := metric.NewMetrics()
	if err != nil {
		logger.ZapLogger.Error("impossible to initialize prometheus metrics", zap.Error(err))
	}
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

	if proxyConf.StartAsAwsLambda {
		apiServer.StartAwsLambda()
	} else {
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			apiServer.Stop(ctx)
		}()
		apiServer.Start()
	}
}

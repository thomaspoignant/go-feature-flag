package main

import (
	"flag"
	"fmt"

	"github.com/spf13/pflag"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/api"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/docs"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/log"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"go.uber.org/zap"
)

// version, releaseDate are override by the makefile during the build.
var version = "localdev"

const banner = `█▀▀ █▀█   █▀▀ █▀▀ ▄▀█ ▀█▀ █ █ █▀█ █▀▀   █▀▀ █   ▄▀█ █▀▀
█▄█ █▄█   █▀  ██▄ █▀█  █  █▄█ █▀▄ ██▄   █▀  █▄▄ █▀█ █▄█

     █▀█ █▀▀ █   ▄▀█ █▄█   █▀█ █▀█ █▀█ ▀▄▀ █▄█
     █▀▄ ██▄ █▄▄ █▀█  █    █▀▀ █▀▄ █▄█ █ █  █ 

GO Feature Flag Relay Proxy
_____________________________________________`

// @title go-feature-flag relay proxy
// @description Swagger for the go-feature-flag relay proxy.
// @description
// @description go-feature-flag relay proxy is using thomaspoignant/go-feature-flag to handle your feature flagging.
// @description It is a proxy to your flags, you can get the values of your flags using APIs.
// @contact.name GO feature flag relay proxy
// @contact.url https://github.com/thomaspoignant/go-feature-flag-relay-proxy
// @license.name MIT
// @license.url https://github.com/thomaspoignant/go-feature-flag/blob/main/LICENSE
// @BasePath /
func main() {
	// Init pFlag for config file
	flag.String("config", "", "Location of your config file")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	// Init logger
	zapLog := log.InitLogger()
	defer func() { _ = zapLog.Sync() }()

	// Loading the configuration in viper
	proxyConf, err := config.ParseConfig(zapLog, version)
	if err != nil {
		zapLog.Fatal("error while reading configuration", zap.Error(err))
	}

	if err := proxyConf.IsValid(); err != nil {
		zapLog.Fatal("configuration error", zap.Error(err))
	}

	if !proxyConf.HideBanner {
		fmt.Println(banner)
	}

	// Init swagger
	docs.SwaggerInfo.Version = proxyConf.Version
	docs.SwaggerInfo.Host = fmt.Sprintf("%s:%d", proxyConf.Host, proxyConf.ListenPort)

	// Init services
	goff, err := service.NewGoFeatureFlagClient(proxyConf, zapLog)
	if err != nil {
		panic(err)
	}

	monitoringService := service.NewMonitoring(goff)

	// Init API server
	apiServer := api.New(proxyConf, monitoringService, goff, zapLog)
	apiServer.Start()
	defer func() { _ = apiServer.Stop }()
}

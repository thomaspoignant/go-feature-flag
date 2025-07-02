package service

import (
	goff "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
)

type Services struct {
	// MonitoringService is the service in charge of managing the monitoring
	MonitoringService Monitoring
	// WebsocketBroadcasterService is the service in charge to manage the websockets in the relay proxy
	WebsocketService WebsocketService
	// GOFeatureFlagService is the GO Feature Flag goff we are using in the relay proxy.
	// deprecated: use FlagsetManager instead
	GOFeatureFlagService *goff.GoFeatureFlag
	// Metrics is the service containing all the prometheus metrics
	Metrics metric.Metrics
	// FlagsetManager is the service in charge of managing the flagsets
	FlagsetManager FlagsetManager
}

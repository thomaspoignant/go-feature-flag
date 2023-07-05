package service

import goff "github.com/thomaspoignant/go-feature-flag"

type Services struct {
	// MonitoringService is the service in charge of managing the monitoring
	MonitoringService Monitoring
	// WebsocketBroadcasterService is the service in charge to manage the websockets
	WebsocketService WebsocketService
	// GOFeatureFlagService is the GO Feature Flag client we are using
	GOFeatureFlagService *goff.GoFeatureFlag
}

package service

import goff "github.com/thomaspoignant/go-feature-flag"

type Services struct {
	// MonitoringService is the service in charge of managing the monitoring
	MonitoringService Monitoring
	// WebsocketBroadcasterService is the service in charge to manage the websockets in the relay proxy
	WebsocketService WebsocketService
	// GOFeatureFlagService is the GO Feature Flag goff we are using in the relay proxy.
	GOFeatureFlagService *goff.GoFeatureFlag
}

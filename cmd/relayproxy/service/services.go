package service

import (
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service/stream"
)

type Services struct {
	// MonitoringService is the service in charge of managing the monitoring
	MonitoringService Monitoring
	// WebsocketBroadcasterService is the service in charge to manage the websockets in the relay proxy
	WebsocketService stream.WebsocketService
	// SSEService is the service in charge of managing SSE connections in the relay proxy
	SSEService stream.SSEService
	// Metrics is the service containing all the prometheus metrics
	Metrics metric.Metrics
	// FlagsetManager is the service in charge of managing the flagsets
	FlagsetManager FlagsetManager
}

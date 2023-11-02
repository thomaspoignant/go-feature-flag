package service

import (
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/notifier"
)

type notifierPrometheus struct {
	metricsService metric.Metrics
}

func NewNotifierRelayProxy(metricsService metric.Metrics) notifier.Notifier {
	return &notifierPrometheus{
		metricsService: metricsService,
	}
}

func (n *notifierPrometheus) Notify(diff notifier.DiffCache) error {

	// flag changed counter
	// flag deleted counter
	// flag added counter
	// flag file

	n.websocketService.BroadcastFlagChanges(diff)
	return nil
}

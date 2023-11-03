package metric

import (
	"github.com/thomaspoignant/go-feature-flag/notifier"
)

type notifierPrometheus struct {
	metricsService Metrics
}

func NewPrometheusNotifier(metricsService Metrics) notifier.Notifier {
	return &notifierPrometheus{
		metricsService: metricsService,
	}
}

func (n *notifierPrometheus) Notify(diff notifier.DiffCache) error {
	if !diff.HasDiff() {
		return nil
	}
	n.metricsService.IncFlagChange()
	for flagName := range diff.Deleted {
		n.metricsService.IncFlagDeleted(flagName)
	}

	for flagName := range diff.Added {
		n.metricsService.IncFlagCreated(flagName)
	}

	for flagName := range diff.Updated {
		n.metricsService.IncFlagUpdated(flagName)
	}
	return nil
}

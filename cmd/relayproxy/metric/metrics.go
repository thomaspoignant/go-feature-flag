package metric

import (
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	versioncollector "github.com/prometheus/client_golang/prometheus/collectors/version"
)

type MetricsOpts struct {
	// enables per-flag metrics for bulk evaluation endpoints
	EnableBulkMetricFlagNames bool
}

// GOFFSubSystem is the name of the prefix we are using for all the metrics
const GOFFSubSystem = "gofeatureflag"

// NewMetrics is the constructor for the custom metrics
// nolint:funlen
func NewMetrics(opts ...MetricsOpts) (Metrics, error) {
	// default opts
	mo := &MetricsOpts{
		EnableBulkMetricFlagNames: false,
	}

	for _, opt := range opts {
		if opt.EnableBulkMetricFlagNames {
			mo.EnableBulkMetricFlagNames = true
		}
	}
	customRegistry := prom.NewRegistry()

	// counts the number of flag evaluations
	flagEvaluationCounter := prom.NewCounterVec(prom.CounterOpts{
		Name:      "flag_evaluations_total",
		Help:      "Counter events for number of flag evaluation.",
		Subsystem: GOFFSubSystem,
	}, []string{"flag_name"})

	// counts the number of call to the all flag endpoint
	allFlagCounter := prom.NewCounter(prom.CounterOpts{
		Name:      "all_flag_evaluations_total",
		Help:      "Counter events for number of all flags requests.",
		Subsystem: GOFFSubSystem,
	})

	// counts the number of bulk evaluations along with flag
	allFlagCounterWithFlag := prom.NewCounterVec(prom.CounterOpts{
		Name:      "all_flags_evaluations_total_with_flag",
		Help:      "Counter events for all flags bulk evaluations with individual flag names",
		Subsystem: GOFFSubSystem,
	}, []string{"flag_name"})

	// counts the number of tracking events collected through the API
	collectEvalDataCounter := prom.NewCounter(prom.CounterOpts{
		Name:      "collect_eval_data_total",
		Help:      "Counter events for data collector.",
		Subsystem: GOFFSubSystem,
	})

	// counts the number of flag changes (create, update, delete) from your configuration
	flagChange := prom.NewCounter(prom.CounterOpts{
		Name:      "flag_changes_total",
		Help:      "Counter that counts the number of flag changes.",
		Subsystem: GOFFSubSystem,
	})

	// counts the number of new flag created from your configuration
	flagCreateCounter := prom.NewCounter(prom.CounterOpts{
		Name:      "flag_create_total",
		Help:      "Counter that counts the number of flag created.",
		Subsystem: GOFFSubSystem,
	})

	// counts the number of flag deleted from your configuration
	flagDeleteCounter := prom.NewCounter(prom.CounterOpts{
		Name:      "flag_delete_total",
		Help:      "Counter that counts the number of flag deleted.",
		Subsystem: GOFFSubSystem,
	})

	// counts the number of flag updated from your configuration
	flagUpdateCounter := prom.NewCounter(prom.CounterOpts{
		Name:      "flag_update_total",
		Help:      "Counter that counts the number of flag updated.",
		Subsystem: GOFFSubSystem,
	})

	// counts the number of update per flag
	flagUpdateCounterVec := prom.NewCounterVec(prom.CounterOpts{
		Name:      "flag_update",
		Help:      "Counter events for number of update per flag.",
		Subsystem: GOFFSubSystem,
	}, []string{"flag_name"})

	// counts the number of delete per flag
	flagDeleteCounterVec := prom.NewCounterVec(prom.CounterOpts{
		Name:      "flag_delete",
		Help:      "Counter events for number of delete per flag.",
		Subsystem: GOFFSubSystem,
	}, []string{"flag_name"})

	// flagCreateCounterVec counts the number of create per flag
	flagCreateCounterVec := prom.NewCounterVec(prom.CounterOpts{
		Name:      "flag_create",
		Help:      "Counter events for number of create per flag.",
		Subsystem: GOFFSubSystem,
	}, []string{"flag_name"})

	// counts the number of flag updated from your configuration
	forceRefreshCounter := prom.NewCounter(prom.CounterOpts{
		Name:      "force_refresh",
		Help:      "Counter that counts the number of force refresh.",
		Subsystem: GOFFSubSystem,
	})

	// counts the number of call to the flag configuration endpoint
	flagConfigurationCounter := prom.NewCounter(prom.CounterOpts{
		Name:      "flag_configuration_total",
		Help:      "Counter events for number of configuration api requests.",
		Subsystem: GOFFSubSystem,
	})

	metricToRegister := []prom.Collector{
		flagEvaluationCounter,
		allFlagCounter,
		allFlagCounterWithFlag,
		collectEvalDataCounter,
		flagChange,
		flagCreateCounter,
		flagDeleteCounter,
		flagUpdateCounter,
		flagUpdateCounterVec,
		flagDeleteCounterVec,
		flagCreateCounterVec,
		forceRefreshCounter,
		flagConfigurationCounter,
		versioncollector.NewCollector(GOFFSubSystem),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		collectors.NewGoCollector(),
	}

	// register all the metric in the custom registry
	for _, metric := range metricToRegister {
		if err := customRegistry.Register(metric); err != nil {
			return Metrics{}, err
		}
	}

	return Metrics{
		opts:                     *mo,
		flagEvaluationCounter:    *flagEvaluationCounter,
		allFlagCounter:           allFlagCounter,
		allFlagCounterWithFlag:   *allFlagCounterWithFlag,
		collectEvalDataCounter:   collectEvalDataCounter,
		flagChange:               flagChange,
		flagCreateCounter:        flagCreateCounter,
		flagDeleteCounter:        flagDeleteCounter,
		flagUpdateCounter:        flagUpdateCounter,
		flagUpdateCounterVec:     *flagUpdateCounterVec,
		flagDeleteCounterVec:     *flagDeleteCounterVec,
		flagCreateCounterVec:     *flagCreateCounterVec,
		forceRefreshCounter:      forceRefreshCounter,
		flagConfigurationCounter: flagConfigurationCounter,
		Registry:                 customRegistry,
	}, nil
}

// Metrics is a struct containing all custom prometheus metrics
type Metrics struct {
	opts                     MetricsOpts
	Registry                 *prom.Registry
	flagEvaluationCounter    prom.CounterVec
	allFlagCounter           prom.Counter
	allFlagCounterWithFlag   prom.CounterVec
	collectEvalDataCounter   prom.Counter
	flagChange               prom.Counter
	flagCreateCounter        prom.Counter
	flagDeleteCounter        prom.Counter
	flagUpdateCounter        prom.Counter
	forceRefreshCounter      prom.Counter
	flagConfigurationCounter prom.Counter
	flagUpdateCounterVec     prom.CounterVec
	flagDeleteCounterVec     prom.CounterVec
	flagCreateCounterVec     prom.CounterVec
}

func (m *Metrics) IncFlagEvaluation(flagName string) {
	if m.flagEvaluationCounter.MetricVec != nil {
		labels := prom.Labels{"flag_name": flagName}
		m.flagEvaluationCounter.With(labels).Inc()
	}
}

// IncAllFlag increment the number call to AllFlag
func (m *Metrics) IncAllFlag(flagNames ...string) {
	if m.allFlagCounter != nil {
		m.allFlagCounter.Inc()
	}

	if m.ShouldCollectBulkMetrics() {
		for _, flagName := range flagNames {
			m.allFlagCounterWithFlag.With(prom.Labels{"flag_name": flagName}).Inc()
		}
	}
}

// IncForceRefresh increment the number call to ForceRefresh
func (m *Metrics) IncForceRefresh() {
	if m.forceRefreshCounter != nil {
		m.forceRefreshCounter.Inc()
	}
}

// IncCollectEvalData is collecting the number of events collected through the API.
func (m *Metrics) IncCollectEvalData(numberEvents float64) {
	if m.collectEvalDataCounter != nil {
		m.collectEvalDataCounter.Add(numberEvents)
	}
}

// IncFlagUpdated is incrementing the counters when a flag is updated.
func (m *Metrics) IncFlagUpdated(flagName string) {
	if m.flagUpdateCounterVec.MetricVec != nil {
		m.flagUpdateCounterVec.With(prom.Labels{"flag_name": flagName}).Inc()
		m.flagUpdateCounter.Inc()
	}
}

// IncFlagDeleted is incrementing the counters when a flag is deleted.
func (m *Metrics) IncFlagDeleted(flagName string) {
	if m.flagDeleteCounterVec.MetricVec != nil {
		m.flagDeleteCounterVec.With(prom.Labels{"flag_name": flagName}).Inc()
		m.flagDeleteCounter.Inc()
	}
}

// IncFlagCreated is incrementing the counters when a flag is created.
func (m *Metrics) IncFlagCreated(flagName string) {
	if m.flagCreateCounterVec.MetricVec != nil {
		m.flagCreateCounterVec.With(prom.Labels{"flag_name": flagName}).Inc()
		m.flagCreateCounter.Inc()
	}
}

// IncFlagChange is incrementing the counters when a flag is created, updated or deleted.
func (m *Metrics) IncFlagChange() {
	if m.flagChange != nil {
		m.flagChange.Inc()
	}
}

// IncFlagConfigurationCall is incrementing the counters when the flag configuration endpoint is called.
func (m *Metrics) IncFlagConfigurationCall() {
	if m.flagConfigurationCounter != nil {
		m.flagConfigurationCounter.Inc()
	}
}

func (m *Metrics) ShouldCollectBulkMetrics() bool {
	return m.opts.EnableBulkMetricFlagNames && m.allFlagCounterWithFlag.MetricVec != nil
}

package metric

import (
	prom "github.com/prometheus/client_golang/prometheus"
)

// GOFFSubSystem is the name of the prefix we are using for all the metrics
const GOFFSubSystem = "gofeatureflag"

// NewMetrics is the constructor for the custom metrics

func NewMetrics() (Metrics, error) {
	customRegistry := prom.NewRegistry()
	flagEvaluationCounter := prom.NewCounterVec(prom.CounterOpts{
		Name:      "flag_evaluations_total",
		Help:      "Counter events for number of flag evaluation.",
		Subsystem: GOFFSubSystem,
	}, []string{"flag_name"})

	allFlagCounter := prom.NewCounter(prom.CounterOpts{
		Name:      "all_flag_evaluations_total",
		Help:      "Counter events for number of all flags requests.",
		Subsystem: GOFFSubSystem,
	})

	collectEvalDataCounter := prom.NewCounter(prom.CounterOpts{
		Name:      "collect_eval_data_total",
		Help:      "Counter events for data collector.",
		Subsystem: GOFFSubSystem,
	})

	flagUpdateCounter := prom.NewCounter(prom.CounterOpts{
		Name:      "flag_changes_total",
		Help:      "Counter that counts the number of flag changes.",
		Subsystem: GOFFSubSystem,
	})

	metricToRegister := []prom.Collector{
		flagEvaluationCounter,
		allFlagCounter,
		collectEvalDataCounter,
		flagUpdateCounter,
	}

	// register all the metric in the custom registry
	for _, metric := range metricToRegister {
		if err := customRegistry.Register(metric); err != nil {
			return Metrics{}, err
		}
	}

	return Metrics{
		flagEvaluationCounter:  *flagEvaluationCounter,
		allFlagCounter:         allFlagCounter,
		collectEvalDataCounter: collectEvalDataCounter,
		Registry:               customRegistry,
	}, nil
}

// Metrics is a struct containing all custom prometheus metrics
type Metrics struct {
	Registry               *prom.Registry
	flagEvaluationCounter  prom.CounterVec
	allFlagCounter         prom.Counter
	collectEvalDataCounter prom.Counter
}

func (m *Metrics) IncFlagEvaluation(flagName string) {
	if m.flagEvaluationCounter.MetricVec != nil {
		labels := prom.Labels{"flag_name": flagName}
		m.flagEvaluationCounter.With(labels).Inc()
	}
}

// IncAllFlag increment the number call to AllFlag
func (m *Metrics) IncAllFlag() {
	if m.allFlagCounter != nil {
		m.allFlagCounter.Inc()
	}
}

// IncCollectEvalData is collecting the number of events collected through the API.
func (m *Metrics) IncCollectEvalData(numberEvents float64) {
	if m.collectEvalDataCounter != nil {
		m.collectEvalDataCounter.Add(numberEvents)
	}
}

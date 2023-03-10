package metric

import (
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	prom "github.com/prometheus/client_golang/prometheus"
)

// CustomMetrics is the name of the field we add in the context.
const CustomMetrics = "custom_metrics"

// NewMetrics is the constructor for the custom metrics
func NewMetrics() *Metrics {
	return &Metrics{
		flagEvaluationCounter: &prometheus.Metric{
			Name:        "flag_evaluations_total",
			Description: "Counter events for number of flag evaluation.",
			Type:        "counter_vec",
			Args:        []string{"flag_name"},
		},
		allFlagCounter: &prometheus.Metric{
			Name:        "all_flag_evaluations_total",
			Description: "Counter events for number of all flags requests.",
			Type:        "counter_vec",
			Args:        []string{},
		},
	}
}

// Metrics is a struct containing all custom prometheus metrics
type Metrics struct {
	flagEvaluationCounter *prometheus.Metric
	allFlagCounter        *prometheus.Metric
}

// MetricList return the available metrics
func (m *Metrics) MetricList() []*prometheus.Metric {
	return []*prometheus.Metric{
		m.flagEvaluationCounter,
		m.allFlagCounter,
	}
}

// AddCustomMetricsMiddleware is the function to add the middleware in echo
func (m *Metrics) AddCustomMetricsMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Set(CustomMetrics, m)
		return next(c)
	}
}

// IncFlagEvaluation increment the number of flag evaluations
func (m *Metrics) IncFlagEvaluation(flagName string) {
	labels := prom.Labels{"flag_name": flagName}
	m.flagEvaluationCounter.MetricCollector.(*prom.CounterVec).With(labels).Inc()
}

// IncAllFlag increment the number call to AllFlag
func (m *Metrics) IncAllFlag() {
	labels := prom.Labels{}
	m.allFlagCounter.MetricCollector.(*prom.CounterVec).With(labels).Inc()
}

package metric

import (
	"github.com/labstack/echo/v4"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo-contrib/prometheus"
	"github.com/stretchr/testify/assert"
)

// Test that NewMetrics creates valid prometheus.Metric objects
func TestNewMetrics(t *testing.T) {
	metrics := NewMetrics()

	// Ensure that the flag evaluation counter has the expected name and type
	assert.Equal(t, "flag_evaluations_total", metrics.flagEvaluationCounter.Name, "Unexpected flag evaluation counter name")
	assert.Equal(t, "counter_vec", metrics.flagEvaluationCounter.Type, "Unexpected flag evaluation counter type")

	// Ensure that the all flag counter has the expected name and type
	assert.Equal(t, "all_flag_evaluations_total", metrics.allFlagCounter.Name, "Unexpected all flag counter name")
	assert.Equal(t, "counter_vec", metrics.allFlagCounter.Type, "Unexpected all flag counter type")

	// Ensure that the flag evaluation counter has the expected argument label
	assert.Equal(t, []string{"flag_name"}, metrics.flagEvaluationCounter.Args, "Unexpected flag evaluation counter argument label")

	// Ensure that the all flag counter has no argument labels
	assert.Equal(t, []string{}, metrics.allFlagCounter.Args, "Unexpected all flag counter argument labels")
}

// Test that MetricList returns all of the available metrics
func TestMetricList(t *testing.T) {
	metrics := NewMetrics()

	// Ensure that MetricList returns both of the metrics created by NewMetrics
	expectedMetrics := []*prometheus.Metric{
		metrics.flagEvaluationCounter,
		metrics.allFlagCounter,
	}
	assert.ElementsMatch(t, expectedMetrics, metrics.MetricList(), "Unexpected metrics returned by MetricList")
}

// Test that AddCustomMetricsMiddleware sets the custom metrics in the echo context
func TestAddCustomMetricsMiddleware(t *testing.T) {
	metrics := NewMetrics()

	// Create a new echo context
	e := echo.New()
	req := httptest.NewRequest(echo.GET, "/", nil)
	rec := httptest.NewRecorder()
	prometheus := prometheus.NewPrometheus("gofeatureflag", nil, metrics.MetricList())
	prometheus.Use(e)

	// Define a handler function that checks that the custom metrics were added to the context
	handler := func(c echo.Context) error {
		assert.NotNil(t, c.Get(CustomMetrics), "Expected custom metrics to be set in context")
		metrics := c.Get(CustomMetrics).(*Metrics)
		metrics.IncAllFlag()
		metrics.IncFlagEvaluation("name")
		metrics.IncFlagEvaluation("name")
		return c.NoContent(200)
	}

	// Add the custom metrics middleware and call the handler function
	middleware := metrics.AddCustomMetricsMiddleware(handler)
	_ = middleware(e.NewContext(req, rec))

	expectedLabels := prom.Labels{"flag_name": "name"}
	gotFlagEvaluation := testutil.ToFloat64(metrics.flagEvaluationCounter.MetricCollector.(*prom.CounterVec).With(expectedLabels))
	gotAllFlag := testutil.ToFloat64(metrics.allFlagCounter.MetricCollector.(*prom.CounterVec))
	assert.Equal(t, float64(2), gotFlagEvaluation, "Unexpected flag evaluation count")
	assert.Equal(t, float64(1), gotAllFlag, "Unexpected all flag count")
}

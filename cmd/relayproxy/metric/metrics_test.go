package metric

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestMetrics_IncAllFlag(t *testing.T) {
	metricSrv, err := NewMetrics()
	assert.NoError(t, err)

	metricSrv.IncAllFlag()
	metricSrv.IncAllFlag()
	metricSrv.IncAllFlag()

	assert.Equal(t, 3.0, testutil.ToFloat64(metricSrv.allFlagCounter))
}

func TestMetrics_IncAllFlag_WithBulkMetricEnabled(t *testing.T) {
	metricSrv, err := NewMetrics(MetricsOpts{
		EnableBulkMetricFlagNames: true,
	})
	assert.NoError(t, err)
	metricSrv.IncAllFlag("test-flag")
	metricSrv.IncAllFlag("test-flag")
	metricSrv.IncAllFlag("test-flag", "test-flag2")

	assert.Equal(
		t,
		3.0,
		testutil.ToFloat64(metricSrv.allFlagCounterWithFlag.WithLabelValues("test-flag")),
	)
	assert.Equal(
		t,
		1.0,
		testutil.ToFloat64(metricSrv.allFlagCounterWithFlag.WithLabelValues("test-flag2")),
	)
	assert.Equal(t, 3.0, testutil.ToFloat64(metricSrv.allFlagCounter))
}

func TestMetrics_IncCollectEvalData(t *testing.T) {
	metricSrv, err := NewMetrics()
	assert.NoError(t, err)

	metricSrv.IncCollectEvalData(123)
	metricSrv.IncCollectEvalData(2)

	assert.Equal(t, 125.0, testutil.ToFloat64(metricSrv.collectEvalDataCounter))
}

func TestMetrics_IncFlagEvaluation(t *testing.T) {
	metricSrv, err := NewMetrics()
	assert.NoError(t, err)

	metricSrv.IncFlagEvaluation("test-flag")
	metricSrv.IncFlagEvaluation("test-flag")
	metricSrv.IncFlagEvaluation("test-flag2")

	assert.Equal(
		t,
		2.0,
		testutil.ToFloat64(metricSrv.flagEvaluationCounter.WithLabelValues("test-flag")),
	)
	assert.Equal(
		t,
		1.0,
		testutil.ToFloat64(metricSrv.flagEvaluationCounter.WithLabelValues("test-flag2")),
	)
}

func TestMetrics_IncFlagCreated(t *testing.T) {
	metricSrv, err := NewMetrics()
	assert.NoError(t, err)

	metricSrv.IncFlagCreated("test-flag")
	metricSrv.IncFlagCreated("test-flag2")

	assert.Equal(t, 2.0, testutil.ToFloat64(metricSrv.flagCreateCounter))
	assert.Equal(
		t,
		1.0,
		testutil.ToFloat64(metricSrv.flagCreateCounterVec.WithLabelValues("test-flag2")),
	)
	assert.Equal(
		t,
		1.0,
		testutil.ToFloat64(metricSrv.flagCreateCounterVec.WithLabelValues("test-flag")),
	)
}

func TestMetrics_IncFlagUpdated(t *testing.T) {
	metricSrv, err := NewMetrics()
	assert.NoError(t, err)

	metricSrv.IncFlagUpdated("test-flag")
	metricSrv.IncFlagUpdated("test-flag2")

	assert.Equal(t, 2.0, testutil.ToFloat64(metricSrv.flagUpdateCounter))
	assert.Equal(
		t,
		1.0,
		testutil.ToFloat64(metricSrv.flagUpdateCounterVec.WithLabelValues("test-flag2")),
	)
	assert.Equal(
		t,
		1.0,
		testutil.ToFloat64(metricSrv.flagUpdateCounterVec.WithLabelValues("test-flag")),
	)
}

func TestMetrics_IncFlagDeleted(t *testing.T) {
	metricSrv, err := NewMetrics()
	assert.NoError(t, err)

	metricSrv.IncFlagDeleted("test-flag")
	metricSrv.IncFlagDeleted("test-flag2")

	assert.Equal(t, 2.0, testutil.ToFloat64(metricSrv.flagDeleteCounter))
	assert.Equal(
		t,
		1.0,
		testutil.ToFloat64(metricSrv.flagDeleteCounterVec.WithLabelValues("test-flag2")),
	)
	assert.Equal(
		t,
		1.0,
		testutil.ToFloat64(metricSrv.flagDeleteCounterVec.WithLabelValues("test-flag")),
	)
}

func TestMetrics_IncFlagChange(t *testing.T) {
	metricSrv, err := NewMetrics()
	assert.NoError(t, err)

	metricSrv.IncFlagChange()
	metricSrv.IncFlagChange()
	metricSrv.IncFlagChange()

	assert.Equal(t, 3.0, testutil.ToFloat64(metricSrv.flagChange))
}

func TestMetrics_IncForceRefresh(t *testing.T) {
	metricSrv, err := NewMetrics()
	assert.NoError(t, err)

	metricSrv.IncForceRefresh()
	metricSrv.IncForceRefresh()
	metricSrv.IncForceRefresh()

	assert.Equal(t, 3.0, testutil.ToFloat64(metricSrv.forceRefreshCounter))
}

func TestMetrics_IncFlagConfigurationCall(t *testing.T) {
	metricSrv, err := NewMetrics()
	assert.NoError(t, err)

	metricSrv.IncFlagConfigurationCall()
	metricSrv.IncFlagConfigurationCall()
	metricSrv.IncFlagConfigurationCall()

	assert.Equal(t, 3.0, testutil.ToFloat64(metricSrv.flagConfigurationCounter))
}

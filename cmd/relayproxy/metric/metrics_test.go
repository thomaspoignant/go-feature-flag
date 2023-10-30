package metric

import (
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMetrics_IncAllFlag(t *testing.T) {
	metricSrv, err := NewMetrics()
	assert.NoError(t, err)

	metricSrv.IncAllFlag()
	metricSrv.IncAllFlag()
	metricSrv.IncAllFlag()

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

	assert.Equal(t, 2.0, testutil.ToFloat64(metricSrv.flagEvaluationCounter.WithLabelValues("test-flag")))
	assert.Equal(t, 1.0, testutil.ToFloat64(metricSrv.flagEvaluationCounter.WithLabelValues("test-flag2")))
}

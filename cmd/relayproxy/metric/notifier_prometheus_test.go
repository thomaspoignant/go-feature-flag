package metric

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/notifier"
)

func TestPrometheusNotifier_with_diff(t *testing.T) {
	m, err := NewMetrics()
	assert.NoError(t, err)

	diff := notifier.DiffCache{
		Deleted: map[string]flag.Flag{
			"test-flag": &flag.InternalFlag{},
		},
		Updated: map[string]notifier.DiffUpdated{
			"test-flag2": {
				Before: &flag.InternalFlag{},
				After:  &flag.InternalFlag{},
			},
		},
		Added: map[string]flag.Flag{
			"test-flagAdd1": &flag.InternalFlag{},
			"test-flagAdd2": &flag.InternalFlag{},
			"test-flagAdd3": &flag.InternalFlag{},
		},
	}

	n := NewPrometheusNotifier(m)
	err = n.Notify(diff)
	assert.NoError(t, err)

	assert.Equal(t, 1.0, testutil.ToFloat64(m.flagDeleteCounter))
	assert.Equal(t, 1.0, testutil.ToFloat64(m.flagUpdateCounter))
	assert.Equal(t, 3.0, testutil.ToFloat64(m.flagCreateCounter))
	assert.Equal(
		t,
		1.0,
		testutil.ToFloat64(m.flagCreateCounterVec.WithLabelValues("test-flagAdd1")),
	)
	assert.Equal(
		t,
		1.0,
		testutil.ToFloat64(m.flagCreateCounterVec.WithLabelValues("test-flagAdd2")),
	)
	assert.Equal(
		t,
		1.0,
		testutil.ToFloat64(m.flagCreateCounterVec.WithLabelValues("test-flagAdd3")),
	)
	assert.Equal(t, 1.0, testutil.ToFloat64(m.flagUpdateCounterVec.WithLabelValues("test-flag2")))
	assert.Equal(t, 1.0, testutil.ToFloat64(m.flagDeleteCounterVec.WithLabelValues("test-flag")))
	assert.Equal(t, 1.0, testutil.ToFloat64(m.flagChange))
}

func TestPrometheusNotifier_no_diff(t *testing.T) {
	m, err := NewMetrics()
	assert.NoError(t, err)

	diff := notifier.DiffCache{
		Deleted: map[string]flag.Flag{},
		Updated: map[string]notifier.DiffUpdated{},
		Added:   map[string]flag.Flag{},
	}

	n := NewPrometheusNotifier(m)
	err = n.Notify(diff)
	assert.NoError(t, err)
	assert.Equal(t, 0.0, testutil.ToFloat64(m.flagDeleteCounter))
	assert.Equal(t, 0.0, testutil.ToFloat64(m.flagUpdateCounter))
	assert.Equal(t, 0.0, testutil.ToFloat64(m.flagCreateCounter))
	assert.Equal(
		t,
		0.0,
		testutil.ToFloat64(m.flagCreateCounterVec.WithLabelValues("test-flagAdd1")),
	)
	assert.Equal(
		t,
		0.0,
		testutil.ToFloat64(m.flagCreateCounterVec.WithLabelValues("test-flagAdd2")),
	)
	assert.Equal(
		t,
		0.0,
		testutil.ToFloat64(m.flagCreateCounterVec.WithLabelValues("test-flagAdd3")),
	)
	assert.Equal(t, 0.0, testutil.ToFloat64(m.flagUpdateCounterVec.WithLabelValues("test-flag2")))
	assert.Equal(t, 0.0, testutil.ToFloat64(m.flagDeleteCounterVec.WithLabelValues("test-flag")))
	assert.Equal(t, 0.0, testutil.ToFloat64(m.flagChange))
}

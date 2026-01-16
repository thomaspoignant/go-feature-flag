package flag

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/modules/core/testutils/testconvert"
)

func TestRuleEvaluateProgressiveRollout(t *testing.T) {
	r := &Rule{
		ProgressiveRollout: &ProgressiveRollout{
			Initial: &ProgressiveRolloutStep{
				Variation:  testconvert.String("A"),
				Percentage: testconvert.Float64(0),
				Date:       testconvert.Time(time.Now().Add(-1 * time.Second)),
			},
			End: &ProgressiveRolloutStep{
				Variation:  testconvert.String("B"),
				Percentage: testconvert.Float64(25),
				Date:       testconvert.Time(time.Now().Add(1 * time.Second)),
			},
		},
	}

	// before initial date everything is A
	variation, err := r.getVariationFromProgressiveRollout(1, time.Now().Add(-2*time.Second))
	require.NoError(t, err)
	assert.Equal(t, "A", variation)

	variation, err = r.getVariationFromProgressiveRollout(99999, time.Now().Add(-2*time.Second))
	require.NoError(t, err)
	assert.Equal(t, "A", variation)

	// after initial date but before end date we have A and B
	variation, err = r.getVariationFromProgressiveRollout(5, time.Now())
	require.NoError(t, err)
	assert.Equal(t, "B", variation)

	variation, err = r.getVariationFromProgressiveRollout(12499, time.Now())
	require.NoError(t, err)
	assert.Equal(t, "B", variation)

	variation, err = r.getVariationFromProgressiveRollout(12500, time.Now())
	require.NoError(t, err)
	assert.Equal(t, "A", variation)

	// at the end date we should have the correct repartition of A and B
	variation, err = r.getVariationFromProgressiveRollout(24999, time.Now().Add(1*time.Second))
	require.NoError(t, err)
	assert.Equal(t, "B", variation)

	variation, err = r.getVariationFromProgressiveRollout(25000, time.Now().Add(1*time.Second))
	require.NoError(t, err)
	assert.Equal(t, "A", variation)

	// after the end date we should have the correct repartition of A and B
	variation, err = r.getVariationFromProgressiveRollout(24999, time.Now().Add(5*time.Second))
	require.NoError(t, err)
	assert.Equal(t, "B", variation)

	variation, err = r.getVariationFromProgressiveRollout(25000, time.Now().Add(5*time.Second))
	require.NoError(t, err)
	assert.Equal(t, "A", variation)
}

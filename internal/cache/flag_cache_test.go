package cache

import (
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/internal/model"
	"github.com/thomaspoignant/go-feature-flag/testutils/testconvert"
	"testing"
)

func TestFlagsCache_Copy(t *testing.T) {
	tests := []struct {
		name string
		fc   FlagsCache
	}{
		{
			name: "Copy with values",
			fc: FlagsCache{
				"test": model.FlagData{
					Disable:    testconvert.Bool(false),
					Rule:       testconvert.String("key eq \"toto\""),
					Percentage: testconvert.Float64(20),
					True:       testconvert.Interface(true),
					False:      testconvert.Interface(false),
					Default:    testconvert.Interface(true),
				},
			},
		},
		{
			name: "Copy without value",
			fc:   FlagsCache{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fc.Copy()
			assert.Equal(t, tt.fc, got)
		})
	}
}

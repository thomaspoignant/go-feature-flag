package cache

import (
	"github.com/stretchr/testify/assert"
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
				"test": {
					Disable:    false,
					Rule:       "key eq \"toto\"",
					Percentage: 20,
					True:       true,
					False:      false,
					Default:    true,
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

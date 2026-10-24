package flag_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/modules/core/ffcontext"
)

func TestInternalFlag_Value_ExpiringFlagsAreNotCacheable(t *testing.T) {
	now := time.Now()
	future := now.Add(time.Hour)

	tests := []struct {
		name          string
		createdAt     *time.Time
		expiresAt     *time.Time
		wantCacheable bool
	}{
		{name: "never expires", wantCacheable: true},
		{name: "expiresAt", expiresAt: &future},
		{name: "createdAt with a ttl", createdAt: &now},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := boolFlag("enabled")
			f.CreatedAt, f.ExpiresAt = tt.createdAt, tt.expiresAt
			_, details := f.Value("flag", ffcontext.NewEvaluationContext("user"), resolverOf(nil))
			assert.Equal(t, tt.wantCacheable, details.Cacheable)
		})
	}
}

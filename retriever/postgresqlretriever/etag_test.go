package postgresqlretriever

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComputeETag(t *testing.T) {
	tests := []struct {
		name        string
		definition  string
		definition2 string
		sameETag    bool
		wantErr     assert.ErrorAssertionFunc
	}{
		{
			name:        "identical definitions produce the same etag",
			definition:  `{"variations":{"enabled":true,"disabled":false},"defaultRule":{"variation":"disabled"}}`,
			definition2: `{"variations":{"enabled":true,"disabled":false},"defaultRule":{"variation":"disabled"}}`,
			sameETag:    true,
			wantErr:     assert.NoError,
		},
		{
			name: "same content with different key order produces the same etag",
			definition: `{"variations":{"enabled":true,"disabled":false},
				"defaultRule":{"variation":"disabled"}}`,
			definition2: `{"defaultRule":{"variation":"disabled"},
				"variations":{"disabled":false,"enabled":true}}`,
			sameETag: true,
			wantErr:  assert.NoError,
		},
		{
			name:        "different definitions produce different etags",
			definition:  `{"variations":{"enabled":true,"disabled":false},"defaultRule":{"variation":"disabled"}}`,
			definition2: `{"variations":{"enabled":true,"disabled":false},"defaultRule":{"variation":"enabled"}}`,
			sameETag:    false,
			wantErr:     assert.NoError,
		},
		{
			name:       "invalid json returns an error",
			definition: `not-json`,
			wantErr:    assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := computeETag([]byte(tt.definition))
			tt.wantErr(t, err)
			if err != nil {
				return
			}
			assert.NotEmpty(t, got)

			if tt.definition2 != "" {
				got2, err2 := computeETag([]byte(tt.definition2))
				assert.NoError(t, err2)
				if tt.sameETag {
					assert.Equal(t, got, got2)
				} else {
					assert.NotEqual(t, got, got2)
				}
			}
		})
	}
}

package evaluation

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
)

func Test_constructMetadataParallel(t *testing.T) {
	sharedFlag := flag.InternalFlag{
		Metadata: &map[string]interface{}{
			"key1": "value1",
		},
	}

	type args struct {
		resolutionDetails flag.ResolutionDetails
	}
	var tests []struct {
		name                  string
		args                  args
		wantEvaluatedRuleName string
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

	// generate test cases
	for i := 0; i < 10_000; i++ {
		ruleName := fmt.Sprintf("rule-%d", i)
		tests = append(tests, struct {
			name                  string
			args                  args
			wantEvaluatedRuleName string
		}{
			name: fmt.Sprintf("Rule %d", i),
			args: args{
				resolutionDetails: flag.ResolutionDetails{
					RuleName: &ruleName,
				},
			},
			wantEvaluatedRuleName: ruleName,
		})
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := constructMetadata(&sharedFlag, tt.args.resolutionDetails)
			assert.Equal(t, tt.wantEvaluatedRuleName, got["evaluatedRuleName"])
		})
	}
}

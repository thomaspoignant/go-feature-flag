package evaluate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/model"
)

func Test_evaluate_Evaluate(t *testing.T) {
	tests := []struct {
		name           string
		evaluate       evaluate
		wantErr        assert.ErrorAssertionFunc
		expectedErr    string
		expectedResult map[string]model.RawVarResult
	}{
		{
			name: "Should error is config file does not exist",
			evaluate: evaluate{
				config:        "testdata/invalid.yaml",
				fileFormat:    "yaml",
				flag:          "test-flag",
				evaluationCtx: `{"targetingKey": "user-123"}`,
			},
			wantErr:     assert.Error,
			expectedErr: "impossible to retrieve the flags, please check your configuration: open testdata/invalid.yaml: no such file or directory",
		},
		{
			name: "Should error if no evaluation context provided",
			evaluate: evaluate{
				config:     "testdata/flag.goff.yaml",
				fileFormat: "yaml",
				flag:       "test-flag",
			},
			wantErr:     assert.Error,
			expectedErr: "invalid evaluation context (missing targeting key)",
		},
		{
			name: "Should error if evaluation context provided has no targeting key",
			evaluate: evaluate{
				config:        "testdata/flag.goff.yaml",
				fileFormat:    "yaml",
				flag:          "test-flag",
				evaluationCtx: `{"id": "user-123"}`,
			},
			wantErr:     assert.Error,
			expectedErr: "invalid evaluation context (missing targeting key)",
		},
		{
			name: "Should evaluate a single flag if flag name is provided",
			evaluate: evaluate{
				config:        "testdata/flag.goff.yaml",
				fileFormat:    "yaml",
				flag:          "test-flag",
				evaluationCtx: `{"targetingKey": "user-123"}`,
			},
			wantErr: assert.NoError,
			expectedResult: map[string]model.RawVarResult{
				"test-flag": {
					TrackEvents:   true,
					VariationType: "Default",
					Failed:        false,
					Version:       "",
					Reason:        "DEFAULT",
					ErrorCode:     "",
					ErrorDetails:  "",
					Value:         false,
					Cacheable:     true,
					Metadata: map[string]any{
						"description": "this is a simple feature flag",
						"issue-link":  "https://jira.xxx/GOFF-01",
					},
				},
			},
		},
		{
			name: "Should evaluate all flags if flag name is not provided",
			evaluate: evaluate{
				config:        "testdata/flag.goff.yaml",
				fileFormat:    "yaml",
				evaluationCtx: `{"targetingKey": "user-123"}`,
			},
			wantErr: assert.NoError,
			expectedResult: map[string]model.RawVarResult{
				"test-flag": {
					TrackEvents:   true,
					VariationType: "Default",
					Failed:        false,
					Version:       "",
					Reason:        "DEFAULT",
					ErrorCode:     "",
					ErrorDetails:  "",
					Value:         false,
					Cacheable:     true,
					Metadata: map[string]any{
						"description": "this is a simple feature flag",
						"issue-link":  "https://jira.xxx/GOFF-01",
					},
				},
				"test-flag2": {
					TrackEvents:   true,
					VariationType: "Default",
					Failed:        false,
					Version:       "",
					Reason:        "DEFAULT",
					ErrorCode:     "",
					ErrorDetails:  "",
					Value:         false,
					Cacheable:     true,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := tt.evaluate.Evaluate()
			tt.wantErr(t, err)

			if tt.expectedErr != "" {
				assert.Equal(t, tt.expectedErr, err.Error())
			}

			if tt.expectedResult != nil {
				assert.Equal(t, tt.expectedResult, m)
			}
		})
	}
}

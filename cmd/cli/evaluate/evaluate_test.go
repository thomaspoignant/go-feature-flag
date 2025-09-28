package evaluate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/retrieverconf"
	retrieverInit "github.com/thomaspoignant/go-feature-flag/cmdhelpers/retrieverconf/init"
	"github.com/thomaspoignant/go-feature-flag/model"
	"github.com/thomaspoignant/go-feature-flag/retriever/bitbucketretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/githubretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/gitlabretriever"
	"github.com/thomaspoignant/go-feature-flag/testutils/mock"
)

func Test_evaluate_Evaluate(t *testing.T) {
	tests := []struct {
		name           string
		evaluate       evaluate
		initEvaluate   func() (evaluate, error)
		wantErr        assert.ErrorAssertionFunc
		expectedErr    string
		expectedResult map[string]model.RawVarResult
	}{
		{
			name: "Should error is config file does not exist",
			initEvaluate: func() (evaluate, error) {
				r, err := retrieverInit.InitRetriever(&retrieverconf.RetrieverConf{
					Kind: "file",
					Path: "testdata/invalid.yaml",
				})
				if err != nil {
					return evaluate{}, err
				}
				return evaluate{
					retriever:     r,
					fileFormat:    "yaml",
					flag:          "test-flag",
					evaluationCtx: `{"targetingKey": "user-123"}`,
				}, nil
			},
			wantErr:     assert.Error,
			expectedErr: "impossible to initialize the retrievers, please check your configuration: impossible to retrieve the flags, please check your configuration: open testdata/invalid.yaml: no such file or directory",
		},
		{
			name: "Should error if no evaluation context provided",
			initEvaluate: func() (evaluate, error) {
				r, err := retrieverInit.InitRetriever(&retrieverconf.RetrieverConf{Kind: "file", Path: "testdata/flag.goff.yaml"})
				if err != nil {
					return evaluate{}, err
				}
				return evaluate{
					retriever:  r,
					fileFormat: "yaml",
					flag:       "test-flag",
				}, nil
			},
			wantErr:     assert.Error,
			expectedErr: "invalid evaluation context (missing targeting key)",
		},
		{
			name: "Should error if evaluation context provided has no targeting key",
			initEvaluate: func() (evaluate, error) {
				r, err := retrieverInit.InitRetriever(&retrieverconf.RetrieverConf{Kind: "file", Path: "testdata/flag.goff.yaml"})
				if err != nil {
					return evaluate{}, err
				}
				return evaluate{
					retriever:     r,
					fileFormat:    "yaml",
					flag:          "test-flag",
					evaluationCtx: `{"id": "user-123"}`,
				}, nil
			},
			wantErr:     assert.Error,
			expectedErr: "invalid evaluation context (missing targeting key)",
		},
		{
			name: "Should evaluate a single flag if flag name is provided",
			initEvaluate: func() (evaluate, error) {
				r, err := retrieverInit.InitRetriever(&retrieverconf.RetrieverConf{Kind: "file", Path: "testdata/flag.goff.yaml"})
				if err != nil {
					return evaluate{}, err
				}
				return evaluate{
					retriever:     r,
					fileFormat:    "yaml",
					flag:          "test-flag",
					evaluationCtx: `{"targetingKey": "user-123"}`,
				}, nil
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
					Metadata: map[string]interface{}{
						"description": "this is a simple feature flag",
						"issue-link":  "https://jira.xxx/GOFF-01",
					},
				},
			},
		},
		{
			name: "Should evaluate all flags if flag name is not provided",
			initEvaluate: func() (evaluate, error) {
				r, err := retrieverInit.InitRetriever(&retrieverconf.RetrieverConf{Kind: "file", Path: "testdata/flag.goff.yaml"})
				if err != nil {
					return evaluate{}, err
				}
				return evaluate{
					retriever:     r,
					fileFormat:    "yaml",
					evaluationCtx: `{"targetingKey": "user-123"}`,
				}, nil
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
					Metadata: map[string]interface{}{
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
		{
			name: "Should evaluate a flag from a github repository",
			initEvaluate: func() (evaluate, error) {
				r, err := retrieverInit.InitRetriever(
					&retrieverconf.RetrieverConf{
						Kind:           "github",
						RepositorySlug: "thomaspoignant/go-feature-flag",
						GithubToken:    "XXX_GH_TOKEN",
						Path:           "testdata/flag-config.yaml"})
				if err != nil {
					return evaluate{}, err
				}

				gitHubRetriever, _ := r.(*githubretriever.Retriever)
				gitHubRetriever.SetHTTPClient(&mock.HTTP{})

				return evaluate{
					retriever:     r,
					fileFormat:    "yaml",
					flag:          "test-flag",
					evaluationCtx: `{"targetingKey": "user-123"}`,
				}, nil
			},
			wantErr: assert.NoError,
			expectedResult: map[string]model.RawVarResult{
				"test-flag": {
					TrackEvents:   true,
					VariationType: "false_var",
					Failed:        false,
					Version:       "",
					Reason:        "DEFAULT",
					ErrorCode:     "",
					ErrorDetails:  "",
					Value:         false,
					Cacheable:     true,
					Metadata:      nil,
				},
			},
		},
		{
			name: "Should evaluate a flag from a gitlab repository",
			initEvaluate: func() (evaluate, error) {
				r, err := retrieverInit.InitRetriever(
					&retrieverconf.RetrieverConf{
						Kind:           "gitlab",
						BaseURL:        baseURL,
						RepositorySlug: "thomaspoignant/go-feature-flag",
						AuthToken:      "XXX",
						Path:           "testdata/flag-config.yaml"})
				if err != nil {
					return evaluate{}, err
				}

				gitLabRetriever, _ := r.(*gitlabretriever.Retriever)
				gitLabRetriever.SetHTTPClient(&mock.HTTP{})

				return evaluate{
					retriever:     r,
					fileFormat:    "yaml",
					flag:          "test-flag",
					evaluationCtx: `{"targetingKey": "user-123"}`,
				}, nil
			},
			wantErr: assert.NoError,
			expectedResult: map[string]model.RawVarResult{
				"test-flag": {
					TrackEvents:   true,
					VariationType: "false_var",
					Failed:        false,
					Version:       "",
					Reason:        "DEFAULT",
					ErrorCode:     "",
					ErrorDetails:  "",
					Value:         false,
					Cacheable:     true,
					Metadata:      nil,
				},
			},
		},
		{
			name: "Should evaluate a flag from a bitbucket repository",
			initEvaluate: func() (evaluate, error) {
				r, err := retrieverInit.InitRetriever(
					&retrieverconf.RetrieverConf{
						Kind:           "bitbucket",
						BaseURL:        baseURL,
						RepositorySlug: "thomaspoignant/go-feature-flag",
						AuthToken:      "XXX",
						Path:           "testdata/flag-config.yaml"})
				if err != nil {
					return evaluate{}, err
				}

				bitBucketRetriever, _ := r.(*bitbucketretriever.Retriever)
				bitBucketRetriever.SetHTTPClient(&mock.HTTP{})

				return evaluate{
					retriever:     r,
					fileFormat:    "yaml",
					flag:          "test-flag",
					evaluationCtx: `{"targetingKey": "user-123"}`,
				}, nil
			},
			wantErr: assert.NoError,
			expectedResult: map[string]model.RawVarResult{
				"test-flag": {
					TrackEvents:   true,
					VariationType: "false_var",
					Failed:        false,
					Version:       "",
					Reason:        "DEFAULT",
					ErrorCode:     "",
					ErrorDetails:  "",
					Value:         false,
					Cacheable:     true,
					Metadata:      nil,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, err := tt.initEvaluate()
			if err != nil {
				tt.wantErr(t, err)
				return
			}
			m, err := e.Evaluate()
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

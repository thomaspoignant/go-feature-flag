package evaluate

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/retrieverconf"
	retrieverInit "github.com/thomaspoignant/go-feature-flag/cmdhelpers/retrieverconf/init"
	"github.com/thomaspoignant/go-feature-flag/modules/core/model"
	"github.com/thomaspoignant/go-feature-flag/retriever/bitbucketretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/gcstorageretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/githubretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/gitlabretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/httpretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/s3retrieverv2"
	"github.com/thomaspoignant/go-feature-flag/testutils"
	"github.com/thomaspoignant/go-feature-flag/testutils/mock"
	"google.golang.org/api/option"
)

func Test_Evaluate(t *testing.T) {
	tests := []struct {
		name           string
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
			name: "Should error if no evaluation context provided with flag containing percentage rules",
			initEvaluate: func() (evaluate, error) {
				r, err := retrieverInit.InitRetriever(&retrieverconf.RetrieverConf{Kind: "file", Path: "testdata/flag.goff.yaml"})
				if err != nil {
					return evaluate{}, err
				}
				return evaluate{
					retriever:  r,
					fileFormat: "yaml",
					flag:       "simple-flag",
				}, nil
			},
			wantErr: assert.NoError,
			expectedResult: map[string]model.RawVarResult{
				"simple-flag": {
					TrackEvents:   true,
					VariationType: "B",
					Failed:        false,
					Version:       "",
					Reason:        "STATIC",
					Value:         true,
					Cacheable:     true,
					Metadata: map[string]any{
						"description": "this is a simple feature flag",
						"issue-link":  "https://jira.xxx/GOFF-01",
					},
				},
			},
		},
		{
			name: "Should perform evaluation if no evaluation context and compatible flag",
			initEvaluate: func() (evaluate, error) {
				r, err := retrieverInit.InitRetriever(&retrieverconf.RetrieverConf{Kind: "file", Path: "testdata/flag.goff.yaml"})
				if err != nil {
					return evaluate{}, err
				}
				return evaluate{
					retriever:  r,
					fileFormat: "yaml",
					flag:       "simple-flag",
				}, nil
			},
			wantErr: assert.NoError,
			expectedResult: map[string]model.RawVarResult{
				"simple-flag": {
					TrackEvents:   true,
					VariationType: "B",
					Failed:        false,
					Version:       "",
					Reason:        "STATIC",
					Value:         true,
					Cacheable:     true,
					Metadata: map[string]any{
						"description": "this is a simple feature flag",
						"issue-link":  "https://jira.xxx/GOFF-01",
					},
				},
			},
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
			wantErr: assert.NoError,
			expectedResult: map[string]model.RawVarResult{
				"test-flag": {
					TrackEvents:   true,
					VariationType: "SdkDefault",
					Failed:        true,
					Version:       "",
					Reason:        "ERROR",
					ErrorCode:     "TARGETING_KEY_MISSING",
					ErrorDetails:  "Error: Empty targeting key",
					Value:         nil,
					Cacheable:     false,
					Metadata: map[string]any{
						"description": "this is a simple feature flag",
						"issue-link":  "https://jira.xxx/GOFF-01",
					},
				},
			},
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
					Metadata: map[string]any{
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
				"simple-flag": {
					TrackEvents:   true,
					VariationType: "B",
					Failed:        false,
					Version:       "",
					Reason:        "STATIC",
					ErrorCode:     "",
					ErrorDetails:  "",
					Value:         true,
					Cacheable:     true,
					Metadata: map[string]any{
						"description": "this is a simple feature flag",
						"issue-link":  "https://jira.xxx/GOFF-01",
					},
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

				gitHubRetriever, ok := r.(*githubretriever.Retriever)
				assert.True(t, ok, "failed to assert retriever to *githubretriever.Retriever")
				gitHubRetriever.SetHTTPClient(&mock.HTTP{})

				return evaluate{
					retriever:     gitHubRetriever,
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
						BaseURL:        "https://gitlab.com/api/v4/",
						RepositorySlug: "thomaspoignant/go-feature-flag",
						AuthToken:      "XXX",
						Path:           "testdata/flag-config.yaml"})
				if err != nil {
					return evaluate{}, err
				}

				gitLabRetriever, ok := r.(*gitlabretriever.Retriever)
				assert.True(t, ok, "failed to assert retriever to *gitlabretriever.Retriever")
				gitLabRetriever.SetHTTPClient(&mock.HTTP{})

				return evaluate{
					retriever:     gitLabRetriever,
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
						BaseURL:        "https://bitbucket.com/api/v4/",
						RepositorySlug: "thomaspoignant/go-feature-flag",
						AuthToken:      "XXX",
						Path:           "testdata/flag-config.yaml"})
				if err != nil {
					return evaluate{}, err
				}

				bitBucketRetriever, ok := r.(*bitbucketretriever.Retriever)
				assert.True(t, ok, "failed to assert retriever to *bitbucketretriever.Retriever")
				bitBucketRetriever.SetHTTPClient(&mock.HTTP{})

				return evaluate{
					retriever:     bitBucketRetriever,
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
			name: "Should evaluate a flag from a S3 repository",
			initEvaluate: func() (evaluate, error) {
				downloader := &testutils.S3ManagerV2Mock{
					TestDataLocation: "./testdata",
				}

				r, err := retrieverInit.InitRetriever(
					&retrieverconf.RetrieverConf{
						Kind:   "s3",
						Bucket: "Bucket",
						Item:   "valid",
					})

				if err != nil {
					return evaluate{}, err
				}

				s3Retriever, ok := r.(*s3retrieverv2.Retriever)
				assert.True(t, ok, "failed to assert retriever to *s3retrieverv2.Retriever")
				s3Retriever.SetDownloader(downloader)

				_ = s3Retriever.Init(context.Background(), nil)

				return evaluate{
					retriever:     s3Retriever,
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
					Metadata: map[string]any{"description": "this is a simple feature flag",
						"issue-link": "https://jira.xxx/GOFF-01"},
				},
			},
		},
		{
			name: "Should evaluate a flag from a HTTP endpoint",
			initEvaluate: func() (evaluate, error) {
				r, err := retrieverInit.InitRetriever(
					&retrieverconf.RetrieverConf{
						Kind:        "http",
						URL:         "http://localhost.example/file",
						HTTPMethod:  http.MethodGet,
						HTTPBody:    "",
						HTTPHeaders: nil,
					})

				if err != nil {
					return evaluate{}, err
				}

				httpRetriever, ok := r.(*httpretriever.Retriever)
				assert.True(t, ok, "failed to assert retriever to *httpretriever.Retriever")
				httpRetriever.SetHTTPClient(&mock.HTTP{})

				return evaluate{
					retriever:     httpRetriever,
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
			name: "Should evaluate a flag from a GCS",
			initEvaluate: func() (evaluate, error) {
				mockedStorage := testutils.NewMockedGCS(t)
				mockedStorage.WithFiles(t, "flags", map[string]string{"testdata/flag-config.yaml": "flag-config.yaml"})

				r, err := retrieverInit.InitRetriever(
					&retrieverconf.RetrieverConf{
						Kind:   "googleStorage",
						Bucket: "flags",
						Object: "flag-config.yaml",
					})

				if err != nil {
					return evaluate{}, err
				}

				gcsRetriever, ok := r.(*gcstorageretriever.Retriever)
				assert.True(t, ok, "failed to assert retriever to *gcstorageretriever.Retriever")
				gcsRetriever.SetOptions([]option.ClientOption{
					option.WithoutAuthentication(),
					option.WithHTTPClient(mockedStorage.Server.HTTPClient()),
				})

				return evaluate{
					retriever:     gcsRetriever,
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
					Metadata: map[string]any{"description": "this is a simple feature flag",
						"issue-link": "https://jira.xxx/GOFF-01"},
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

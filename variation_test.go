package ffclient

import (
	"errors"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thejerf/slogassert"
	"github.com/thomaspoignant/go-feature-flag/exporter"
	"github.com/thomaspoignant/go-feature-flag/exporter/logsexporter"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/internal/cache"
	"github.com/thomaspoignant/go-feature-flag/model/dto"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/modules/core/model"
	"github.com/thomaspoignant/go-feature-flag/modules/core/testutils/testconvert"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
	"github.com/thomaspoignant/go-feature-flag/testutils"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

type cacheMock struct {
	flag flag.Flag
	err  error
}

func NewCacheMock(flag flag.Flag, err error) cache.Manager {
	return &cacheMock{
		flag: flag,
		err:  err,
	}
}

func (c *cacheMock) GetLatestUpdateDate() time.Time {
	return time.Now()
}

func (c *cacheMock) ConvertToFlagStruct(_ []byte, _ string) (map[string]dto.DTO, error) {
	return nil, nil
}
func (c *cacheMock) UpdateCache(_ map[string]dto.DTO, _ *fflog.FFLogger, _ bool) error {
	return nil
}

func (c *cacheMock) Close() {}
func (c *cacheMock) GetFlag(_ string) (flag.Flag, error) {
	return c.flag, c.err
}
func (c *cacheMock) AllFlags() (map[string]flag.Flag, error) { return nil, nil }

func TestBoolVariation(t *testing.T) {
	type args struct {
		flagKey       string
		evaluationCtx ffcontext.Context
		defaultValue  bool
		cacheMock     cache.Manager
		offline       bool
		disableInit   bool
	}
	tests := []struct {
		name        string
		args        args
		want        bool
		wantErr     bool
		expectedLog string
	}{
		{
			name: "Call variation before init of SDK",
			args: args{
				flagKey:       "key-not-exist",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  false,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Variations: &map[string]*any{
						"var_true":  testconvert.Interface(true),
						"var_false": testconvert.Interface(false),
					},
					Rules: &[]flag.Rule{{
						Query:       testconvert.String("key eq \"key\""),
						Percentages: &map[string]float64{"var_true": 100, "var_false": 0}},
					},
					BucketingKey: nil,
					DefaultRule: &flag.Rule{
						VariationResult: testconvert.String("var_false"),
					},
				}, nil),
				disableInit: true,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "Get default value if flag disable",
			args: args{
				flagKey:       "disable-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  true,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Disable: testconvert.Bool(true),
				}, nil),
			},
			want:        true,
			wantErr:     false,
			expectedLog: `user="random-key", flag="disable-flag", value="true", variation="SdkDefault"`,
		},
		{
			name: "Get error when cache not init",
			args: args{
				flagKey:       "key-not-exist",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  true,
				cacheMock: NewCacheMock(
					&flag.InternalFlag{},
					errors.New("impossible to read the toggle before the initialisation")),
			},
			want:        true,
			wantErr:     true,
			expectedLog: `user="random-key", flag="key-not-exist", value="true", variation="SdkDefault"`,
		},
		{
			name: "Get default value with key not exist",
			args: args{
				flagKey:       "key-not-exist",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  true,
				cacheMock: NewCacheMock(
					&flag.InternalFlag{},
					errors.New("flag [key-not-exist] does not exists"),
				),
			},
			want:        true,
			wantErr:     true,
			expectedLog: `user="random-key", flag="key-not-exist", value="true", variation="SdkDefault"`,
		},
		{
			name: "Get default value, rule not apply",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  true,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("key eq \"key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface(true),
						"False":   testconvert.Interface(false),
						"True":    testconvert.Interface(true),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want:        true,
			wantErr:     false,
			expectedLog: `user="random-key", flag="test-flag", value="true", variation="Default"`,
		},
		{
			name: "Get true value, rule apply",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewAnonymousEvaluationContext("random-key"),
				defaultValue:  true,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("key eq \"random-key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface(false),
						"False":   testconvert.Interface(false),
						"True":    testconvert.Interface(true),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want:        true,
			wantErr:     false,
			expectedLog: `user="random-key", flag="test-flag", value="true", variation="True"`,
		},
		{
			name: "Get false value, rule apply",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewAnonymousEvaluationContext("random-key-ssss1"),
				defaultValue:  true,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("anonymous eq true"),
							Percentages: &map[string]float64{
								"False": 90,
								"True":  10,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface(false),
						"False":   testconvert.Interface(false),
						"True":    testconvert.Interface(true),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want:        false,
			wantErr:     false,
			expectedLog: `user="random-key-ssss1", flag="test-flag", value="false", variation="False"`,
		},
		{
			name: "Get default value, when rule apply and not right type",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key-ssss1"),
				defaultValue:  true,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("key eq \"random-key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface("xxx"),
						"False":   testconvert.Interface("xxx"),
						"True":    testconvert.Interface("xxx"),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want:        true,
			wantErr:     true,
			expectedLog: `user="random-key-ssss1", flag="test-flag", value="true", variation="SdkDefault"`,
		},
		{
			name: "No exported log",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewAnonymousEvaluationContext("random-key"),
				defaultValue:  true,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("key eq \"random-key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface(false),
						"False":   testconvert.Interface(false),
						"True":    testconvert.Interface(true),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
					TrackEvents: testconvert.Bool(false),
				}, nil),
			},
			want:        true,
			wantErr:     false,
			expectedLog: "",
		},
		{
			name: "Get sdk default value if offline",
			args: args{
				offline:       true,
				flagKey:       "disable-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  false,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Disable: testconvert.Bool(true),
				}, nil),
			},
			want:        false,
			wantErr:     false,
			expectedLog: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := slogassert.New(t, slog.LevelInfo, nil)
			logger := slog.New(handler)
			fflog := &fflog.FFLogger{
				LeveledLogger: slog.New(handler),
			}

			if !tt.args.disableInit {
				c := retriever.ManagerConfig{
					FileFormat:              "YAML",
					DisableNotifierOnInit:   false,
					StartWithRetrieverError: false,
					EnablePollingJitter:     true,
					PollingInterval:         500,
				}
				ff = &GoFeatureFlag{
					retrieverManager: retriever.NewManager(c, []retriever.Retriever{}, tt.args.cacheMock, fflog),
					config: Config{
						PollingInterval: 0,
						LeveledLogger:   logger,
						Offline:         tt.args.offline,
					},
					featureEventDataExporter: exporter.NewManager[exporter.FeatureEvent](
						[]exporter.Config{exporter.Config{
							FlushInterval:    0,
							MaxEventInMemory: 0,
							Exporter: &logsexporter.Exporter{
								LogFormat: "user=\"{{ .UserKey}}\", flag=\"{{ .Key}}\", " +
									"value=\"{{ .Value}}\", variation=\"{{ .Variation}}\"",
							},
						}},
						exporter.DefaultExporterCleanQueueInterval,
						fflog,
					),
				}
			}

			got, err := BoolVariation(tt.args.flagKey, tt.args.evaluationCtx, tt.args.defaultValue)

			if tt.expectedLog != "" {
				time.Sleep(
					40 * time.Millisecond,
				) // since the log is async, we are waiting to be sure it's written
				if tt.expectedLog == "" {
					handler.AssertEmpty()
				} else {
					handler.Assert(func(message slogassert.LogMessage) bool {
						if !strings.Contains(message.Message, tt.expectedLog) {
							handler.Fail("impossible to find %s in %s", tt.expectedLog, message.Message)
							return false
						}
						return true
					})
				}
			}

			if tt.wantErr {
				assert.Error(t, err, "BoolVariation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got, "BoolVariation() got = %v, want %v", got, tt.want)

			// clean logger
			ff = nil
		})
	}
}

func TestBoolVariationDetails(t *testing.T) {
	type args struct {
		flagKey       string
		evaluationCtx ffcontext.Context
		defaultValue  bool
		cacheMock     cache.Manager
		offline       bool
		disableInit   bool
	}
	tests := []struct {
		name        string
		args        args
		want        model.VariationResult[bool]
		wantErr     bool
		expectedLog string
	}{
		{
			name: "Call variation before init of SDK",
			args: args{
				flagKey:       "key-not-exist",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  false,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Variations: &map[string]*any{
						"var_true":  testconvert.Interface(true),
						"var_false": testconvert.Interface(false),
					},
					Rules: &[]flag.Rule{{
						Query:       testconvert.String("key eq \"key\""),
						Percentages: &map[string]float64{"var_true": 100, "var_false": 0}},
					},
					BucketingKey: nil,
					DefaultRule: &flag.Rule{
						VariationResult: testconvert.String("var_false"),
					},
				}, nil),
				disableInit: true,
			},
			wantErr: true,
		},
		{
			name: "Get default value if flag disable",
			args: args{
				flagKey:       "disable-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  true,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Disable: testconvert.Bool(true),
				}, nil),
			},
			want: model.VariationResult[bool]{
				VariationType: flag.VariationSDKDefault,
				Failed:        false,
				Reason:        flag.ReasonDisabled,
				Value:         true,
				TrackEvents:   true,
				Cacheable:     true,
			},
			wantErr:     false,
			expectedLog: `user="random-key", flag="disable-flag", value="true", variation="SdkDefault"`,
		},
		{
			name: "Get error when cache not init",
			args: args{
				flagKey:       "key-not-exist",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  true,
				cacheMock: NewCacheMock(
					&flag.InternalFlag{},
					errors.New("impossible to read the toggle before the initialisation")),
			},
			wantErr:     true,
			expectedLog: `user="random-key", flag="key-not-exist", value="true", variation="SdkDefault"`,
		},
		{
			name: "Get default value with key not exist",
			args: args{
				flagKey:       "key-not-exist",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  true,
				cacheMock: NewCacheMock(
					&flag.InternalFlag{},
					errors.New("flag [key-not-exist] does not exists"),
				),
			},
			wantErr:     true,
			expectedLog: `user="random-key", flag="key-not-exist", value="true", variation="SdkDefault"`,
		},
		{
			name: "Get default value, rule not apply",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  true,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("key eq \"key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface(true),
						"False":   testconvert.Interface(false),
						"True":    testconvert.Interface(true),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want: model.VariationResult[bool]{
				VariationType: "Default",
				Failed:        false,
				Reason:        flag.ReasonDefault,
				Value:         true,
				TrackEvents:   true,
				Cacheable:     true,
			},
			wantErr:     false,
			expectedLog: `user="random-key", flag="test-flag", value="true", variation="Default"`,
		},
		{
			name: "Get true value, rule apply",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewAnonymousEvaluationContext("random-key"),
				defaultValue:  true,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("key eq \"random-key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface(false),
						"False":   testconvert.Interface(false),
						"True":    testconvert.Interface(true),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want: model.VariationResult[bool]{
				VariationType: "True",
				Failed:        false,
				Reason:        flag.ReasonTargetingMatch,
				Value:         true,
				TrackEvents:   true,
				Cacheable:     true,
				Metadata: map[string]any{
					"evaluatedRuleName": "rule1",
				},
			},
			wantErr:     false,
			expectedLog: `user="random-key", flag="test-flag", value="true", variation="True"`,
		},
		{
			name: "Get rule name on metadata, rule apply",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewAnonymousEvaluationContext("random-key"),
				defaultValue:  true,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("key eq \"random-key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface(false),
						"False":   testconvert.Interface(false),
						"True":    testconvert.Interface(true),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want: model.VariationResult[bool]{
				VariationType: "True",
				Failed:        false,
				Reason:        flag.ReasonTargetingMatch,
				Value:         true,
				TrackEvents:   true,
				Cacheable:     true,
				Metadata: map[string]any{
					"evaluatedRuleName": "rule1",
				},
			},
			wantErr:     false,
			expectedLog: `user="random-key", flag="test-flag", value="true", variation="True"`,
		},
		{
			name: "Get no rule name on metadata, rule apply has not name",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewAnonymousEvaluationContext("random-key"),
				defaultValue:  true,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Query: testconvert.String("key eq \"random-key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface(false),
						"False":   testconvert.Interface(false),
						"True":    testconvert.Interface(true),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want: model.VariationResult[bool]{
				VariationType: "True",
				Failed:        false,
				Reason:        flag.ReasonTargetingMatch,
				Value:         true,
				TrackEvents:   true,
				Cacheable:     true,
			},
			wantErr:     false,
			expectedLog: `user="random-key", flag="test-flag", value="true", variation="True"`,
		},
		{
			name: "Get false value, rule apply",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewAnonymousEvaluationContext("random-key-ssss1"),
				defaultValue:  true,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("anonymous eq true"),
							Percentages: &map[string]float64{
								"False": 90,
								"True":  10,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface(false),
						"False":   testconvert.Interface(false),
						"True":    testconvert.Interface(true),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want: model.VariationResult[bool]{
				VariationType: "False",
				Failed:        false,
				Reason:        flag.ReasonTargetingMatchSplit,
				Value:         false,
				TrackEvents:   true,
				Cacheable:     true,
				Metadata: map[string]any{
					"evaluatedRuleName": "rule1",
				},
			},
			wantErr:     false,
			expectedLog: `user="random-key-ssss1", flag="test-flag", value="false", variation="False"`,
		},
		{
			name: "Get default value, when rule apply and not right type",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key-ssss1"),
				defaultValue:  true,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("key eq \"random-key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface("xxx"),
						"False":   testconvert.Interface("xxx"),
						"True":    testconvert.Interface("xxx"),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want: model.VariationResult[bool]{
				VariationType: flag.VariationSDKDefault,
				Failed:        true,
				Reason:        flag.ReasonError,
				ErrorCode:     flag.ErrorCodeTypeMismatch,
				Value:         false,
				TrackEvents:   true,
			},
			wantErr:     true,
			expectedLog: `user="random-key-ssss1", flag="test-flag", value="true", variation="SdkDefault"`,
		},
		{
			name: "Get sdk default value if offline",
			args: args{
				offline:       true,
				flagKey:       "disable-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  false,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Disable: testconvert.Bool(true),
				}, nil),
			},
			want: model.VariationResult[bool]{
				VariationType: flag.VariationSDKDefault,
				Failed:        false,
				Reason:        flag.ReasonOffline,
				Value:         false,
				TrackEvents:   false,
			},
			wantErr:     false,
			expectedLog: "",
		},
		{
			name: "Evaluation without targetingKey: success",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewEvaluationContext(""),
				defaultValue:  true,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Variations: &map[string]*any{
						"on":  testconvert.Interface(true),
						"off": testconvert.Interface(false),
					},
					DefaultRule: &flag.Rule{
						VariationResult: testconvert.String("on"),
					},
				}, nil),
			},
			want: model.VariationResult[bool]{
				VariationType: "on",
				Failed:        false,
				Reason:        flag.ReasonStatic,
				Value:         true,
				TrackEvents:   true,
				Cacheable:     true,
			},
			wantErr:     false,
			expectedLog: `user="", flag="test-flag", value="true", variation="on"`,
		},
		{
			name: "Evaluation without targetingKey: fail because flag requires bucketing",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewEvaluationContext(""),
				defaultValue:  false,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Variations: &map[string]*any{
						"on":  testconvert.Interface(true),
						"off": testconvert.Interface(false),
					},
					DefaultRule: &flag.Rule{
						Percentages: &map[string]float64{
							"on":  100,
							"off": 0,
						},
					},
				}, nil),
			},
			want: model.VariationResult[bool]{
				VariationType: "SdkDefault",
				Failed:        true,
				Reason:        flag.ReasonError,
				ErrorCode:     flag.ErrorCodeTargetingKeyMissing,
				ErrorDetails:  "Error: Empty targeting key",
				Value:         false,
				TrackEvents:   true,
				Cacheable:     false,
			},
			wantErr:     false,
			expectedLog: `user="", flag="test-flag", value="false", variation="SdkDefault"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := slogassert.New(t, slog.LevelInfo, nil)
			logger := slog.New(handler)

			if !tt.args.disableInit {
				c := retriever.ManagerConfig{
					FileFormat:              "YAML",
					DisableNotifierOnInit:   false,
					StartWithRetrieverError: false,
					EnablePollingJitter:     true,
					PollingInterval:         500,
				}
				ff = &GoFeatureFlag{
					retrieverManager: retriever.NewManager(c, []retriever.Retriever{}, tt.args.cacheMock, &fflog.FFLogger{LeveledLogger: logger}),
					config: Config{
						PollingInterval: 0,
						LeveledLogger:   logger,
						Offline:         tt.args.offline,
					},
					featureEventDataExporter: exporter.NewManager[exporter.FeatureEvent](
						[]exporter.Config{exporter.Config{
							FlushInterval:    0,
							MaxEventInMemory: 0,
							Exporter: &logsexporter.Exporter{
								LogFormat: "user=\"{{ .UserKey}}\", flag=\"{{ .Key}}\", " +
									"value=\"{{ .Value}}\", variation=\"{{ .Variation}}\"",
							},
						}},
						exporter.DefaultExporterCleanQueueInterval,
						&fflog.FFLogger{LeveledLogger: logger},
					),
				}
			}

			got, err := BoolVariationDetails(tt.args.flagKey, tt.args.evaluationCtx, tt.args.defaultValue)

			if tt.expectedLog != "" {
				time.Sleep(
					40 * time.Millisecond,
				) // since the log is async, we are waiting to be sure it's written
				if tt.expectedLog == "" {
					handler.AssertEmpty()
				} else {
					handler.Assert(func(message slogassert.LogMessage) bool {
						if !strings.Contains(message.Message, tt.expectedLog) {
							handler.Fail("impossible to find %s in %s", tt.expectedLog, message.Message)
							return false
						}
						return true
					})
				}
			}

			if tt.wantErr {
				assert.Error(
					t,
					err,
					"BoolVariationDetails() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
				return
			}
			assert.Equal(t, tt.want, got, "BoolVariationDetails() got = %v, want %v", got, tt.want)

			// clean logger
			ff = nil
		})
	}
}

func TestFloat64Variation(t *testing.T) {
	type args struct {
		flagKey       string
		evaluationCtx ffcontext.Context
		defaultValue  float64
		cacheMock     cache.Manager
		offline       bool
		disableInit   bool
	}
	tests := []struct {
		name        string
		args        args
		want        float64
		wantErr     bool
		expectedLog string
	}{
		{
			name: "Call variation before init of SDK",
			args: args{
				flagKey:       "key-not-exist",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  123.3,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Variations: &map[string]*any{
						"var_true":  testconvert.Interface(true),
						"var_false": testconvert.Interface(false),
					},
					Rules: &[]flag.Rule{{
						Query:       testconvert.String("key eq \"key\""),
						Percentages: &map[string]float64{"var_true": 100, "var_false": 0}},
					},
					BucketingKey: nil,
					DefaultRule: &flag.Rule{
						VariationResult: testconvert.String("var_false"),
					},
				}, nil),
				disableInit: true,
			},
			want:    123.3,
			wantErr: true,
		},
		{
			name: "Get default value if flag disable",
			args: args{
				flagKey:       "disable-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  120.12,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Disable: testconvert.Bool(true),
				}, nil),
			},
			want:        120.12,
			wantErr:     false,
			expectedLog: `user="random-key", flag="disable-flag", value="120.12", variation="SdkDefault"`,
		},
		{
			name: "Get error when cache not init",
			args: args{
				flagKey:       "key-not-exist",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  118.12,
				cacheMock: NewCacheMock(
					&flag.InternalFlag{},
					errors.New("impossible to read the toggle before the initialisation")),
			},
			want:        118.12,
			wantErr:     true,
			expectedLog: `user="random-key", flag="key-not-exist", value="118.12", variation="SdkDefault"`,
		},
		{
			name: "Get default value with key not exist",
			args: args{
				flagKey:       "key-not-exist",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  118.12,
				cacheMock: NewCacheMock(
					&flag.InternalFlag{},
					errors.New("flag [key-not-exist] does not exists"),
				),
			},
			want:        118.12,
			wantErr:     true,
			expectedLog: `user="random-key", flag="key-not-exist", value="118.12", variation="SdkDefault"`,
		},
		{
			name: "Get default value, rule not apply",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  118.12,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("key eq \"key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface(119.12),
						"False":   testconvert.Interface(121.12),
						"True":    testconvert.Interface(120.12),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want:        119.12,
			wantErr:     false,
			expectedLog: `user="random-key", flag="test-flag", value="119.12", variation="Default"`,
		},
		{
			name: "Get true value, rule apply",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewAnonymousEvaluationContext("random-key"),
				defaultValue:  118.12,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("key eq \"random-key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface(119.12),
						"False":   testconvert.Interface(121.12),
						"True":    testconvert.Interface(120.12),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want:        120.12,
			wantErr:     false,
			expectedLog: `user="random-key", flag="test-flag", value="120.12", variation="True"`,
		},
		{
			name: "Get false value, rule apply",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewAnonymousEvaluationContext("random-key-ssss1"),
				defaultValue:  118.12,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("anonymous eq true"),
							Percentages: &map[string]float64{
								"False": 90,
								"True":  10,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface(119.12),
						"False":   testconvert.Interface(121.12),
						"True":    testconvert.Interface(120.12),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want:        121.12,
			wantErr:     false,
			expectedLog: `user="random-key-ssss1", flag="test-flag", value="121.12", variation="False"`,
		},
		{
			name: "Get default value, when rule apply and not right type",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key-ssss1"),
				defaultValue:  118.12,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("key eq \"random-key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface("xxx"),
						"False":   testconvert.Interface("xxx"),
						"True":    testconvert.Interface("xxx"),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want:        118.12,
			wantErr:     true,
			expectedLog: `user="random-key-ssss1", flag="test-flag", value="118.12", variation="SdkDefault"`,
		},
		{
			name: "No exported log",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewAnonymousEvaluationContext("random-key"),
				defaultValue:  118.12,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("key eq \"random-key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface(119.12),
						"False":   testconvert.Interface(121.12),
						"True":    testconvert.Interface(120.12),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
					TrackEvents: testconvert.Bool(false),
				}, nil),
			},
			want:        120.12,
			wantErr:     false,
			expectedLog: "",
		},
		{
			name: "Get sdk default value if offline",
			args: args{
				offline:       true,
				flagKey:       "disable-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  118.12,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Disable: testconvert.Bool(true),
				}, nil),
			},
			want:        118.12,
			wantErr:     false,
			expectedLog: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := slogassert.New(t, slog.LevelInfo, nil)
			logger := slog.New(handler)

			if !tt.args.disableInit {
				config := retriever.ManagerConfig{
					FileFormat:              "YAML",
					DisableNotifierOnInit:   false,
					StartWithRetrieverError: false,
					EnablePollingJitter:     true,
					PollingInterval:         500,
				}
				ff = &GoFeatureFlag{
					retrieverManager: retriever.NewManager(config, []retriever.Retriever{}, tt.args.cacheMock, &fflog.FFLogger{LeveledLogger: logger}),
					config: Config{
						PollingInterval: 0,
						LeveledLogger:   logger,
						Offline:         tt.args.offline,
					},
					featureEventDataExporter: exporter.NewManager[exporter.FeatureEvent](
						[]exporter.Config{exporter.Config{
							FlushInterval:    0,
							MaxEventInMemory: 0,
							Exporter: &logsexporter.Exporter{
								LogFormat: "user=\"{{ .UserKey}}\", flag=\"{{ .Key}}\", " +
									"value=\"{{ .Value}}\", variation=\"{{ .Variation}}\"",
							},
						}},
						exporter.DefaultExporterCleanQueueInterval,
						&fflog.FFLogger{LeveledLogger: logger},
					),
				}
			}

			got, err := Float64Variation(tt.args.flagKey, tt.args.evaluationCtx, tt.args.defaultValue)

			if tt.expectedLog != "" {
				time.Sleep(
					40 * time.Millisecond,
				) // since the log is async, we are waiting to be sure it's written
				if tt.expectedLog == "" {
					handler.AssertEmpty()
				} else {
					handler.Assert(func(message slogassert.LogMessage) bool {
						if !strings.Contains(message.Message, tt.expectedLog) {
							handler.Fail("impossible to find %s in %s", tt.expectedLog, message.Message)
							return false
						}
						return true
					})
				}
			}
			if tt.wantErr {
				assert.Error(t, err, "Float64Variation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got, "Float64Variation() got = %v, want %v", got, tt.want)

			// clean logger
			ff = nil
		})
	}
}

func TestFloat64VariationDetails(t *testing.T) {
	type args struct {
		flagKey       string
		evaluationCtx ffcontext.Context
		defaultValue  float64
		cacheMock     cache.Manager
		offline       bool
		disableInit   bool
	}
	tests := []struct {
		name        string
		args        args
		want        model.VariationResult[float64]
		wantErr     bool
		expectedLog string
	}{
		{
			name: "Call variation before init of SDK",
			args: args{
				flagKey:       "key-not-exist",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  123.3,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Variations: &map[string]*any{
						"var_true":  testconvert.Interface(true),
						"var_false": testconvert.Interface(false),
					},
					Rules: &[]flag.Rule{{
						Query:       testconvert.String("key eq \"key\""),
						Percentages: &map[string]float64{"var_true": 100, "var_false": 0}},
					},
					BucketingKey: nil,
					DefaultRule: &flag.Rule{
						VariationResult: testconvert.String("var_false"),
					},
				}, nil),
				disableInit: true,
			},
			wantErr: true,
		},
		{
			name: "Get default value if flag disable",
			args: args{
				flagKey:       "disable-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  120.12,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Disable: testconvert.Bool(true),
				}, nil),
			},
			want: model.VariationResult[float64]{
				Value:         120.12,
				TrackEvents:   true,
				VariationType: flag.VariationSDKDefault,
				Failed:        false,
				Reason:        flag.ReasonDisabled,
				Cacheable:     true,
			},
			wantErr:     false,
			expectedLog: `user="random-key", flag="disable-flag", value="120.12", variation="SdkDefault"`,
		},
		{
			name: "Get error when cache not init",
			args: args{
				flagKey:       "key-not-exist",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  118.12,
				cacheMock: NewCacheMock(
					&flag.InternalFlag{},
					errors.New("impossible to read the toggle before the initialisation")),
			},
			wantErr:     true,
			expectedLog: `user="random-key", flag="key-not-exist", value="118.12", variation="SdkDefault"`,
		},
		{
			name: "Get default value with key not exist",
			args: args{
				flagKey:       "key-not-exist",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  118.12,
				cacheMock: NewCacheMock(
					&flag.InternalFlag{},
					errors.New("flag [key-not-exist] does not exists"),
				),
			},
			wantErr:     true,
			expectedLog: `user="random-key", flag="key-not-exist", value="118.12", variation="SdkDefault"`,
		},
		{
			name: "Get default value, rule not apply",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  118.12,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("key eq \"key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface(119.12),
						"False":   testconvert.Interface(121.12),
						"True":    testconvert.Interface(120.12),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want: model.VariationResult[float64]{
				Value:         119.12,
				TrackEvents:   true,
				VariationType: "Default",
				Failed:        false,
				Reason:        flag.ReasonDefault,
				Cacheable:     true,
			},
			wantErr:     false,
			expectedLog: `user="random-key", flag="test-flag", value="119.12", variation="Default"`,
		},
		{
			name: "Get true value, rule apply",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewAnonymousEvaluationContext("random-key"),
				defaultValue:  118.12,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("key eq \"random-key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface(119.12),
						"False":   testconvert.Interface(121.12),
						"True":    testconvert.Interface(120.12),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want: model.VariationResult[float64]{
				Value:         120.12,
				TrackEvents:   true,
				VariationType: "True",
				Failed:        false,
				Reason:        flag.ReasonTargetingMatch,
				Cacheable:     true,
				Metadata: map[string]any{
					"evaluatedRuleName": "rule1",
				},
			},
			wantErr:     false,
			expectedLog: `user="random-key", flag="test-flag", value="120.12", variation="True"`,
		},
		{
			name: "Get false value, rule apply",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewAnonymousEvaluationContext("random-key-ssss1"),
				defaultValue:  118.12,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("anonymous eq true"),
							Percentages: &map[string]float64{
								"False": 90,
								"True":  10,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface(119.12),
						"False":   testconvert.Interface(121.12),
						"True":    testconvert.Interface(120.12),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want: model.VariationResult[float64]{
				Value:         121.12,
				TrackEvents:   true,
				VariationType: "False",
				Failed:        false,
				Reason:        flag.ReasonTargetingMatchSplit,
				Cacheable:     true,
				Metadata: map[string]any{
					"evaluatedRuleName": "rule1",
				},
			},
			wantErr:     false,
			expectedLog: `user="random-key-ssss1", flag="test-flag", value="121.12", variation="False"`,
		},
		{
			name: "Get default value, when rule apply and not right type",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key-ssss1"),
				defaultValue:  118.12,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("key eq \"random-key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface("xxx"),
						"False":   testconvert.Interface("xxx"),
						"True":    testconvert.Interface("xxx"),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want: model.VariationResult[float64]{
				Value:         118.12,
				TrackEvents:   true,
				VariationType: "Default",
				Failed:        false,
				Reason:        flag.ReasonDefault,
			},
			wantErr:     true,
			expectedLog: `user="random-key-ssss1", flag="test-flag", value="118.12", variation="SdkDefault"`,
		},
		{
			name: "Get sdk default value if offline",
			args: args{
				offline:       true,
				flagKey:       "disable-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  118.12,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Disable: testconvert.Bool(true),
				}, nil),
			},
			want: model.VariationResult[float64]{
				Value:         118.12,
				TrackEvents:   false,
				VariationType: flag.VariationSDKDefault,
				Failed:        false,
				Reason:        flag.ReasonOffline,
			},
			wantErr:     false,
			expectedLog: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := slogassert.New(t, slog.LevelInfo, nil)
			logger := slog.New(handler)

			if !tt.args.disableInit {
				c := retriever.ManagerConfig{
					FileFormat:              "YAML",
					DisableNotifierOnInit:   false,
					StartWithRetrieverError: false,
					EnablePollingJitter:     true,
					PollingInterval:         500,
				}
				ff = &GoFeatureFlag{
					retrieverManager: retriever.NewManager(c, []retriever.Retriever{}, tt.args.cacheMock, &fflog.FFLogger{LeveledLogger: logger}),
					config: Config{
						PollingInterval: 0,
						LeveledLogger:   logger,
						Offline:         tt.args.offline,
					},
					featureEventDataExporter: exporter.NewManager[exporter.FeatureEvent](
						[]exporter.Config{exporter.Config{
							FlushInterval:    0,
							MaxEventInMemory: 0,
							Exporter: &logsexporter.Exporter{
								LogFormat: "user=\"{{ .UserKey}}\", flag=\"{{ .Key}}\", " +
									"value=\"{{ .Value}}\", variation=\"{{ .Variation}}\"",
							},
						}},
						exporter.DefaultExporterCleanQueueInterval,
						&fflog.FFLogger{LeveledLogger: logger},
					),
				}
			}

			got, err := Float64VariationDetails(tt.args.flagKey, tt.args.evaluationCtx, tt.args.defaultValue)

			if tt.expectedLog != "" {
				time.Sleep(
					40 * time.Millisecond,
				) // since the log is async, we are waiting to be sure it's written
				if tt.expectedLog == "" {
					handler.AssertEmpty()
				} else {
					handler.Assert(func(message slogassert.LogMessage) bool {
						if !strings.Contains(message.Message, tt.expectedLog) {
							handler.Fail("impossible to find %s in %s", tt.expectedLog, message.Message)
							return false
						}
						return true
					})
				}
			}
			if tt.wantErr {
				assert.Error(t, err, "Float64Variation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got, "Float64Variation() got = %v, want %v", got, tt.want)

			// clean logger
			ff = nil
		})
	}
}

func TestJSONArrayVariation(t *testing.T) {
	type args struct {
		flagKey       string
		evaluationCtx ffcontext.Context
		defaultValue  []any
		cacheMock     cache.Manager
		offline       bool
		disableInit   bool
	}
	tests := []struct {
		name        string
		args        args
		want        []any
		wantErr     bool
		expectedLog string
	}{
		{
			name: "Call variation before init of SDK",
			args: args{
				flagKey:       "key-not-exist",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  []any{},
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Variations: &map[string]*any{
						"var_true":  testconvert.Interface(true),
						"var_false": testconvert.Interface(false),
					},
					Rules: &[]flag.Rule{{
						Query:       testconvert.String("key eq \"key\""),
						Percentages: &map[string]float64{"var_true": 100, "var_false": 0}},
					},
					BucketingKey: nil,
					DefaultRule: &flag.Rule{
						VariationResult: testconvert.String("var_false"),
					},
				}, nil),
				disableInit: true,
			},
			want:    []any{},
			wantErr: true,
		},
		{
			name: "Get default value if flag disable",
			args: args{
				flagKey:       "disable-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  []any{"toto"},
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Disable: testconvert.Bool(true),
				}, nil),
			},
			want:        []any{"toto"},
			wantErr:     false,
			expectedLog: `user="random-key", flag="disable-flag", value="[toto]", variation="SdkDefault"`,
		},
		{
			name: "Get error when cache not init",
			args: args{
				flagKey:       "key-not-exist",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  []any{"toto"},
				cacheMock: NewCacheMock(
					&flag.InternalFlag{},
					errors.New("impossible to read the toggle before the initialisation")),
			},
			want:        []any{"toto"},
			wantErr:     true,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key\", flag=\"key-not-exist\", value=\"\\[toto\\]\"\n",
		},
		{
			name: "Get default value with key not exist",
			args: args{
				flagKey:       "key-not-exist",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  []any{"toto"},
				cacheMock: NewCacheMock(
					&flag.InternalFlag{},
					errors.New("flag [key-not-exist] does not exists"),
				),
			},
			want:        []any{"toto"},
			wantErr:     true,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key\", flag=\"key-not-exist\", value=\"\\[toto\\]\"\n",
		},
		{
			name: "Get default value, rule not apply",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  []any{"toto"},
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("key eq \"key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface([]any{"default"}),
						"True":    testconvert.Interface([]any{"true"}),
						"False":   testconvert.Interface([]any{"false"}),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want:        []any{"default"},
			wantErr:     false,
			expectedLog: `user="random-key", flag="test-flag", value="[default]", variation="Default"`,
		},
		{
			name: "Get true value, rule apply",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewAnonymousEvaluationContext("random-key"),
				defaultValue:  []any{"toto"},
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("key eq \"random-key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface([]any{"default"}),
						"True":    testconvert.Interface([]any{"true"}),
						"False":   testconvert.Interface([]any{"false"}),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want:        []any{"true"},
			wantErr:     false,
			expectedLog: `user="random-key", flag="test-flag", value="[true]", variation="True"`,
		},
		{
			name: "Get false value, rule apply",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewAnonymousEvaluationContext("random-key-ssss1"),
				defaultValue:  []any{"toto"},
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Variations: &map[string]*any{
						"Default": testconvert.Interface([]any{"default"}),
						"True":    testconvert.Interface([]any{"true"}),
						"False":   testconvert.Interface([]any{"false"}),
					},
					DefaultRule: &flag.Rule{
						Name: testconvert.String("defaultRule"),
						Percentages: &map[string]float64{
							"False": 90,
							"True":  10,
						},
					},
				}, nil),
			},
			want:        []any{"false"},
			wantErr:     false,
			expectedLog: `user="random-key-ssss1", flag="test-flag", value="[false]", variation="False"`,
		},
		{
			name: "Get default value, when rule apply and not right type",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key-ssss1"),
				defaultValue:  []any{"toto"},
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("key eq \"key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface("xxx"),
						"False":   testconvert.Interface("xxx"),
						"True":    testconvert.Interface("xxx"),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want:        []any{"toto"},
			wantErr:     true,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key-ssss1\", flag=\"test-flag\", value=\"\\[toto\\]\"\n",
		},
		{
			name: "No exported log",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key-ssss1"),
				defaultValue:  []any{"toto"},
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Variations: &map[string]*any{
						"Default": testconvert.Interface([]any{"default"}),
						"True":    testconvert.Interface([]any{"true"}),
						"False":   testconvert.Interface([]any{"false"}),
					},
					DefaultRule: &flag.Rule{
						Name: testconvert.String("defaultRule"),
						Percentages: &map[string]float64{
							"False": 0,
							"True":  100,
						},
					},
					TrackEvents: testconvert.Bool(false),
				}, nil),
			},
			want:        []any{"true"},
			wantErr:     false,
			expectedLog: "",
		},
		{
			name: "Get sdk default value if offline",
			args: args{
				offline:       true,
				flagKey:       "disable-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  []any{"toto"},
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Disable: testconvert.Bool(true),
				}, nil),
			},
			want:        []any{"toto"},
			wantErr:     false,
			expectedLog: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := slogassert.New(t, slog.LevelInfo, nil)
			logger := slog.New(handler)

			if !tt.args.disableInit {
				c := retriever.ManagerConfig{
					FileFormat:              "YAML",
					DisableNotifierOnInit:   false,
					StartWithRetrieverError: false,
					EnablePollingJitter:     true,
					PollingInterval:         500,
				}
				ff = &GoFeatureFlag{
					retrieverManager: retriever.NewManager(c, []retriever.Retriever{}, tt.args.cacheMock, &fflog.FFLogger{LeveledLogger: logger}),
					config: Config{
						PollingInterval: 0,
						LeveledLogger:   logger,
						Offline:         tt.args.offline,
					},
					featureEventDataExporter: exporter.NewManager[exporter.FeatureEvent](
						[]exporter.Config{exporter.Config{
							FlushInterval:    0,
							MaxEventInMemory: 0,
							Exporter: &logsexporter.Exporter{
								LogFormat: "user=\"{{ .UserKey}}\", flag=\"{{ .Key}}\", " +
									"value=\"{{ .Value}}\", variation=\"{{ .Variation}}\"",
							},
						}},
						exporter.DefaultExporterCleanQueueInterval,
						&fflog.FFLogger{LeveledLogger: logger},
					),
				}
			}

			got, err := JSONArrayVariation(tt.args.flagKey, tt.args.evaluationCtx, tt.args.defaultValue)

			if tt.wantErr {
				assert.Error(t, err, "JSONArrayVariation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got, "JSONArrayVariation() got = %v, want %v", got, tt.want)
			if tt.expectedLog != "" {
				time.Sleep(
					40 * time.Millisecond,
				) // since the log is async, we are waiting to be sure it's written
				if tt.expectedLog == "" {
					handler.AssertEmpty()
				} else {
					handler.Assert(func(message slogassert.LogMessage) bool {
						if !strings.Contains(message.Message, tt.expectedLog) {
							handler.Fail("impossible to find %s in %s", tt.expectedLog, message.Message)
							return false
						}
						return true
					})
				}
			}
			// clean logger
			ff = nil
		})
	}
}

func TestJSONArrayVariationDetails(t *testing.T) {
	type args struct {
		flagKey       string
		evaluationCtx ffcontext.Context
		defaultValue  []any
		cacheMock     cache.Manager
		offline       bool
		disableInit   bool
	}
	tests := []struct {
		name        string
		args        args
		want        model.VariationResult[[]any]
		wantErr     bool
		expectedLog string
	}{
		{
			name: "Call variation before init of SDK",
			args: args{
				flagKey:       "key-not-exist",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  []any{},
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Variations: &map[string]*any{
						"var_true":  testconvert.Interface(true),
						"var_false": testconvert.Interface(false),
					},
					Rules: &[]flag.Rule{{
						Query:       testconvert.String("key eq \"key\""),
						Percentages: &map[string]float64{"var_true": 100, "var_false": 0}},
					},
					BucketingKey: nil,
					DefaultRule: &flag.Rule{
						VariationResult: testconvert.String("var_false"),
					},
				}, nil),
				disableInit: true,
			},
			wantErr: true,
		},
		{
			name: "Get default value if flag disable",
			args: args{
				flagKey:       "disable-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  []any{"toto"},
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Disable: testconvert.Bool(true),
				}, nil),
			},
			want: model.VariationResult[[]any]{
				TrackEvents:   true,
				VariationType: flag.VariationSDKDefault,
				Failed:        false,
				Reason:        flag.ReasonDisabled,
				Value:         []any{"toto"},
				Cacheable:     true,
			},
			wantErr:     false,
			expectedLog: `user="random-key", flag="disable-flag", value="[toto]", variation="SdkDefault"`,
		},
		{
			name: "Get error when cache not init",
			args: args{
				flagKey:       "key-not-exist",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  []any{"toto"},
				cacheMock: NewCacheMock(
					&flag.InternalFlag{},
					errors.New("impossible to read the toggle before the initialisation")),
			},
			wantErr:     true,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key\", flag=\"key-not-exist\", value=\"\\[toto\\]\"\n",
		},
		{
			name: "Get default value with key not exist",
			args: args{
				flagKey:       "key-not-exist",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  []any{"toto"},
				cacheMock: NewCacheMock(
					&flag.InternalFlag{},
					errors.New("flag [key-not-exist] does not exists"),
				),
			},
			wantErr:     true,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key\", flag=\"key-not-exist\", value=\"\\[toto\\]\"\n",
		},
		{
			name: "Get default value, rule not apply",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  []any{"toto"},
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("key eq \"key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface([]any{"default"}),
						"True":    testconvert.Interface([]any{"true"}),
						"False":   testconvert.Interface([]any{"false"}),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want: model.VariationResult[[]any]{
				TrackEvents:   true,
				VariationType: "Default",
				Failed:        false,
				Reason:        flag.ReasonDefault,
				Value:         []any{"default"},
				Cacheable:     true,
			},
			wantErr:     false,
			expectedLog: `user="random-key", flag="test-flag", value="[default]", variation="Default"`,
		},
		{
			name: "Get true value, rule apply",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewAnonymousEvaluationContext("random-key"),
				defaultValue:  []any{"toto"},
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("key eq \"random-key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface([]any{"default"}),
						"True":    testconvert.Interface([]any{"true"}),
						"False":   testconvert.Interface([]any{"false"}),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want: model.VariationResult[[]any]{
				TrackEvents:   true,
				VariationType: "True",
				Failed:        false,
				Reason:        flag.ReasonTargetingMatch,
				Value:         []any{"true"},
				Cacheable:     true,
				Metadata: map[string]any{
					"evaluatedRuleName": "rule1",
				},
			},
			wantErr:     false,
			expectedLog: `user="random-key", flag="test-flag", value="[true]", variation="True"`,
		},
		{
			name: "Get false value, rule apply",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewAnonymousEvaluationContext("random-key-ssss1"),
				defaultValue:  []any{"toto"},
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Variations: &map[string]*any{
						"Default": testconvert.Interface([]any{"default"}),
						"True":    testconvert.Interface([]any{"true"}),
						"False":   testconvert.Interface([]any{"false"}),
					},
					DefaultRule: &flag.Rule{
						Name: testconvert.String("defaultRule"),
						Percentages: &map[string]float64{
							"False": 90,
							"True":  10,
						},
					},
				}, nil),
			},
			want: model.VariationResult[[]any]{
				TrackEvents:   true,
				VariationType: "False",
				Failed:        false,
				Reason:        flag.ReasonSplit,
				Value:         []any{"false"},
				Cacheable:     true,
			},
			wantErr:     false,
			expectedLog: `user="random-key-ssss1", flag="test-flag", value="[false]", variation="False"`,
		},
		{
			name: "Get default value, when rule apply and not right type",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key-ssss1"),
				defaultValue:  []any{"toto"},
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("key eq \"key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface("xxx"),
						"False":   testconvert.Interface("xxx"),
						"True":    testconvert.Interface("xxx"),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want: model.VariationResult[[]any]{
				TrackEvents:   true,
				VariationType: "Default",
				Failed:        false,
				Reason:        flag.ReasonDefault,
				Value:         []any{"toto"},
			},
			wantErr:     true,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key-ssss1\", flag=\"test-flag\", value=\"\\[toto\\]\"\n",
		},
		{
			name: "Get sdk default value if offline",
			args: args{
				offline:       true,
				flagKey:       "disable-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  []any{"toto"},
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Disable: testconvert.Bool(true),
				}, nil),
			},
			want: model.VariationResult[[]any]{
				TrackEvents:   false,
				VariationType: flag.VariationSDKDefault,
				Failed:        false,
				Reason:        flag.ReasonOffline,
				Value:         []any{"toto"},
			},
			wantErr:     false,
			expectedLog: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := slogassert.New(t, slog.LevelInfo, nil)
			logger := slog.New(handler)

			if !tt.args.disableInit {
				c := retriever.ManagerConfig{
					FileFormat:              "YAML",
					DisableNotifierOnInit:   false,
					StartWithRetrieverError: false,
					EnablePollingJitter:     true,
					PollingInterval:         500,
				}
				ff = &GoFeatureFlag{
					retrieverManager: retriever.NewManager(c, []retriever.Retriever{}, tt.args.cacheMock, &fflog.FFLogger{LeveledLogger: logger}),
					config: Config{
						PollingInterval: 0,
						LeveledLogger:   logger,
						Offline:         tt.args.offline,
					},
					featureEventDataExporter: exporter.NewManager[exporter.FeatureEvent](
						[]exporter.Config{exporter.Config{
							FlushInterval:    0,
							MaxEventInMemory: 0,
							Exporter: &logsexporter.Exporter{
								LogFormat: "user=\"{{ .UserKey}}\", flag=\"{{ .Key}}\", " +
									"value=\"{{ .Value}}\", variation=\"{{ .Variation}}\"",
							},
						}},
						exporter.DefaultExporterCleanQueueInterval,
						&fflog.FFLogger{LeveledLogger: logger},
					),
				}
			}

			got, err := JSONArrayVariationDetails(
				tt.args.flagKey,
				tt.args.evaluationCtx,
				tt.args.defaultValue,
			)

			if tt.wantErr {
				assert.Error(t, err, "JSONArrayVariation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got, "JSONArrayVariation() got = %v, want %v", got, tt.want)
			if tt.expectedLog != "" {
				time.Sleep(
					40 * time.Millisecond,
				) // since the log is async, we are waiting to be sure it's written
				if tt.expectedLog == "" {
					handler.AssertEmpty()
				} else {
					handler.Assert(func(message slogassert.LogMessage) bool {
						if !strings.Contains(message.Message, tt.expectedLog) {
							handler.Fail("impossible to find %s in %s", tt.expectedLog, message.Message)
							return false
						}
						return true
					})
				}
			}
			// clean logger
			ff = nil
		})
	}
}

func TestJSONVariation(t *testing.T) {
	type args struct {
		flagKey       string
		evaluationCtx ffcontext.Context
		defaultValue  map[string]any
		cacheMock     cache.Manager
		offline       bool
		disableInit   bool
	}
	tests := []struct {
		name        string
		args        args
		want        map[string]any
		wantErr     bool
		expectedLog string
	}{
		{
			name: "Call variation before init of SDK",
			args: args{
				flagKey:       "key-not-exist",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  map[string]any{},
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Variations: &map[string]*any{
						"var_true":  testconvert.Interface(true),
						"var_false": testconvert.Interface(false),
					},
					Rules: &[]flag.Rule{{
						Query:       testconvert.String("key eq \"key\""),
						Percentages: &map[string]float64{"var_true": 100, "var_false": 0}},
					},
					BucketingKey: nil,
					DefaultRule: &flag.Rule{
						VariationResult: testconvert.String("var_false"),
					},
				}, nil),
				disableInit: true,
			},
			want:    map[string]any{},
			wantErr: true,
		},
		{
			name: "Get default value if flag disable",
			args: args{
				flagKey:       "disable-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  map[string]any{"default-notkey": true},
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Disable: testconvert.Bool(true),
				}, nil),
			},
			want:        map[string]any{"default-notkey": true},
			wantErr:     false,
			expectedLog: `user="random-key", flag="disable-flag", value="map[default-notkey:true]", variation="SdkDefault"`,
		},
		{
			name: "Get error when cache not init",
			args: args{
				flagKey:       "key-not-exist",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  map[string]any{"default-notkey": true},
				cacheMock: NewCacheMock(
					&flag.InternalFlag{},
					errors.New("impossible to read the toggle before the initialisation")),
			},
			want:        map[string]any{"default-notkey": true},
			wantErr:     true,
			expectedLog: `user="random-key", flag="key-not-exist", value="map[default-notkey:true]", variation="SdkDefault"`,
		},
		{
			name: "Get default value with key not exist",
			args: args{
				flagKey:       "key-not-exist",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  map[string]any{"default-notkey": true},
				cacheMock: NewCacheMock(
					&flag.InternalFlag{},
					errors.New("flag [key-not-exist] does not exists"),
				),
			},
			want:        map[string]any{"default-notkey": true},
			wantErr:     true,
			expectedLog: `user="random-key", flag="key-not-exist", value="map[default-notkey:true]", variation="SdkDefault"`,
		},
		{
			name: "Get default value, rule not apply",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  map[string]any{"default-notkey": true},
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("key eq \"key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface(map[string]any{"default": true}),
						"True":    testconvert.Interface(map[string]any{"true": true}),
						"False":   testconvert.Interface(map[string]any{"false": true}),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want:        map[string]any{"default": true},
			wantErr:     false,
			expectedLog: `user="random-key", flag="test-flag", value="map[default:true]", variation="Default"`,
		},
		{
			name: "Get true value, rule apply",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewAnonymousEvaluationContext("random-key"),
				defaultValue:  map[string]any{"default-notkey": true},
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("key eq \"random-key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface(map[string]any{"default": true}),
						"True":    testconvert.Interface(map[string]any{"true": true}),
						"False":   testconvert.Interface(map[string]any{"false": true}),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want:        map[string]any{"true": true},
			wantErr:     false,
			expectedLog: `user="random-key", flag="test-flag", value="map[true:true]", variation="True"`,
		},
		{
			name: "Get false value, rule apply",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewAnonymousEvaluationContext("random-key-ssss1"),
				defaultValue:  map[string]any{"default-notkey": true},
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("anonymous eq true"),
							Percentages: &map[string]float64{
								"False": 90,
								"True":  10,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface(map[string]any{"default": true}),
						"True":    testconvert.Interface(map[string]any{"true": true}),
						"False":   testconvert.Interface(map[string]any{"false": true}),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want:        map[string]any{"false": true},
			wantErr:     false,
			expectedLog: `user="random-key-ssss1", flag="test-flag", value="map[false:true]", variation="False"`,
		},
		{
			name: "Get default value, when rule apply and not right type",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key-ssss1"),
				defaultValue:  map[string]any{"default-notkey": true},
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("key eq \"key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface("xxx"),
						"False":   testconvert.Interface("xxx"),
						"True":    testconvert.Interface("xxx"),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want:        map[string]any{"default-notkey": true},
			wantErr:     true,
			expectedLog: `user="random-key-ssss1", flag="test-flag", value="map[default-notkey:true]", variation="SdkDefault"`,
		},
		{
			name: "Get sdk default value if offline",
			args: args{
				offline:       true,
				flagKey:       "disable-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  map[string]any{"default-notkey": true},
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Disable: testconvert.Bool(true),
				}, nil),
			},
			want:        map[string]any{"default-notkey": true},
			wantErr:     false,
			expectedLog: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := slogassert.New(t, slog.LevelInfo, nil)
			logger := slog.New(handler)

			if !tt.args.disableInit {
				c := retriever.ManagerConfig{
					FileFormat:              "YAML",
					DisableNotifierOnInit:   false,
					StartWithRetrieverError: false,
					EnablePollingJitter:     true,
					PollingInterval:         500,
				}
				ff = &GoFeatureFlag{
					retrieverManager: retriever.NewManager(c, []retriever.Retriever{}, tt.args.cacheMock, &fflog.FFLogger{LeveledLogger: logger}),
					config: Config{
						PollingInterval: 0,
						LeveledLogger:   logger,
						Offline:         tt.args.offline,
					},
					featureEventDataExporter: exporter.NewManager[exporter.FeatureEvent](
						[]exporter.Config{exporter.Config{
							FlushInterval:    0,
							MaxEventInMemory: 0,
							Exporter: &logsexporter.Exporter{
								LogFormat: "user=\"{{ .UserKey}}\", flag=\"{{ .Key}}\", " +
									"value=\"{{ .Value}}\", variation=\"{{ .Variation}}\"",
							},
						}},
						exporter.DefaultExporterCleanQueueInterval,
						&fflog.FFLogger{LeveledLogger: logger},
					),
				}
			}

			got, err := JSONVariation(tt.args.flagKey, tt.args.evaluationCtx, tt.args.defaultValue)

			if tt.expectedLog != "" {
				time.Sleep(
					40 * time.Millisecond,
				) // since the log is async, we are waiting to be sure it's written
				if tt.expectedLog == "" {
					handler.AssertEmpty()
				} else {
					handler.Assert(func(message slogassert.LogMessage) bool {
						if !strings.Contains(message.Message, tt.expectedLog) {
							handler.Fail("impossible to find %s in %s", tt.expectedLog, message.Message)
							return false
						}
						return true
					})
				}
			}

			if tt.wantErr {
				assert.Error(t, err, "JSONVariation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got, "JSONVariation() got = %v, want %v", got, tt.want)

			// clean logger
			ff = nil
		})
	}
}

func TestJSONVariationDetails(t *testing.T) {
	type args struct {
		flagKey       string
		evaluationCtx ffcontext.Context
		defaultValue  map[string]any
		cacheMock     cache.Manager
		offline       bool
		disableInit   bool
	}
	tests := []struct {
		name        string
		args        args
		want        model.VariationResult[map[string]any]
		wantErr     bool
		expectedLog string
	}{
		{
			name: "Get default value if flag disable",
			args: args{
				flagKey:       "disable-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  map[string]any{"default-notkey": true},
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Disable: testconvert.Bool(true),
				}, nil),
			},
			want: model.VariationResult[map[string]any]{
				TrackEvents:   true,
				VariationType: flag.VariationSDKDefault,
				Failed:        false,
				Reason:        flag.ReasonDisabled,
				Value:         map[string]any{"default-notkey": true},
				Cacheable:     true,
			},
			wantErr:     false,
			expectedLog: `user="random-key", flag="disable-flag", value="map[default-notkey:true]", variation="SdkDefault"`,
		},
		{
			name: "Get default value, rule not apply",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  map[string]any{"default-notkey": true},
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("key eq \"key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface(map[string]any{"default": true}),
						"True":    testconvert.Interface(map[string]any{"true": true}),
						"False":   testconvert.Interface(map[string]any{"false": true}),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want: model.VariationResult[map[string]any]{
				TrackEvents:   true,
				VariationType: "Default",
				Failed:        false,
				Reason:        flag.ReasonDefault,
				Value:         map[string]any{"default": true},
				Cacheable:     true,
			},
			wantErr:     false,
			expectedLog: `user="random-key", flag="test-flag", value="map[default:true]", variation="Default"`,
		},
		{
			name: "Get true value, rule apply",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewAnonymousEvaluationContext("random-key"),
				defaultValue:  map[string]any{"default-notkey": true},
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("key eq \"random-key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface(map[string]any{"default": true}),
						"True":    testconvert.Interface(map[string]any{"true": true}),
						"False":   testconvert.Interface(map[string]any{"false": true}),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want: model.VariationResult[map[string]any]{
				TrackEvents:   true,
				VariationType: "True",
				Failed:        false,
				Reason:        flag.ReasonTargetingMatch,
				Value:         map[string]any{"true": true},
				Cacheable:     true,
				Metadata: map[string]any{
					"evaluatedRuleName": "rule1",
				},
			},
			wantErr:     false,
			expectedLog: `user="random-key", flag="test-flag", value="map[true:true]", variation="True"`,
		},
		{
			name: "Get false value, rule apply",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewAnonymousEvaluationContext("random-key-ssss1"),
				defaultValue:  map[string]any{"default-notkey": true},
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("anonymous eq true"),
							Percentages: &map[string]float64{
								"False": 90,
								"True":  10,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface(map[string]any{"default": true}),
						"True":    testconvert.Interface(map[string]any{"true": true}),
						"False":   testconvert.Interface(map[string]any{"false": true}),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want: model.VariationResult[map[string]any]{
				TrackEvents:   true,
				VariationType: "False",
				Failed:        false,
				Reason:        flag.ReasonTargetingMatchSplit,
				Value:         map[string]any{"false": true},
				Cacheable:     true,
				Metadata: map[string]any{
					"evaluatedRuleName": "rule1",
				},
			},
			wantErr:     false,
			expectedLog: `user="random-key-ssss1", flag="test-flag", value="map[false:true]", variation="False"`,
		},
		{
			name: "Get sdk default value if offline",
			args: args{
				offline:       true,
				flagKey:       "disable-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  map[string]any{"default-notkey": true},
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Disable: testconvert.Bool(true),
				}, nil),
			},
			want: model.VariationResult[map[string]any]{
				TrackEvents:   false,
				VariationType: flag.VariationSDKDefault,
				Failed:        false,
				Reason:        flag.ReasonOffline,
				Value:         map[string]any{"default-notkey": true},
			},
			wantErr:     false,
			expectedLog: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := slogassert.New(t, slog.LevelInfo, nil)
			logger := slog.New(handler)

			if !tt.args.disableInit {
				c := retriever.ManagerConfig{
					FileFormat:              "YAML",
					DisableNotifierOnInit:   false,
					StartWithRetrieverError: false,
					EnablePollingJitter:     true,
					PollingInterval:         500,
				}
				ff = &GoFeatureFlag{
					retrieverManager: retriever.NewManager(c, []retriever.Retriever{}, tt.args.cacheMock, &fflog.FFLogger{LeveledLogger: logger}),
					config: Config{
						PollingInterval: 0,
						LeveledLogger:   logger,
						Offline:         tt.args.offline,
					},
					featureEventDataExporter: exporter.NewManager[exporter.FeatureEvent](
						[]exporter.Config{
							exporter.Config{
								FlushInterval:    0,
								MaxEventInMemory: 0,
								Exporter: &logsexporter.Exporter{
									LogFormat: "user=\"{{ .UserKey}}\", flag=\"{{ .Key}}\", " +
										"value=\"{{ .Value}}\", variation=\"{{ .Variation}}\"",
								},
							}},
						exporter.DefaultExporterCleanQueueInterval,
						&fflog.FFLogger{LeveledLogger: logger},
					),
				}
			}

			got, err := JSONVariationDetails(tt.args.flagKey, tt.args.evaluationCtx, tt.args.defaultValue)

			if tt.expectedLog != "" {
				time.Sleep(
					40 * time.Millisecond,
				) // since the log is async, we are waiting to be sure it's written
				if tt.expectedLog == "" {
					handler.AssertEmpty()
				} else {
					handler.Assert(func(message slogassert.LogMessage) bool {
						if !strings.Contains(message.Message, tt.expectedLog) {
							handler.Fail("impossible to find %s in %s", tt.expectedLog, message.Message)
							return false
						}
						return true
					})
				}
			}

			if tt.wantErr {
				assert.Error(t, err, "JSONVariation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got, "JSONVariation() got = %v, want %v", got, tt.want)

			// clean logger
			ff = nil
		})
	}
}

func TestStringVariation(t *testing.T) {
	type args struct {
		flagKey       string
		evaluationCtx ffcontext.Context
		defaultValue  string
		cacheMock     cache.Manager
		offline       bool
		disableInit   bool
	}
	tests := []struct {
		name        string
		args        args
		want        string
		wantErr     bool
		expectedLog string
	}{
		{
			name: "Call variation before init of SDK",
			args: args{
				flagKey:       "key-not-exist",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  "",
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Variations: &map[string]*any{
						"var_true":  testconvert.Interface(true),
						"var_false": testconvert.Interface(false),
					},
					Rules: &[]flag.Rule{{
						Query:       testconvert.String("key eq \"key\""),
						Percentages: &map[string]float64{"var_true": 100, "var_false": 0}},
					},
					BucketingKey: nil,
					DefaultRule: &flag.Rule{
						VariationResult: testconvert.String("var_false"),
					},
				}, nil),
				disableInit: true,
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Get default value if flag disable",
			args: args{
				flagKey:       "disable-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  "default-notkey",
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Disable: testconvert.Bool(true),
				}, nil),
			},
			want:        "default-notkey",
			wantErr:     false,
			expectedLog: `user="random-key", flag="disable-flag", value="default-notkey", variation="SdkDefault"`,
		},
		{
			name: "Get error when cache not init",
			args: args{
				flagKey:       "key-not-exist",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  "default-notkey",
				cacheMock: NewCacheMock(
					&flag.InternalFlag{},
					errors.New("impossible to read the toggle before the initialisation")),
			},
			want:        "default-notkey",
			wantErr:     true,
			expectedLog: `user="random-key", flag="key-not-exist", value="default-notkey", variation="SdkDefault"`,
		},
		{
			name: "Get default value with key not exist",
			args: args{
				flagKey:       "key-not-exist",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  "default-notkey",
				cacheMock: NewCacheMock(
					&flag.InternalFlag{},
					errors.New("flag [key-not-exist] does not exists"),
				),
			},
			want:        "default-notkey",
			wantErr:     true,
			expectedLog: `user="random-key", flag="key-not-exist", value="default-notkey", variation="SdkDefault"`,
		},

		{
			name: "Get default value, rule not apply",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  "default-notkey",
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("key eq \"key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface("default"),
						"True":    testconvert.Interface("true"),
						"False":   testconvert.Interface("false"),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want:        "default",
			wantErr:     false,
			expectedLog: `user="random-key", flag="test-flag", value="default", variation="Default"`,
		},
		{
			name: "Get true value, rule apply",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewAnonymousEvaluationContext("random-key"),
				defaultValue:  "default-notkey",
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("key eq \"random-key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface("default"),
						"True":    testconvert.Interface("true"),
						"False":   testconvert.Interface("false"),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want:        "true",
			wantErr:     false,
			expectedLog: `user="random-key", flag="test-flag", value="true", variation="True"`,
		},
		{
			name: "Get false value, rule apply",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewAnonymousEvaluationContext("random-key-ssss1"),
				defaultValue:  "default-notkey",
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("anonymous eq true"),
							Percentages: &map[string]float64{
								"False": 90,
								"True":  10,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface("default"),
						"True":    testconvert.Interface("true"),
						"False":   testconvert.Interface("false"),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want:        "false",
			wantErr:     false,
			expectedLog: `user="random-key-ssss1", flag="test-flag", value="false", variation="False"`,
		},
		{
			name: "Get default value, when rule apply and not right type",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key-ssss1"),
				defaultValue:  "default-notkey",
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("key eq \"key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface(1),
						"False":   testconvert.Interface(2),
						"True":    testconvert.Interface(3),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want:        "default-notkey",
			wantErr:     true,
			expectedLog: `user="random-key-ssss1", flag="test-flag", value="default-notkey", variation="SdkDefault"`,
		},
		{
			name: "Get sdk default value if offline",
			args: args{
				offline:       true,
				flagKey:       "disable-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  "default-notkey",
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Disable: testconvert.Bool(true),
				}, nil),
			},
			want:        "default-notkey",
			wantErr:     false,
			expectedLog: "",
		},
		{
			name: "Semver prerelease comparison - alpha less than beta",
			args: args{
				flagKey: "test-flag",
				evaluationCtx: ffcontext.NewEvaluationContextBuilder("user-key").
					AddCustom("version", "1.0.0-beta").
					Build(),
				defaultValue: "default",
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:            testconvert.String("semver_rule"),
							Query:           testconvert.String("version gt \"1.0.0-alpha\""),
							VariationResult: testconvert.String("beta"),
						},
					},
					Variations: &map[string]*any{
						"beta":    testconvert.Interface("beta-release"),
						"default": testconvert.Interface("default-value"),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("default"),
					},
				}, nil),
			},
			want:        "beta-release",
			wantErr:     false,
			expectedLog: `user="user-key", flag="test-flag", value="beta-release", variation="beta"`,
		},
		{
			name: "Semver prerelease comparison - numeric prerelease identifiers",
			args: args{
				flagKey: "test-flag",
				evaluationCtx: ffcontext.NewEvaluationContextBuilder("user-key").
					AddCustom("version", "1.0.0-2345").
					Build(),
				defaultValue: "default",
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:            testconvert.String("semver_rule"),
							Query:           testconvert.String("version gt \"1.0.0-1234\""),
							VariationResult: testconvert.String("prerelease"),
						},
					},
					Variations: &map[string]*any{
						"prerelease": testconvert.Interface("newer-prerelease"),
						"default":    testconvert.Interface("default-value"),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("default"),
					},
				}, nil),
			},
			want:        "newer-prerelease",
			wantErr:     false,
			expectedLog: `user="user-key", flag="test-flag", value="newer-prerelease", variation="prerelease"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := slogassert.New(t, slog.LevelInfo, nil)
			logger := slog.New(handler)

			if !tt.args.disableInit {
				c := retriever.ManagerConfig{
					FileFormat:              "YAML",
					DisableNotifierOnInit:   false,
					StartWithRetrieverError: false,
					EnablePollingJitter:     true,
					PollingInterval:         500,
				}
				ff = &GoFeatureFlag{
					retrieverManager: retriever.NewManager(c, []retriever.Retriever{}, tt.args.cacheMock, &fflog.FFLogger{LeveledLogger: logger}),
					config: Config{
						PollingInterval: 0,
						LeveledLogger:   logger,
						Offline:         tt.args.offline,
					},
					featureEventDataExporter: exporter.NewManager[exporter.FeatureEvent](
						[]exporter.Config{
							exporter.Config{
								FlushInterval:    0,
								MaxEventInMemory: 0,
								Exporter: &logsexporter.Exporter{
									LogFormat: "user=\"{{ .UserKey}}\", flag=\"{{ .Key}}\", " +
										"value=\"{{ .Value}}\", variation=\"{{ .Variation}}\"",
								},
							}},
						exporter.DefaultExporterCleanQueueInterval,
						&fflog.FFLogger{LeveledLogger: logger},
					),
				}
			}
			got, err := StringVariation(tt.args.flagKey, tt.args.evaluationCtx, tt.args.defaultValue)

			if tt.expectedLog != "" {
				time.Sleep(
					40 * time.Millisecond,
				) // since the log is async, we are waiting to be sure it's written
				if tt.expectedLog == "" {
					handler.AssertEmpty()
				} else {
					handler.Assert(func(message slogassert.LogMessage) bool {
						if !strings.Contains(message.Message, tt.expectedLog) {
							handler.Fail("impossible to find %s in %s", tt.expectedLog, message.Message)
							return false
						}
						return true
					})
				}
			}

			if tt.wantErr {
				assert.Error(t, err, "StringVariation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got, "StringVariation() got = %v, want %v", got, tt.want)

			// clean logger
			ff = nil
		})
	}
}

func TestStringVariationDetails(t *testing.T) {
	type args struct {
		flagKey       string
		evaluationCtx ffcontext.Context
		defaultValue  string
		cacheMock     cache.Manager
		offline       bool
		disableInit   bool
	}
	tests := []struct {
		name        string
		args        args
		want        model.VariationResult[string]
		wantErr     bool
		expectedLog string
	}{
		{
			name: "Get default value if flag disable",
			args: args{
				flagKey:       "disable-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  "default-notkey",
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Disable: testconvert.Bool(true),
				}, nil),
			},
			want: model.VariationResult[string]{
				TrackEvents:   true,
				VariationType: flag.VariationSDKDefault,
				Failed:        false,
				Reason:        flag.ReasonDisabled,
				Value:         "default-notkey",
				Cacheable:     true,
			},
			wantErr:     false,
			expectedLog: `user="random-key", flag="disable-flag", value="default-notkey", variation="SdkDefault"`,
		},
		{
			name: "Get default value, rule not apply",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  "default-notkey",
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("key eq \"key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface("default"),
						"True":    testconvert.Interface("true"),
						"False":   testconvert.Interface("false"),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want: model.VariationResult[string]{
				TrackEvents:   true,
				VariationType: "Default",
				Failed:        false,
				Reason:        flag.ReasonDefault,
				Value:         "default",
				Cacheable:     true,
			},
			wantErr:     false,
			expectedLog: `user="random-key", flag="test-flag", value="default", variation="Default"`,
		},
		{
			name: "Get true value, rule apply",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewAnonymousEvaluationContext("random-key"),
				defaultValue:  "default-notkey",
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("key eq \"random-key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface("default"),
						"True":    testconvert.Interface("true"),
						"False":   testconvert.Interface("false"),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want: model.VariationResult[string]{
				TrackEvents:   true,
				VariationType: "True",
				Failed:        false,
				Reason:        flag.ReasonTargetingMatch,
				Value:         "true",
				Cacheable:     true,
				Metadata: map[string]any{
					"evaluatedRuleName": "rule1",
				},
			},
			wantErr:     false,
			expectedLog: `user="random-key", flag="test-flag", value="true", variation="True"`,
		},
		{
			name: "Get false value, rule apply",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewAnonymousEvaluationContext("random-key-ssss1"),
				defaultValue:  "default-notkey",
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("anonymous eq true"),
							Percentages: &map[string]float64{
								"False": 90,
								"True":  10,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface("default"),
						"True":    testconvert.Interface("true"),
						"False":   testconvert.Interface("false"),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want: model.VariationResult[string]{
				TrackEvents:   true,
				VariationType: "False",
				Failed:        false,
				Reason:        flag.ReasonTargetingMatchSplit,
				Value:         "false",
				Cacheable:     true,
				Metadata: map[string]any{
					"evaluatedRuleName": "rule1",
				},
			},
			wantErr:     false,
			expectedLog: `user="random-key-ssss1", flag="test-flag", value="false", variation="False"`,
		},
		{
			name: "Get sdk default value if offline",
			args: args{
				offline:       true,
				flagKey:       "disable-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  "default-notkey",
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Disable: testconvert.Bool(true),
				}, nil),
			},
			want: model.VariationResult[string]{
				TrackEvents:   false,
				VariationType: flag.VariationSDKDefault,
				Failed:        false,
				Reason:        flag.ReasonOffline,
				Value:         "default-notkey",
			},
			wantErr:     false,
			expectedLog: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := slogassert.New(t, slog.LevelInfo, nil)
			logger := slog.New(handler)

			if !tt.args.disableInit {
				c := retriever.ManagerConfig{
					FileFormat:              "YAML",
					DisableNotifierOnInit:   false,
					StartWithRetrieverError: false,
					EnablePollingJitter:     true,
					PollingInterval:         500,
				}
				ff = &GoFeatureFlag{
					retrieverManager: retriever.NewManager(c, []retriever.Retriever{}, tt.args.cacheMock, &fflog.FFLogger{LeveledLogger: logger}),
					config: Config{
						PollingInterval: 0,
						LeveledLogger:   logger,
						Offline:         tt.args.offline,
					},
					featureEventDataExporter: exporter.NewManager[exporter.FeatureEvent](
						[]exporter.Config{exporter.Config{
							FlushInterval:    0,
							MaxEventInMemory: 0,
							Exporter: &logsexporter.Exporter{
								LogFormat: "user=\"{{ .UserKey}}\", flag=\"{{ .Key}}\", " +
									"value=\"{{ .Value}}\", variation=\"{{ .Variation}}\"",
							},
						}},
						exporter.DefaultExporterCleanQueueInterval,
						&fflog.FFLogger{LeveledLogger: logger},
					),
				}
			}
			got, err := StringVariationDetails(tt.args.flagKey, tt.args.evaluationCtx, tt.args.defaultValue)

			if tt.expectedLog != "" {
				time.Sleep(
					40 * time.Millisecond,
				) // since the log is async, we are waiting to be sure it's written
				if tt.expectedLog == "" {
					handler.AssertEmpty()
				} else {
					handler.Assert(func(message slogassert.LogMessage) bool {
						if !strings.Contains(message.Message, tt.expectedLog) {
							handler.Fail("impossible to find %s in %s", tt.expectedLog, message.Message)
							return false
						}
						return true
					})
				}
			}

			if tt.wantErr {
				assert.Error(t, err, "StringVariation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got, "StringVariation() got = %v, want %v", got, tt.want)

			// clean logger
			ff = nil
		})
	}
}

func TestIntVariation(t *testing.T) {
	type args struct {
		flagKey       string
		evaluationCtx ffcontext.Context
		defaultValue  int
		cacheMock     cache.Manager
		offline       bool
		disableInit   bool
	}
	tests := []struct {
		name        string
		args        args
		want        int
		wantErr     bool
		expectedLog string
	}{
		{
			name: "Call variation before init of SDK",
			args: args{
				flagKey:       "key-not-exist",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  1,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Variations: &map[string]*any{
						"var_true":  testconvert.Interface(true),
						"var_false": testconvert.Interface(false),
					},
					Rules: &[]flag.Rule{{
						Query:       testconvert.String("key eq \"key\""),
						Percentages: &map[string]float64{"var_true": 100, "var_false": 0}},
					},
					BucketingKey: nil,
					DefaultRule: &flag.Rule{
						VariationResult: testconvert.String("var_false"),
					},
				}, nil),
				disableInit: true,
			},
			want:    1,
			wantErr: true,
		},
		{
			name: "Get default value if flag disable",
			args: args{
				flagKey:       "disable-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  125,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Disable: testconvert.Bool(true),
				}, nil),
			},
			want:        125,
			wantErr:     false,
			expectedLog: `user="random-key", flag="disable-flag", value="125", variation="SdkDefault"`,
		},
		{
			name: "Get error when cache not init",
			args: args{
				flagKey:       "key-not-exist",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  118,
				cacheMock: NewCacheMock(
					&flag.InternalFlag{},
					errors.New("impossible to read the toggle before the initialisation")),
			},
			want:        118,
			wantErr:     true,
			expectedLog: `user="random-key", flag="key-not-exist", value="118", variation="SdkDefault"`,
		},
		{
			name: "Get default value with key not exist",
			args: args{
				flagKey:       "key-not-exist",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  118,
				cacheMock: NewCacheMock(
					&flag.InternalFlag{},
					errors.New("flag [key-not-exist] does not exists"),
				),
			},
			want:        118,
			wantErr:     true,
			expectedLog: `user="random-key", flag="key-not-exist", value="118", variation="SdkDefault"`,
		},
		{
			name: "Get default value rule not apply",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  118,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("key eq \"key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface(119),
						"True":    testconvert.Interface(120),
						"False":   testconvert.Interface(121),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want:        119,
			wantErr:     false,
			expectedLog: `user="random-key", flag="test-flag", value="119", variation="Default"`,
		},
		{
			name: "Get true value, rule apply",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewAnonymousEvaluationContext("random-key"),
				defaultValue:  118,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("key eq \"random-key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface(119),
						"True":    testconvert.Interface(120),
						"False":   testconvert.Interface(121),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want:        120,
			wantErr:     false,
			expectedLog: `user="random-key", flag="test-flag", value="120", variation="True"`,
		},
		{
			name: "Get false value, rule apply",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewAnonymousEvaluationContext("random-key-ssss1"),
				defaultValue:  118,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("anonymous eq true"),
							Percentages: &map[string]float64{
								"False": 90,
								"True":  10,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface(119),
						"True":    testconvert.Interface(120),
						"False":   testconvert.Interface(121),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want:        121,
			wantErr:     false,
			expectedLog: `user="random-key-ssss1", flag="test-flag", value="121", variation="False"`,
		},
		{
			name: "Get default value, when rule apply and not right type",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key-ssss1"),
				defaultValue:  118,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("key eq \"key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface("xxx"),
						"False":   testconvert.Interface("xxx"),
						"True":    testconvert.Interface("xxx"),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want:        118,
			wantErr:     true,
			expectedLog: `user="random-key-ssss1", flag="test-flag", value="118", variation="SdkDefault"`,
		},
		{
			name: "Convert float to Int",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewAnonymousEvaluationContext("random-key"),
				defaultValue:  118,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("key eq \"random-key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface(119.1),
						"True":    testconvert.Interface(120.1),
						"False":   testconvert.Interface(121.1),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want:        120,
			wantErr:     false,
			expectedLog: `user="random-key", flag="test-flag", value="120", variation="True"`,
		},
		{
			name: "Get sdk default value if offline",
			args: args{
				offline:       true,
				flagKey:       "disable-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  125,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Disable: testconvert.Bool(true),
				}, nil),
			},
			want:        125,
			wantErr:     false,
			expectedLog: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := slogassert.New(t, slog.LevelInfo, nil)
			logger := slog.New(handler)

			if !tt.args.disableInit {
				c := retriever.ManagerConfig{
					FileFormat:              "YAML",
					DisableNotifierOnInit:   false,
					StartWithRetrieverError: false,
					EnablePollingJitter:     true,
					PollingInterval:         500,
				}
				ff = &GoFeatureFlag{
					retrieverManager: retriever.NewManager(c, []retriever.Retriever{}, tt.args.cacheMock, &fflog.FFLogger{LeveledLogger: logger}),
					config: Config{
						PollingInterval: 0,
						LeveledLogger:   logger,
						Offline:         tt.args.offline,
					},
					featureEventDataExporter: exporter.NewManager[exporter.FeatureEvent](
						[]exporter.Config{exporter.Config{
							FlushInterval:    0,
							MaxEventInMemory: 0,
							Exporter: &logsexporter.Exporter{
								LogFormat: "user=\"{{ .UserKey}}\", flag=\"{{ .Key}}\", " +
									"value=\"{{ .Value}}\", variation=\"{{ .Variation}}\"",
							},
						}},
						exporter.DefaultExporterCleanQueueInterval,
						&fflog.FFLogger{LeveledLogger: logger},
					),
				}
			}
			got, err := IntVariation(tt.args.flagKey, tt.args.evaluationCtx, tt.args.defaultValue)

			if tt.expectedLog != "" {
				time.Sleep(
					40 * time.Millisecond,
				) // since the log is async, we are waiting to be sure it's written
				if tt.expectedLog == "" {
					handler.AssertEmpty()
				} else {
					handler.Assert(func(message slogassert.LogMessage) bool {
						if !strings.Contains(message.Message, tt.expectedLog) {
							handler.Fail("impossible to find %s in %s", tt.expectedLog, message.Message)
							return false
						}
						return true
					})
				}
			}

			if tt.wantErr {
				assert.Error(t, err, "IntVariation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got, "IntVariation() got = %v, want %v", got, tt.want)

			// clean logger
			ff = nil
		})
	}
}

func TestIntVariationDetails(t *testing.T) {
	type args struct {
		flagKey       string
		evaluationCtx ffcontext.Context
		defaultValue  int
		cacheMock     cache.Manager
		offline       bool
		disableInit   bool
	}
	tests := []struct {
		name        string
		args        args
		want        model.VariationResult[int]
		wantErr     bool
		expectedLog string
	}{
		{
			name: "Get default value if flag disable",
			args: args{
				flagKey:       "disable-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  125,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Disable: testconvert.Bool(true),
				}, nil),
			},
			want: model.VariationResult[int]{
				TrackEvents:   true,
				VariationType: flag.VariationSDKDefault,
				Failed:        false,
				Reason:        flag.ReasonDisabled,
				Value:         125,
				Cacheable:     true,
			},
			wantErr:     false,
			expectedLog: `user="random-key", flag="disable-flag", value="125", variation="SdkDefault"`,
		},
		{
			name: "Get default value rule not apply",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  118,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("key eq \"key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface(119),
						"True":    testconvert.Interface(120),
						"False":   testconvert.Interface(121),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want: model.VariationResult[int]{
				TrackEvents:   true,
				VariationType: "Default",
				Failed:        false,
				Reason:        flag.ReasonDefault,
				Value:         119,
				Cacheable:     true,
			},
			wantErr:     false,
			expectedLog: `user="random-key", flag="test-flag", value="119", variation="Default"`,
		},
		{
			name: "Get true value, rule apply",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewAnonymousEvaluationContext("random-key"),
				defaultValue:  118,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("key eq \"random-key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface(119),
						"True":    testconvert.Interface(120),
						"False":   testconvert.Interface(121),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want: model.VariationResult[int]{
				TrackEvents:   true,
				VariationType: "True",
				Failed:        false,
				Reason:        flag.ReasonTargetingMatch,
				Value:         120,
				Cacheable:     true,
				Metadata: map[string]any{
					"evaluatedRuleName": "rule1",
				},
			},
			wantErr:     false,
			expectedLog: `user="random-key", flag="test-flag", value="120", variation="True"`,
		},
		{
			name: "Get false value, rule apply",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewAnonymousEvaluationContext("random-key-ssss1"),
				defaultValue:  118,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("anonymous eq true"),
							Percentages: &map[string]float64{
								"False": 90,
								"True":  10,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface(119),
						"True":    testconvert.Interface(120),
						"False":   testconvert.Interface(121),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want: model.VariationResult[int]{
				TrackEvents:   true,
				VariationType: "False",
				Failed:        false,
				Reason:        flag.ReasonTargetingMatchSplit,
				Value:         121,
				Cacheable:     true,
				Metadata: map[string]any{
					"evaluatedRuleName": "rule1",
				},
			},
			wantErr:     false,
			expectedLog: `user="random-key-ssss1", flag="test-flag", value="121", variation="False"`,
		},
		{
			name: "Convert float to Int",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewAnonymousEvaluationContext("random-key"),
				defaultValue:  118,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("key eq \"random-key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface(119.1),
						"True":    testconvert.Interface(120.1),
						"False":   testconvert.Interface(121.1),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want: model.VariationResult[int]{
				TrackEvents:   true,
				VariationType: "True",
				Failed:        false,
				Reason:        flag.ReasonTargetingMatch,
				Value:         120,
				Cacheable:     true,
				Metadata: map[string]any{
					"evaluatedRuleName": "rule1",
				},
			},
			wantErr:     false,
			expectedLog: `user="random-key", flag="test-flag", value="120", variation="True"`,
		},
		{
			name: "Get sdk default value if offline",
			args: args{
				offline:       true,
				flagKey:       "disable-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  125,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Disable: testconvert.Bool(true),
				}, nil),
			},
			want: model.VariationResult[int]{
				TrackEvents:   false,
				VariationType: flag.VariationSDKDefault,
				Failed:        false,
				Reason:        flag.ReasonOffline,
				Value:         125,
			},
			wantErr:     false,
			expectedLog: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := slogassert.New(t, slog.LevelInfo, nil)
			logger := slog.New(handler)

			if !tt.args.disableInit {
				c := retriever.ManagerConfig{
					FileFormat:              "YAML",
					DisableNotifierOnInit:   false,
					StartWithRetrieverError: false,
					EnablePollingJitter:     true,
					PollingInterval:         500,
				}
				ff = &GoFeatureFlag{
					retrieverManager: retriever.NewManager(c, []retriever.Retriever{}, tt.args.cacheMock, &fflog.FFLogger{LeveledLogger: logger}),
					config: Config{
						PollingInterval: 0,
						LeveledLogger:   logger,
						Offline:         tt.args.offline,
					},
					featureEventDataExporter: exporter.NewManager[exporter.FeatureEvent](
						[]exporter.Config{exporter.Config{
							FlushInterval:    0,
							MaxEventInMemory: 0,
							Exporter: &logsexporter.Exporter{
								LogFormat: "user=\"{{ .UserKey}}\", flag=\"{{ .Key}}\", " +
									"value=\"{{ .Value}}\", variation=\"{{ .Variation}}\"",
							},
						}},
						exporter.DefaultExporterCleanQueueInterval,
						&fflog.FFLogger{LeveledLogger: logger},
					),
				}
			}
			got, err := IntVariationDetails(tt.args.flagKey, tt.args.evaluationCtx, tt.args.defaultValue)

			if tt.expectedLog != "" {
				time.Sleep(
					40 * time.Millisecond,
				) // since the log is async, we are waiting to be sure it's written
				if tt.expectedLog == "" {
					handler.AssertEmpty()
				} else {
					handler.Assert(func(message slogassert.LogMessage) bool {
						if !strings.Contains(message.Message, tt.expectedLog) {
							handler.Fail("impossible to find %s in %s", tt.expectedLog, message.Message)
							return false
						}
						return true
					})
				}
			}

			if tt.wantErr {
				assert.Error(t, err, "IntVariation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got, "IntVariation() got = %v, want %v", got, tt.want)

			// clean logger
			ff = nil
		})
	}
}

func TestRawVariation(t *testing.T) {
	type args struct {
		flagKey       string
		evaluationCtx ffcontext.Context
		defaultValue  any
		cacheMock     cache.Manager
		offline       bool
		disableInit   bool
	}
	tests := []struct {
		name        string
		args        args
		want        model.RawVarResult
		wantErr     bool
		expectedLog string
	}{
		{
			name: "Call variation before init of SDK",
			args: args{
				flagKey:       "key-not-exist",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  "",
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Variations: &map[string]*any{
						"var_true":  testconvert.Interface(true),
						"var_false": testconvert.Interface(false),
					},
					Rules: &[]flag.Rule{{
						Query:       testconvert.String("key eq \"key\""),
						Percentages: &map[string]float64{"var_true": 100, "var_false": 0}},
					},
					BucketingKey: nil,
					DefaultRule: &flag.Rule{
						VariationResult: testconvert.String("var_false"),
					},
				}, nil),
				disableInit: true,
			},
			wantErr: true,
			want: model.RawVarResult{
				VariationType: flag.VariationSDKDefault,
				Failed:        true,
				Reason:        flag.ReasonError,
				ErrorCode:     flag.ErrorCodeProviderNotReady,
				Value:         "",
			},
		},
		{
			name: "Get default value if flag disable",
			args: args{
				flagKey:       "disable-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  true,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Disable: testconvert.Bool(true),
				}, nil),
			},
			want: model.RawVarResult{
				Value:         true,
				VariationType: flag.VariationSDKDefault,
				Failed:        false,
				TrackEvents:   true,
				Reason:        flag.ReasonDisabled,
				Cacheable:     true,
			},
			wantErr:     false,
			expectedLog: `user="random-key", flag="disable-flag", value="true", variation="SdkDefault"`,
		},
		{
			name: "Get error when cache not init",
			args: args{
				flagKey:       "key-not-exist",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  "defaultValue",
				cacheMock: NewCacheMock(
					&flag.InternalFlag{},
					errors.New("impossible to read the toggle before the initialisation")),
			},
			want: model.RawVarResult{
				Value:         "defaultValue",
				VariationType: flag.VariationSDKDefault,
				Failed:        true,
				TrackEvents:   true,
				Reason:        flag.ReasonError,
				ErrorCode:     flag.ErrorCodeFlagNotFound,
				Cacheable:     false,
			},
			wantErr:     true,
			expectedLog: `user="random-key", flag="key-not-exist", value="defaultValue", variation="SdkDefault"`,
		},
		{
			name: "Get default value with key not exist",
			args: args{
				flagKey:       "key-not-exist",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  123456,
				cacheMock: NewCacheMock(
					&flag.InternalFlag{},
					errors.New("flag [key-not-exist] does not exists"),
				),
			},
			want: model.RawVarResult{
				Value:         123456,
				VariationType: flag.VariationSDKDefault,
				Failed:        true,
				TrackEvents:   true,
				Reason:        flag.ReasonError,
				ErrorCode:     flag.ErrorCodeFlagNotFound,
			},
			wantErr:     true,
			expectedLog: `user="random-key", flag="key-not-exist", value="123456", variation="SdkDefault"`,
		},
		{
			name: "Get default value, rule not apply",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  map[string]any{"test123": "test"},
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("key eq \"key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface(map[string]any{"test": "test"}),
						"True":    testconvert.Interface(map[string]any{"test2": "test"}),
						"False":   testconvert.Interface(map[string]any{"test3": "test"}),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want: model.RawVarResult{
				Value:         map[string]any{"test": "test"},
				VariationType: "Default",
				Failed:        false,
				TrackEvents:   true,
				Reason:        flag.ReasonDefault,
				Cacheable:     true,
			},
			wantErr:     false,
			expectedLog: `user="random-key", flag="test-flag", value="map[test:test]", variation="Default"`,
		},
		{
			name: "Get true value, rule apply",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewAnonymousEvaluationContext("random-key"),
				defaultValue:  map[string]any{"test123": "test"},
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("key eq \"random-key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface(map[string]any{"test": "test"}),
						"True":    testconvert.Interface(map[string]any{"test2": "test"}),
						"False":   testconvert.Interface(map[string]any{"test3": "test"}),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want: model.RawVarResult{
				Value:         map[string]any{"test2": "test"},
				VariationType: "True",
				Failed:        false,
				TrackEvents:   true,
				Reason:        flag.ReasonTargetingMatch,
				Cacheable:     true,
				Metadata: map[string]any{
					"evaluatedRuleName": "rule1",
				},
			},
			wantErr:     false,
			expectedLog: `user="random-key", flag="test-flag", value="map[test2:test]", variation="True"`,
		},
		{
			name: "Get false value, rule apply",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewAnonymousEvaluationContext("random-key-ssss1"),
				defaultValue:  map[string]any{"test123": "test"},
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("anonymous eq true"),
							Percentages: &map[string]float64{
								"False": 90,
								"True":  10,
							},
						},
					},
					Variations: &map[string]*any{
						"Default": testconvert.Interface(map[string]any{"test": "test"}),
						"True":    testconvert.Interface(map[string]any{"test2": "test"}),
						"False":   testconvert.Interface(map[string]any{"test3": "test"}),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want: model.RawVarResult{
				Value:         map[string]any{"test3": "test"},
				VariationType: "False",
				Failed:        false,
				TrackEvents:   true,
				Reason:        flag.ReasonTargetingMatchSplit,
				Cacheable:     true,
				Metadata: map[string]any{
					"evaluatedRuleName": "rule1",
				},
			},
			wantErr:     false,
			expectedLog: `user="random-key-ssss1", flag="test-flag", value="map[test3:test]", variation="False"`,
		},
		{
			name: "No exported log",
			args: args{
				flagKey:       "test-flag",
				evaluationCtx: ffcontext.NewAnonymousEvaluationContext("random-key"),
				defaultValue:  true,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Variations: &map[string]*any{
						"Default": testconvert.Interface(false),
						"True":    testconvert.Interface(true),
						"False":   testconvert.Interface(false),
					},
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("rule1"),
							Query: testconvert.String("key eq \"random-key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("defaultRule"),
						VariationResult: testconvert.String("Default"),
					},
					TrackEvents: testconvert.Bool(false),
				}, nil),
			},
			want: model.RawVarResult{
				Value:         true,
				VariationType: "True",
				Failed:        false,
				TrackEvents:   false,
				Reason:        flag.ReasonTargetingMatch,
				Cacheable:     true,
				Metadata: map[string]any{
					"evaluatedRuleName": "rule1",
				},
			},
			wantErr:     false,
			expectedLog: "",
		},
		{
			name: "Get sdk default value if offline",
			args: args{
				offline:       true,
				flagKey:       "disable-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  false,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Disable: testconvert.Bool(true),
				}, nil),
			},
			want: model.RawVarResult{
				Value:         false,
				VariationType: flag.VariationSDKDefault,
				Failed:        false,
				TrackEvents:   false,
				Reason:        flag.ReasonOffline,
			},
			wantErr:     false,
			expectedLog: "",
		},
		{
			name: "should use interface default value if the flag is disabled",
			args: args{
				flagKey:       "disable-flag",
				evaluationCtx: ffcontext.NewEvaluationContext("random-key"),
				defaultValue:  nil,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Disable: testconvert.Bool(true),
				}, nil),
			},
			want: model.RawVarResult{
				TrackEvents:   true,
				VariationType: flag.VariationSDKDefault,
				Failed:        false,
				Reason:        flag.ReasonDisabled,
				Value:         nil,
				Cacheable:     true,
			},
			wantErr:     false,
			expectedLog: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := slogassert.New(t, slog.LevelInfo, nil)
			logger := slog.New(handler)

			if !tt.args.disableInit {
				c := retriever.ManagerConfig{
					FileFormat:              "YAML",
					DisableNotifierOnInit:   false,
					StartWithRetrieverError: false,
					EnablePollingJitter:     true,
					PollingInterval:         500,
				}
				ff = &GoFeatureFlag{
					retrieverManager: retriever.NewManager(c, []retriever.Retriever{}, tt.args.cacheMock, &fflog.FFLogger{LeveledLogger: logger}),
					config: Config{
						PollingInterval: 0,
						LeveledLogger:   logger,
						Offline:         tt.args.offline,
					},
					featureEventDataExporter: exporter.NewManager[exporter.FeatureEvent](
						[]exporter.Config{exporter.Config{
							FlushInterval:    0,
							MaxEventInMemory: 0,
							Exporter: &logsexporter.Exporter{
								LogFormat: "user=\"{{ .UserKey}}\", flag=\"{{ .Key}}\", " +
									"value=\"{{ .Value}}\", variation=\"{{ .Variation}}\"",
							},
						}},
						exporter.DefaultExporterCleanQueueInterval,
						&fflog.FFLogger{LeveledLogger: logger},
					),
				}
			}

			got, err := ff.RawVariation(tt.args.flagKey, tt.args.evaluationCtx, tt.args.defaultValue)

			if tt.expectedLog != "" {
				time.Sleep(
					40 * time.Millisecond,
				) // since the log is async, we are waiting to be sure it's written
				if tt.expectedLog == "" {
					handler.AssertEmpty()
				} else {
					handler.Assert(func(message slogassert.LogMessage) bool {
						if !strings.Contains(message.Message, tt.expectedLog) {
							handler.Fail("impossible to find %s in %s", tt.expectedLog, message.Message)
							return false
						}
						return true
					})
				}
			}

			if tt.wantErr {
				assert.Error(t, err, "RawVariation() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.want, got, "RawVariation() got = %v, want %v", got, tt.want)

			// clean logger
			ff = nil
		})
	}
}

func Test_OverrideContextEnrichmentWithEnvironment(t *testing.T) {
	tempFile, err := os.CreateTemp("", "")
	require.NoError(t, err)
	defer tempFile.Close()

	err = os.WriteFile(tempFile.Name(), []byte(`
flag1:
 variations:
   enabled: true
   disabled: false
 targeting:
   - query: env eq "staging"
     variation: enabled
 defaultRule:
   variation: disabled

`), 0644)
	require.NoError(t, err)

	goff, err := New(Config{
		PollingInterval: 500 * time.Millisecond,
		Retriever:       &fileretriever.Retriever{Path: tempFile.Name()},
		EvaluationContextEnrichment: map[string]any{
			"env": "staging",
		},
	})
	require.NoError(t, err)

	res, err1 := goff.BoolVariation(
		"flag1",
		ffcontext.NewEvaluationContextBuilder("my-key").Build(),
		false,
	)
	assert.True(t, res)
	assert.NoError(t, err1)
	allFlags := goff.AllFlagsState(ffcontext.NewEvaluationContextBuilder("my-key").Build())
	assert.Equal(t, true, allFlags.GetFlags()["flag1"].Value)

	goff2, err2 := New(Config{
		PollingInterval: 500 * time.Millisecond,
		Retriever:       &fileretriever.Retriever{Path: tempFile.Name()},
		Environment:     "staging",
		EvaluationContextEnrichment: map[string]any{
			"env": "staging",
		},
	})
	require.NoError(t, err2)
	res2, err3 := goff2.BoolVariation(
		"flag1",
		ffcontext.NewEvaluationContextBuilder("my-key").Build(),
		false,
	)
	assert.True(t, res2)
	assert.NoError(t, err3)
	allFlags2 := goff2.AllFlagsState(ffcontext.NewEvaluationContextBuilder("my-key").Build())
	assert.Equal(t, true, allFlags2.GetFlags()["flag1"].Value)

	// Explicit environment should override the environment from the enrichment
	goff3, err4 := New(Config{
		PollingInterval: 500 * time.Millisecond,
		Retriever:       &fileretriever.Retriever{Path: tempFile.Name()},
		Environment:     "staging",
		EvaluationContextEnrichment: map[string]any{
			"env": "prod",
		},
	})
	require.NoError(t, err4)
	res3, err5 := goff3.BoolVariation(
		"flag1",
		ffcontext.NewEvaluationContextBuilder("my-key").Build(),
		false,
	)
	assert.True(t, res3)
	assert.NoError(t, err5)

	allFlags3 := goff3.AllFlagsState(ffcontext.NewEvaluationContextBuilder("my-key").Build())
	assert.Equal(t, true, allFlags3.GetFlags()["flag1"].Value)
}

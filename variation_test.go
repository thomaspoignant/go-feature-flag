package ffclient

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"

	"github.com/thomaspoignant/go-feature-flag/internal/model"

	"github.com/thomaspoignant/go-feature-flag/exporter/fileexporter"

	"github.com/thomaspoignant/go-feature-flag/exporter/logsexporter"
	"github.com/thomaspoignant/go-feature-flag/internal/dataexporter"

	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/ffuser"
	"github.com/thomaspoignant/go-feature-flag/internal/cache"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	"github.com/thomaspoignant/go-feature-flag/testutils"
	"github.com/thomaspoignant/go-feature-flag/testutils/testconvert"
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

func (c *cacheMock) UpdateCache(loadedFlags []byte, fileFormat string, log *log.Logger) error {
	return nil
}
func (c *cacheMock) Close() {}
func (c *cacheMock) GetFlag(key string) (flag.Flag, error) {
	return c.flag, c.err
}
func (c *cacheMock) AllFlags() (map[string]flag.Flag, error) { return nil, nil }

func TestBoolVariation(t *testing.T) {
	type args struct {
		flagKey      string
		user         ffuser.User
		defaultValue bool
		cacheMock    cache.Manager
		offline      bool
	}
	tests := []struct {
		name        string
		args        args
		want        bool
		wantErr     bool
		expectedLog string
	}{
		{
			name: "Get default value if flag disable",
			args: args{
				flagKey:      "disable-flag",
				user:         ffuser.NewUser("random-key"),
				defaultValue: true,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Disable: testconvert.Bool(true),
				}, nil),
			},
			want:        true,
			wantErr:     false,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key\", flag=\"disable-flag\", value=\"true\", variation=\"SdkDefault\"\n",
		},
		{
			name: "Get error when not init",
			args: args{
				flagKey:      "key-not-exist",
				user:         ffuser.NewUser("random-key"),
				defaultValue: true,
				cacheMock: NewCacheMock(
					&flag.InternalFlag{},
					errors.New("impossible to read the toggle before the initialisation")),
			},
			want:        true,
			wantErr:     true,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key\", flag=\"key-not-exist\", value=\"true\", variation=\"SdkDefault\"\n",
		},
		{
			name: "Get default value with key not exist",
			args: args{
				flagKey:      "key-not-exist",
				user:         ffuser.NewUser("random-key"),
				defaultValue: true,
				cacheMock:    NewCacheMock(&flag.InternalFlag{}, errors.New("flag [key-not-exist] does not exists")),
			},
			want:        true,
			wantErr:     true,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key\", flag=\"key-not-exist\", value=\"true\", variation=\"SdkDefault\"\n",
		},
		{
			name: "Get default value, rule not apply",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewUser("random-key"),
				defaultValue: true,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("legacyRuleV0"),
							Query: testconvert.String("key eq \"key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*interface{}{
						"Default": testconvert.Interface(true),
						"False":   testconvert.Interface(false),
						"True":    testconvert.Interface(true),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("legacyDefaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want:        true,
			wantErr:     false,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key\", flag=\"test-flag\", value=\"true\", variation=\"Default\"\n",
		},
		{
			name: "Get true value, rule apply",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewAnonymousUser("random-key"),
				defaultValue: true,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("legacyRuleV0"),
							Query: testconvert.String("key eq \"random-key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*interface{}{
						"Default": testconvert.Interface(false),
						"False":   testconvert.Interface(false),
						"True":    testconvert.Interface(true),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("legacyDefaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want:        true,
			wantErr:     false,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key\", flag=\"test-flag\", value=\"true\", variation=\"True\"\n",
		},
		{
			name: "Get false value, rule apply",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewAnonymousUser("random-key-ssss1"),
				defaultValue: true,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("legacyRuleV0"),
							Query: testconvert.String("anonymous eq true"),
							Percentages: &map[string]float64{
								"False": 90,
								"True":  10,
							},
						},
					},
					Variations: &map[string]*interface{}{
						"Default": testconvert.Interface(false),
						"False":   testconvert.Interface(false),
						"True":    testconvert.Interface(true),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("legacyDefaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want:        false,
			wantErr:     false,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key-ssss1\", flag=\"test-flag\", value=\"false\", variation=\"False\"\n",
		},
		{
			name: "Get default value, when rule apply and not right type",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewUser("random-key-ssss1"),
				defaultValue: true,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("legacyRuleV0"),
							Query: testconvert.String("key eq \"random-key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*interface{}{
						"Default": testconvert.Interface("xxx"),
						"False":   testconvert.Interface("xxx"),
						"True":    testconvert.Interface("xxx"),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("legacyDefaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want:        true,
			wantErr:     true,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key-ssss1\", flag=\"test-flag\", value=\"true\", variation=\"SdkDefault\"\n",
		},
		{
			name: "No exported log",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewAnonymousUser("random-key"),
				defaultValue: true,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("legacyRuleV0"),
							Query: testconvert.String("key eq \"random-key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*interface{}{
						"Default": testconvert.Interface(false),
						"False":   testconvert.Interface(false),
						"True":    testconvert.Interface(true),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("legacyDefaultRule"),
						VariationResult: testconvert.String("Default"),
					},
					TrackEvents: testconvert.Bool(false),
				}, nil),
			},
			want:        true,
			wantErr:     false,
			expectedLog: "^$",
		},
		{
			name: "Get sdk default value if offline",
			args: args{
				offline:      true,
				flagKey:      "disable-flag",
				user:         ffuser.NewUser("random-key"),
				defaultValue: false,
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
			// init logger
			file, _ := ioutil.TempFile("", "log")
			logger := log.New(file, "", 0)

			ff = &GoFeatureFlag{
				bgUpdater: newBackgroundUpdater(5),
				cache:     tt.args.cacheMock,
				config: Config{
					PollingInterval: 0,
					Logger:          logger,
					Offline:         tt.args.offline,
				},
				dataExporter: dataexporter.NewScheduler(context.Background(), 0, 0,
					&logsexporter.Exporter{
						LogFormat: "[{{ .FormattedDate}}] user=\"{{ .UserKey}}\", flag=\"{{ .Key}}\", " +
							"value=\"{{ .Value}}\", variation=\"{{ .Variation}}\"",
					}, logger),
			}

			got, err := BoolVariation(tt.args.flagKey, tt.args.user, tt.args.defaultValue)

			if tt.expectedLog != "" {
				content, _ := ioutil.ReadFile(file.Name())
				assert.Regexp(t, tt.expectedLog, string(content))
			}

			if tt.wantErr {
				assert.Error(t, err, "BoolVariation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got, "BoolVariation() got = %v, want %v", got, tt.want)

			// clean logger
			ff = nil
			_ = file.Close()
		})
	}
}

func TestFloat64Variation(t *testing.T) {
	type args struct {
		flagKey      string
		user         ffuser.User
		defaultValue float64
		cacheMock    cache.Manager
		offline      bool
	}
	tests := []struct {
		name        string
		args        args
		want        float64
		wantErr     bool
		expectedLog string
	}{
		{
			name: "Get default value if flag disable",
			args: args{
				flagKey:      "disable-flag",
				user:         ffuser.NewUser("random-key"),
				defaultValue: 120.12,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Disable: testconvert.Bool(true),
				}, nil),
			},
			want:        120.12,
			wantErr:     false,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key\", flag=\"disable-flag\", value=\"120.12\", variation=\"SdkDefault\"\n",
		},
		{
			name: "Get error when not init",
			args: args{
				flagKey:      "key-not-exist",
				user:         ffuser.NewUser("random-key"),
				defaultValue: 118.12,
				cacheMock: NewCacheMock(
					&flag.InternalFlag{},
					errors.New("impossible to read the toggle before the initialisation")),
			},
			want:        118.12,
			wantErr:     true,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key\", flag=\"key-not-exist\", value=\"118.12\", variation=\"SdkDefault\"\n",
		},
		{
			name: "Get default value with key not exist",
			args: args{
				flagKey:      "key-not-exist",
				user:         ffuser.NewUser("random-key"),
				defaultValue: 118.12,
				cacheMock:    NewCacheMock(&flag.InternalFlag{}, errors.New("flag [key-not-exist] does not exists")),
			},
			want:        118.12,
			wantErr:     true,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key\", flag=\"key-not-exist\", value=\"118.12\", variation=\"SdkDefault\"\n",
		},
		{
			name: "Get default value, rule not apply",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewUser("random-key"),
				defaultValue: 118.12,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("legacyRuleV0"),
							Query: testconvert.String("key eq \"key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*interface{}{
						"Default": testconvert.Interface(119.12),
						"False":   testconvert.Interface(121.12),
						"True":    testconvert.Interface(120.12),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("legacyDefaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want:        119.12,
			wantErr:     false,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key\", flag=\"test-flag\", value=\"119.12\", variation=\"Default\"\n",
		},
		{
			name: "Get true value, rule apply",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewAnonymousUser("random-key"),
				defaultValue: 118.12,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("legacyRuleV0"),
							Query: testconvert.String("key eq \"random-key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*interface{}{
						"Default": testconvert.Interface(119.12),
						"False":   testconvert.Interface(121.12),
						"True":    testconvert.Interface(120.12),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("legacyDefaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want:        120.12,
			wantErr:     false,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key\", flag=\"test-flag\", value=\"120.12\", variation=\"True\"\n",
		},
		{
			name: "Get false value, rule apply",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewAnonymousUser("random-key-ssss1"),
				defaultValue: 118.12,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("legacyRuleV0"),
							Query: testconvert.String("anonymous eq true"),
							Percentages: &map[string]float64{
								"False": 90,
								"True":  10,
							},
						},
					},
					Variations: &map[string]*interface{}{
						"Default": testconvert.Interface(119.12),
						"False":   testconvert.Interface(121.12),
						"True":    testconvert.Interface(120.12),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("legacyDefaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want:        121.12,
			wantErr:     false,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key-ssss1\", flag=\"test-flag\", value=\"121.12\", variation=\"False\"\n",
		},
		{
			name: "Get default value, when rule apply and not right type",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewUser("random-key-ssss1"),
				defaultValue: 118.12,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("legacyRuleV0"),
							Query: testconvert.String("key eq \"random-key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*interface{}{
						"Default": testconvert.Interface("xxx"),
						"False":   testconvert.Interface("xxx"),
						"True":    testconvert.Interface("xxx"),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("legacyDefaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want:        118.12,
			wantErr:     true,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key-ssss1\", flag=\"test-flag\", value=\"118.12\", variation=\"SdkDefault\"\n",
		},
		{
			name: "No exported log",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewAnonymousUser("random-key"),
				defaultValue: 118.12,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("legacyRuleV0"),
							Query: testconvert.String("key eq \"random-key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*interface{}{
						"Default": testconvert.Interface(119.12),
						"False":   testconvert.Interface(121.12),
						"True":    testconvert.Interface(120.12),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("legacyDefaultRule"),
						VariationResult: testconvert.String("Default"),
					},
					TrackEvents: testconvert.Bool(false),
				}, nil),
			},
			want:        120.12,
			wantErr:     false,
			expectedLog: "^$",
		},
		{
			name: "Get sdk default value if offline",
			args: args{
				offline:      true,
				flagKey:      "disable-flag",
				user:         ffuser.NewUser("random-key"),
				defaultValue: 118.12,
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
			// init logger
			file, _ := ioutil.TempFile("", "log")
			logger := log.New(file, "", 0)

			ff = &GoFeatureFlag{
				bgUpdater: newBackgroundUpdater(5),
				cache:     tt.args.cacheMock,
				config: Config{
					PollingInterval: 0,
					Logger:          logger,
					Offline:         tt.args.offline,
				},
				dataExporter: dataexporter.NewScheduler(context.Background(), 0, 0,
					&logsexporter.Exporter{
						LogFormat: "[{{ .FormattedDate}}] user=\"{{ .UserKey}}\", flag=\"{{ .Key}}\", " +
							"value=\"{{ .Value}}\", variation=\"{{ .Variation}}\"",
					}, logger),
			}

			got, err := Float64Variation(tt.args.flagKey, tt.args.user, tt.args.defaultValue)

			if tt.expectedLog != "" {
				content, _ := ioutil.ReadFile(file.Name())
				assert.Regexp(t, tt.expectedLog, string(content))
			}
			if tt.wantErr {
				assert.Error(t, err, "Float64Variation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got, "Float64Variation() got = %v, want %v", got, tt.want)

			// clean logger
			ff = nil
			_ = file.Close()
		})
	}
}

func TestJSONArrayVariation(t *testing.T) {
	type args struct {
		flagKey      string
		user         ffuser.User
		defaultValue []interface{}
		cacheMock    cache.Manager
		offline      bool
	}
	tests := []struct {
		name        string
		args        args
		want        []interface{}
		wantErr     bool
		expectedLog string
	}{
		{
			name: "Get default value if flag disable",
			args: args{
				flagKey:      "disable-flag",
				user:         ffuser.NewUser("random-key"),
				defaultValue: []interface{}{"toto"},
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Disable: testconvert.Bool(true),
				}, nil),
			},
			want:        []interface{}{"toto"},
			wantErr:     false,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key\", flag=\"disable-flag\", value=\"\\[toto\\]\"\n",
		},
		{
			name: "Get error when not init",
			args: args{
				flagKey:      "key-not-exist",
				user:         ffuser.NewUser("random-key"),
				defaultValue: []interface{}{"toto"},
				cacheMock: NewCacheMock(
					&flag.InternalFlag{},
					errors.New("impossible to read the toggle before the initialisation")),
			},
			want:        []interface{}{"toto"},
			wantErr:     true,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key\", flag=\"key-not-exist\", value=\"\\[toto\\]\"\n",
		},
		{
			name: "Get default value with key not exist",
			args: args{
				flagKey:      "key-not-exist",
				user:         ffuser.NewUser("random-key"),
				defaultValue: []interface{}{"toto"},
				cacheMock:    NewCacheMock(&flag.InternalFlag{}, errors.New("flag [key-not-exist] does not exists")),
			},
			want:        []interface{}{"toto"},
			wantErr:     true,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key\", flag=\"key-not-exist\", value=\"\\[toto\\]\"\n",
		},
		{
			name: "Get default value, rule not apply",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewUser("random-key"),
				defaultValue: []interface{}{"toto"},
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("legacyRuleV0"),
							Query: testconvert.String("key eq \"key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*interface{}{
						"Default": testconvert.Interface([]interface{}{"default"}),
						"True":    testconvert.Interface([]interface{}{"true"}),
						"False":   testconvert.Interface([]interface{}{"false"}),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("legacyDefaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want:        []interface{}{"default"},
			wantErr:     false,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key\", flag=\"test-flag\", value=\"\\[default\\]\"\n",
		},
		{
			name: "Get true value, rule apply",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewAnonymousUser("random-key"),
				defaultValue: []interface{}{"toto"},
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("legacyRuleV0"),
							Query: testconvert.String("key eq \"random-key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*interface{}{
						"Default": testconvert.Interface([]interface{}{"default"}),
						"True":    testconvert.Interface([]interface{}{"true"}),
						"False":   testconvert.Interface([]interface{}{"false"}),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("legacyDefaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want:        []interface{}{"true"},
			wantErr:     false,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key\", flag=\"test-flag\", value=\"\\[true\\]\"\n",
		},
		{
			name: "Get false value, rule apply",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewAnonymousUser("random-key-ssss1"),
				defaultValue: []interface{}{"toto"},
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Variations: &map[string]*interface{}{
						"Default": testconvert.Interface([]interface{}{"default"}),
						"True":    testconvert.Interface([]interface{}{"true"}),
						"False":   testconvert.Interface([]interface{}{"false"}),
					},
					DefaultRule: &flag.Rule{
						Name: testconvert.String("legacyDefaultRule"),
						Percentages: &map[string]float64{
							"False": 90,
							"True":  10,
						},
					},
				}, nil),
			},
			want:        []interface{}{"false"},
			wantErr:     false,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key-ssss1\", flag=\"test-flag\", value=\"\\[false\\]\"\n",
		},
		{
			name: "Get default value, when rule apply and not right type",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewUser("random-key-ssss1"),
				defaultValue: []interface{}{"toto"},
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("legacyRuleV0"),
							Query: testconvert.String("key eq \"key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*interface{}{
						"Default": testconvert.Interface("xxx"),
						"False":   testconvert.Interface("xxx"),
						"True":    testconvert.Interface("xxx"),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("legacyDefaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want:        []interface{}{"toto"},
			wantErr:     true,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key-ssss1\", flag=\"test-flag\", value=\"\\[toto\\]\"\n",
		},
		{
			name: "No exported log",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewUser("random-key-ssss1"),
				defaultValue: []interface{}{"toto"},
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Variations: &map[string]*interface{}{
						"Default": testconvert.Interface([]interface{}{"default"}),
						"True":    testconvert.Interface([]interface{}{"true"}),
						"False":   testconvert.Interface([]interface{}{"false"}),
					},
					DefaultRule: &flag.Rule{
						Name: testconvert.String("legacyDefaultRule"),
						Percentages: &map[string]float64{
							"False": 0,
							"True":  100,
						},
					},
					TrackEvents: testconvert.Bool(false),
				}, nil),
			},
			want:        []interface{}{"true"},
			wantErr:     false,
			expectedLog: "^$",
		},
		{
			name: "Get sdk default value if offline",
			args: args{
				offline:      true,
				flagKey:      "disable-flag",
				user:         ffuser.NewUser("random-key"),
				defaultValue: []interface{}{"toto"},
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Disable: testconvert.Bool(true),
				}, nil),
			},
			want:        []interface{}{"toto"},
			wantErr:     false,
			expectedLog: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// init logger
			file, _ := ioutil.TempFile("", "log")
			logger := log.New(file, "", 0)

			ff = &GoFeatureFlag{
				bgUpdater: newBackgroundUpdater(5),
				cache:     tt.args.cacheMock,
				config: Config{
					PollingInterval: 0,
					Logger:          logger,
					Offline:         tt.args.offline,
				},
				dataExporter: dataexporter.NewScheduler(context.Background(), 0, 0,
					&logsexporter.Exporter{}, logger),
			}

			got, err := JSONArrayVariation(tt.args.flagKey, tt.args.user, tt.args.defaultValue)

			if tt.wantErr {
				assert.Error(t, err, "JSONArrayVariation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got, "JSONArrayVariation() got = %v, want %v", got, tt.want)
			if tt.expectedLog != "" {
				content, _ := ioutil.ReadFile(file.Name())
				assert.Regexp(t, tt.expectedLog, string(content))
			}
			// clean logger
			ff = nil
			_ = file.Close()
		})
	}
}

func TestJSONVariation(t *testing.T) {
	type args struct {
		flagKey      string
		user         ffuser.User
		defaultValue map[string]interface{}
		cacheMock    cache.Manager
		offline      bool
	}
	tests := []struct {
		name        string
		args        args
		want        map[string]interface{}
		wantErr     bool
		expectedLog string
	}{
		{
			name: "Get default value if flag disable",
			args: args{
				flagKey:      "disable-flag",
				user:         ffuser.NewUser("random-key"),
				defaultValue: map[string]interface{}{"default-notkey": true},
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Disable: testconvert.Bool(true),
				}, nil),
			},
			want:        map[string]interface{}{"default-notkey": true},
			wantErr:     false,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key\", flag=\"disable-flag\", value=\"map\\[default-notkey:true\\]\", variation=\"SdkDefault\"\n",
		},
		{
			name: "Get error when not init",
			args: args{
				flagKey:      "key-not-exist",
				user:         ffuser.NewUser("random-key"),
				defaultValue: map[string]interface{}{"default-notkey": true},
				cacheMock: NewCacheMock(
					&flag.InternalFlag{},
					errors.New("impossible to read the toggle before the initialisation")),
			},
			want:        map[string]interface{}{"default-notkey": true},
			wantErr:     true,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key\", flag=\"key-not-exist\", value=\"map\\[default-notkey:true\\]\", variation=\"SdkDefault\"\n",
		},
		{
			name: "Get default value with key not exist",
			args: args{
				flagKey:      "key-not-exist",
				user:         ffuser.NewUser("random-key"),
				defaultValue: map[string]interface{}{"default-notkey": true},
				cacheMock:    NewCacheMock(&flag.InternalFlag{}, errors.New("flag [key-not-exist] does not exists")),
			},
			want:        map[string]interface{}{"default-notkey": true},
			wantErr:     true,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key\", flag=\"key-not-exist\", value=\"map\\[default-notkey:true\\]\", variation=\"SdkDefault\"\n",
		},
		{
			name: "Get default value, rule not apply",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewUser("random-key"),
				defaultValue: map[string]interface{}{"default-notkey": true},
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("legacyRuleV0"),
							Query: testconvert.String("key eq \"key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*interface{}{
						"Default": testconvert.Interface(map[string]interface{}{"default": true}),
						"True":    testconvert.Interface(map[string]interface{}{"true": true}),
						"False":   testconvert.Interface(map[string]interface{}{"false": true}),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("legacyDefaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want:        map[string]interface{}{"default": true},
			wantErr:     false,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key\", flag=\"test-flag\", value=\"map\\[default:true\\]\", variation=\"Default\"\n",
		},
		{
			name: "Get true value, rule apply",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewAnonymousUser("random-key"),
				defaultValue: map[string]interface{}{"default-notkey": true},
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("legacyRuleV0"),
							Query: testconvert.String("key eq \"random-key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*interface{}{
						"Default": testconvert.Interface(map[string]interface{}{"default": true}),
						"True":    testconvert.Interface(map[string]interface{}{"true": true}),
						"False":   testconvert.Interface(map[string]interface{}{"false": true}),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("legacyDefaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want:        map[string]interface{}{"true": true},
			wantErr:     false,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key\", flag=\"test-flag\", value=\"map\\[true:true\\]\", variation=\"True\"\n",
		},
		{
			name: "Get false value, rule apply",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewAnonymousUser("random-key-ssss1"),
				defaultValue: map[string]interface{}{"default-notkey": true},
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("legacyRuleV0"),
							Query: testconvert.String("anonymous eq true"),
							Percentages: &map[string]float64{
								"False": 90,
								"True":  10,
							},
						},
					},
					Variations: &map[string]*interface{}{
						"Default": testconvert.Interface(map[string]interface{}{"default": true}),
						"True":    testconvert.Interface(map[string]interface{}{"true": true}),
						"False":   testconvert.Interface(map[string]interface{}{"false": true}),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("legacyDefaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want:        map[string]interface{}{"false": true},
			wantErr:     false,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key-ssss1\", flag=\"test-flag\", value=\"map\\[false:true\\]\", variation=\"False\"\n",
		},
		{
			name: "Get default value, when rule apply and not right type",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewUser("random-key-ssss1"),
				defaultValue: map[string]interface{}{"default-notkey": true},
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("legacyRuleV0"),
							Query: testconvert.String("key eq \"key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*interface{}{
						"Default": testconvert.Interface("xxx"),
						"False":   testconvert.Interface("xxx"),
						"True":    testconvert.Interface("xxx"),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("legacyDefaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want:        map[string]interface{}{"default-notkey": true},
			wantErr:     true,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key-ssss1\", flag=\"test-flag\", value=\"map\\[default-notkey:true\\]\", variation=\"SdkDefault\"\n",
		},
		{
			name: "Get sdk default value if offline",
			args: args{
				offline:      true,
				flagKey:      "disable-flag",
				user:         ffuser.NewUser("random-key"),
				defaultValue: map[string]interface{}{"default-notkey": true},
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Disable: testconvert.Bool(true),
				}, nil),
			},
			want:        map[string]interface{}{"default-notkey": true},
			wantErr:     false,
			expectedLog: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// init logger
			file, _ := ioutil.TempFile("", "log")
			logger := log.New(file, "", 0)

			ff = &GoFeatureFlag{
				bgUpdater: newBackgroundUpdater(5),
				cache:     tt.args.cacheMock,
				config: Config{
					PollingInterval: 0,
					Logger:          logger,
					Offline:         tt.args.offline,
				},
				dataExporter: dataexporter.NewScheduler(context.Background(), 0, 0,
					&logsexporter.Exporter{
						LogFormat: "[{{ .FormattedDate}}] user=\"{{ .UserKey}}\", flag=\"{{ .Key}}\", " +
							"value=\"{{ .Value}}\", variation=\"{{ .Variation}}\"",
					}, logger),
			}

			got, err := JSONVariation(tt.args.flagKey, tt.args.user, tt.args.defaultValue)

			if tt.expectedLog != "" {
				content, _ := ioutil.ReadFile(file.Name())
				assert.Regexp(t, tt.expectedLog, string(content))
			}

			if tt.wantErr {
				assert.Error(t, err, "JSONVariation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got, "JSONVariation() got = %v, want %v", got, tt.want)

			// clean logger
			ff = nil
			_ = file.Close()
		})
	}
}

func TestStringVariation(t *testing.T) {
	type args struct {
		flagKey      string
		user         ffuser.User
		defaultValue string
		cacheMock    cache.Manager
		offline      bool
	}
	tests := []struct {
		name        string
		args        args
		want        string
		wantErr     bool
		expectedLog string
	}{
		{
			name: "Get default value if flag disable",
			args: args{
				flagKey:      "disable-flag",
				user:         ffuser.NewUser("random-key"),
				defaultValue: "default-notkey",
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Disable: testconvert.Bool(true),
				}, nil),
			},
			want:        "default-notkey",
			wantErr:     false,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key\", flag=\"disable-flag\", value=\"default-notkey\", variation=\"SdkDefault\"\n",
		},
		{
			name: "Get error when not init",
			args: args{
				flagKey:      "key-not-exist",
				user:         ffuser.NewUser("random-key"),
				defaultValue: "default-notkey",
				cacheMock: NewCacheMock(
					&flag.InternalFlag{},
					errors.New("impossible to read the toggle before the initialisation")),
			},
			want:        "default-notkey",
			wantErr:     true,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key\", flag=\"key-not-exist\", value=\"default-notkey\", variation=\"SdkDefault\"\n",
		},
		{
			name: "Get default value with key not exist",
			args: args{
				flagKey:      "key-not-exist",
				user:         ffuser.NewUser("random-key"),
				defaultValue: "default-notkey",
				cacheMock:    NewCacheMock(&flag.InternalFlag{}, errors.New("flag [key-not-exist] does not exists")),
			},
			want:        "default-notkey",
			wantErr:     true,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key\", flag=\"key-not-exist\", value=\"default-notkey\", variation=\"SdkDefault\"\n",
		},

		{
			name: "Get default value, rule not apply",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewUser("random-key"),
				defaultValue: "default-notkey",
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("legacyRuleV0"),
							Query: testconvert.String("key eq \"key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*interface{}{
						"Default": testconvert.Interface("default"),
						"True":    testconvert.Interface("true"),
						"False":   testconvert.Interface("false"),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("legacyDefaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want:        "default",
			wantErr:     false,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key\", flag=\"test-flag\", value=\"default\", variation=\"Default\"\n",
		},
		{
			name: "Get true value, rule apply",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewAnonymousUser("random-key"),
				defaultValue: "default-notkey",
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("legacyRuleV0"),
							Query: testconvert.String("key eq \"random-key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*interface{}{
						"Default": testconvert.Interface("default"),
						"True":    testconvert.Interface("true"),
						"False":   testconvert.Interface("false"),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("legacyDefaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want:        "true",
			wantErr:     false,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key\", flag=\"test-flag\", value=\"true\", variation=\"True\"\n",
		},
		{
			name: "Get false value, rule apply",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewAnonymousUser("random-key-ssss1"),
				defaultValue: "default-notkey",
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("legacyRuleV0"),
							Query: testconvert.String("anonymous eq true"),
							Percentages: &map[string]float64{
								"False": 90,
								"True":  10,
							},
						},
					},
					Variations: &map[string]*interface{}{
						"Default": testconvert.Interface("default"),
						"True":    testconvert.Interface("true"),
						"False":   testconvert.Interface("false"),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("legacyDefaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want:        "false",
			wantErr:     false,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key-ssss1\", flag=\"test-flag\", value=\"false\", variation=\"False\"\n",
		},
		{
			name: "Get default value, when rule apply and not right type",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewUser("random-key-ssss1"),
				defaultValue: "default-notkey",
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("legacyRuleV0"),
							Query: testconvert.String("key eq \"key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*interface{}{
						"Default": testconvert.Interface(1),
						"False":   testconvert.Interface(2),
						"True":    testconvert.Interface(3),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("legacyDefaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want:        "default-notkey",
			wantErr:     true,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key-ssss1\", flag=\"test-flag\", value=\"default-notkey\", variation=\"SdkDefault\"\n",
		},
		{
			name: "Get sdk default value if offline",
			args: args{
				offline:      true,
				flagKey:      "disable-flag",
				user:         ffuser.NewUser("random-key"),
				defaultValue: "default-notkey",
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Disable: testconvert.Bool(true),
				}, nil),
			},
			want:        "default-notkey",
			wantErr:     false,
			expectedLog: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// init logger
			file, _ := ioutil.TempFile("", "log")
			logger := log.New(file, "", 0)

			ff = &GoFeatureFlag{
				bgUpdater: newBackgroundUpdater(5),
				cache:     tt.args.cacheMock,
				config: Config{
					PollingInterval: 0,
					Logger:          logger,
					Offline:         tt.args.offline,
				},
				dataExporter: dataexporter.NewScheduler(context.Background(), 0, 0,
					&logsexporter.Exporter{
						LogFormat: "[{{ .FormattedDate}}] user=\"{{ .UserKey}}\", flag=\"{{ .Key}}\", " +
							"value=\"{{ .Value}}\", variation=\"{{ .Variation}}\"",
					}, logger),
			}
			got, err := StringVariation(tt.args.flagKey, tt.args.user, tt.args.defaultValue)

			if tt.expectedLog != "" {
				content, _ := ioutil.ReadFile(file.Name())
				assert.Regexp(t, tt.expectedLog, string(content))
			}

			if tt.wantErr {
				assert.Error(t, err, "StringVariation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got, "StringVariation() got = %v, want %v", got, tt.want)

			// clean logger
			ff = nil
			_ = file.Close()
		})
	}
}

func TestIntVariation(t *testing.T) {
	type args struct {
		flagKey      string
		user         ffuser.User
		defaultValue int
		cacheMock    cache.Manager
		offline      bool
	}
	tests := []struct {
		name        string
		args        args
		want        int
		wantErr     bool
		expectedLog string
	}{
		{
			name: "Get default value if flag disable",
			args: args{
				flagKey:      "disable-flag",
				user:         ffuser.NewUser("random-key"),
				defaultValue: 125,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Disable: testconvert.Bool(true),
				}, nil),
			},
			want:        125,
			wantErr:     false,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key\", flag=\"disable-flag\", value=\"125\", variation=\"SdkDefault\"\n",
		},
		{
			name: "Get error when not init",
			args: args{
				flagKey:      "key-not-exist",
				user:         ffuser.NewUser("random-key"),
				defaultValue: 118,
				cacheMock: NewCacheMock(
					&flag.InternalFlag{},
					errors.New("impossible to read the toggle before the initialisation")),
			},
			want:        118,
			wantErr:     true,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key\", flag=\"key-not-exist\", value=\"118\", variation=\"SdkDefault\"\n",
		},
		{
			name: "Get default value with key not exist",
			args: args{
				flagKey:      "key-not-exist",
				user:         ffuser.NewUser("random-key"),
				defaultValue: 118,
				cacheMock:    NewCacheMock(&flag.InternalFlag{}, errors.New("flag [key-not-exist] does not exists")),
			},
			want:        118,
			wantErr:     true,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key\", flag=\"key-not-exist\", value=\"118\", variation=\"SdkDefault\"\n",
		},
		{
			name: "Get default value rule not apply",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewUser("random-key"),
				defaultValue: 118,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("legacyRuleV0"),
							Query: testconvert.String("key eq \"key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*interface{}{
						"Default": testconvert.Interface(119),
						"True":    testconvert.Interface(120),
						"False":   testconvert.Interface(121),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("legacyDefaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want:        119,
			wantErr:     false,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key\", flag=\"test-flag\", value=\"119\", variation=\"Default\"\n",
		},
		{
			name: "Get true value, rule apply",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewAnonymousUser("random-key"),
				defaultValue: 118,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("legacyRuleV0"),
							Query: testconvert.String("key eq \"random-key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*interface{}{
						"Default": testconvert.Interface(119),
						"True":    testconvert.Interface(120),
						"False":   testconvert.Interface(121),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("legacyDefaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want:        120,
			wantErr:     false,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key\", flag=\"test-flag\", value=\"120\", variation=\"True\"\n",
		},
		{
			name: "Get false value, rule apply",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewAnonymousUser("random-key-ssss1"),
				defaultValue: 118,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("legacyRuleV0"),
							Query: testconvert.String("anonymous eq true"),
							Percentages: &map[string]float64{
								"False": 90,
								"True":  10,
							},
						},
					},
					Variations: &map[string]*interface{}{
						"Default": testconvert.Interface(119),
						"True":    testconvert.Interface(120),
						"False":   testconvert.Interface(121),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("legacyDefaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want:        121,
			wantErr:     false,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key-ssss1\", flag=\"test-flag\", value=\"121\", variation=\"False\"\n",
		},
		{
			name: "Get default value, when rule apply and not right type",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewUser("random-key-ssss1"),
				defaultValue: 118,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("legacyRuleV0"),
							Query: testconvert.String("key eq \"key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*interface{}{
						"Default": testconvert.Interface("xxx"),
						"False":   testconvert.Interface("xxx"),
						"True":    testconvert.Interface("xxx"),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("legacyDefaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want:        118,
			wantErr:     true,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key-ssss1\", flag=\"test-flag\", value=\"118\", variation=\"SdkDefault\"\n",
		},
		{
			name: "Convert float to Int",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewAnonymousUser("random-key"),
				defaultValue: 118,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("legacyRuleV0"),
							Query: testconvert.String("key eq \"random-key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*interface{}{
						"Default": testconvert.Interface(119.1),
						"True":    testconvert.Interface(120.1),
						"False":   testconvert.Interface(121.1),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("legacyDefaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want:        120,
			wantErr:     false,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key\", flag=\"test-flag\", value=\"120\", variation=\"True\"\n",
		},
		{
			name: "Get sdk default value if offline",
			args: args{
				offline:      true,
				flagKey:      "disable-flag",
				user:         ffuser.NewUser("random-key"),
				defaultValue: 125,
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
			// init logger
			file, _ := ioutil.TempFile("", "log")
			logger := log.New(file, "", 0)

			ff = &GoFeatureFlag{
				bgUpdater: newBackgroundUpdater(5),
				cache:     tt.args.cacheMock,
				config: Config{
					PollingInterval: 0,
					Logger:          logger,
					Offline:         tt.args.offline,
				},
				dataExporter: dataexporter.NewScheduler(context.Background(), 0, 0,
					&logsexporter.Exporter{
						LogFormat: "[{{ .FormattedDate}}] user=\"{{ .UserKey}}\", flag=\"{{ .Key}}\", " +
							"value=\"{{ .Value}}\", variation=\"{{ .Variation}}\"",
					}, logger),
			}
			got, err := IntVariation(tt.args.flagKey, tt.args.user, tt.args.defaultValue)

			if tt.expectedLog != "" {
				content, _ := ioutil.ReadFile(file.Name())
				assert.Regexp(t, tt.expectedLog, string(content))
			}

			if tt.wantErr {
				assert.Error(t, err, "IntVariation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got, "IntVariation() got = %v, want %v", got, tt.want)

			// clean logger
			ff = nil
			_ = file.Close()
		})
	}
}

func TestAllFlagsState(t *testing.T) {
	tests := []struct {
		name       string
		config     Config
		valid      bool
		jsonOutput string
		initModule bool
	}{
		{
			name: "Valid multiple types",
			config: Config{
				Retriever: &fileretriever.Retriever{
					Path: "./testdata/ffclient/all_flags/config_flag/flag-config-all-flags.yaml",
				},
			},
			valid:      true,
			jsonOutput: "./testdata/ffclient/all_flags/marshal_json/valid_multiple_types.json",
			initModule: true,
		},
		{
			name: "Error in flag-0",
			config: Config{
				Retriever: &fileretriever.Retriever{
					Path: "./testdata/ffclient/all_flags/config_flag/flag-config-with-error.yaml",
				},
			},
			valid:      false,
			jsonOutput: "./testdata/ffclient/all_flags/marshal_json/error_in_flag_0.json",
			initModule: true,
		},
		{
			name: "module not init",
			config: Config{
				Retriever: &fileretriever.Retriever{
					Path: "./testdata/ffclient/all_flags/config_flag/flag-config-all-flags.yaml",
				},
			},
			valid:      false,
			jsonOutput: "./testdata/ffclient/all_flags/marshal_json/module_not_init.json",
			initModule: false,
		},
		{
			name: "offline",
			config: Config{
				Offline: true,
				Retriever: &fileretriever.Retriever{
					Path: "./testdata/ffclient/all_flags/config_flag/flag-config-all-flags.yaml",
				},
			},
			valid:      true,
			jsonOutput: "./testdata/ffclient/all_flags/marshal_json/offline.json",
			initModule: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// init logger
			exportDir, _ := ioutil.TempDir("", "export")
			tt.config.DataExporter = DataExporter{
				FlushInterval:    1000,
				MaxEventInMemory: 1,
				Exporter:         &fileexporter.Exporter{OutputDir: exportDir},
			}

			var goff *GoFeatureFlag
			var err error
			if tt.initModule {
				goff, err = New(tt.config)
				assert.NoError(t, err)
				defer goff.Close()
			} else {
				// we close directly so we can test with module not init
				goff, _ = New(tt.config)
				goff.Close()
			}

			user := ffuser.NewUser("random-key")
			allFlagsState := goff.AllFlagsState(user)
			assert.Equal(t, tt.valid, allFlagsState.IsValid())

			// expected JSON output - we force the timestamp
			expected, _ := ioutil.ReadFile(tt.jsonOutput)
			var f map[string]interface{}
			_ = json.Unmarshal(expected, &f)
			if expectedFlags, ok := f["flags"].(map[string]interface{}); ok {
				for _, value := range expectedFlags {
					if valueObj, ok := value.(map[string]interface{}); ok {
						assert.NotNil(t, valueObj["timestamp"])
						assert.NotEqual(t, 0, valueObj["timestamp"])
						valueObj["timestamp"] = time.Now().Unix()
					}
				}
			}
			expectedJSON, _ := json.Marshal(f)
			marshaled, err := allFlagsState.MarshalJSON()
			assert.NoError(t, err)
			assert.JSONEq(t, string(expectedJSON), string(marshaled))

			// no data exported
			files, _ := os.ReadDir(exportDir)
			assert.Equal(t, 0, len(files))
		})
	}
}

func TestAllFlagsFromCache(t *testing.T) {
	tests := []struct {
		name       string
		config     Config
		initModule bool
	}{
		{
			name: "Valid multiple types",
			config: Config{
				Retriever: &fileretriever.Retriever{
					Path: "./testdata/ffclient/all_flags/config_flag/flag-config-all-flags.yaml",
				},
			},
			initModule: true,
		},
		{
			name: "module not init",
			config: Config{
				Retriever: &fileretriever.Retriever{
					Path: "./testdata/ffclient/all_flags/config_flag/flag-config-all-flags.yaml",
				},
			},
			initModule: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var goff *GoFeatureFlag
			var err error
			if tt.initModule {
				goff, err = New(tt.config)
				assert.NoError(t, err)
				defer goff.Close()

				flags, err := goff.GetFlagsFromCache()
				assert.NoError(t, err)

				cf, _ := goff.cache.AllFlags()
				assert.Equal(t, flags, cf)
			} else {
				// we close directly so we can test with module not init
				goff, _ = New(tt.config)
				goff.Close()

				_, err := goff.GetFlagsFromCache()
				assert.Error(t, err)
			}
		})
	}
}

func TestRawVariation(t *testing.T) {
	type args struct {
		flagKey      string
		user         ffuser.User
		defaultValue interface{}
		cacheMock    cache.Manager
		offline      bool
	}
	tests := []struct {
		name        string
		args        args
		want        model.RawVarResult
		wantErr     bool
		expectedLog string
	}{
		{
			name: "Get default value if flag disable",
			args: args{
				flagKey:      "disable-flag",
				user:         ffuser.NewUser("random-key"),
				defaultValue: true,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Disable: testconvert.Bool(true),
				}, nil),
			},
			want: model.RawVarResult{
				Value: true,
				VariationResult: model.VariationResult{
					VariationType: flag.VariationSDKDefault,
					Failed:        false,
					TrackEvents:   true,
					Reason:        flag.ReasonDisabled,
				},
			},
			wantErr:     false,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key\", flag=\"disable-flag\", value=\"true\", variation=\"SdkDefault\"\n",
		},
		{
			name: "Get error when not init",
			args: args{
				flagKey:      "key-not-exist",
				user:         ffuser.NewUser("random-key"),
				defaultValue: "defaultValue",
				cacheMock: NewCacheMock(
					&flag.InternalFlag{},
					errors.New("impossible to read the toggle before the initialisation")),
			},
			want: model.RawVarResult{
				Value: "defaultValue",
				VariationResult: model.VariationResult{
					VariationType: flag.VariationSDKDefault,
					Failed:        true,
					TrackEvents:   true,
					Reason:        flag.ReasonError,
					ErrorCode:     flag.ErrorCodeFlagNotFound,
				},
			},
			wantErr:     true,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key\", flag=\"key-not-exist\", value=\"defaultValue\", variation=\"SdkDefault\"\n",
		},
		{
			name: "Get default value with key not exist",
			args: args{
				flagKey:      "key-not-exist",
				user:         ffuser.NewUser("random-key"),
				defaultValue: 123456,
				cacheMock:    NewCacheMock(&flag.InternalFlag{}, errors.New("flag [key-not-exist] does not exists")),
			},
			want: model.RawVarResult{
				Value: 123456,
				VariationResult: model.VariationResult{
					VariationType: flag.VariationSDKDefault,
					Failed:        true,
					TrackEvents:   true,
					Reason:        flag.ReasonError,
					ErrorCode:     flag.ErrorCodeFlagNotFound,
				},
			},
			wantErr:     true,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key\", flag=\"key-not-exist\", value=\"123456\", variation=\"SdkDefault\"\n",
		},
		{
			name: "Get default value, rule not apply",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewUser("random-key"),
				defaultValue: map[string]interface{}{"test123": "test"},
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("legacyRuleV0"),
							Query: testconvert.String("key eq \"key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*interface{}{
						"Default": testconvert.Interface(map[string]interface{}{"test": "test"}),
						"True":    testconvert.Interface(map[string]interface{}{"test2": "test"}),
						"False":   testconvert.Interface(map[string]interface{}{"test3": "test"}),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("legacyDefaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want: model.RawVarResult{
				Value: map[string]interface{}{"test": "test"},
				VariationResult: model.VariationResult{
					VariationType: "Default",
					Failed:        false,
					TrackEvents:   true,
					Reason:        flag.ReasonDefault,
				},
			},
			wantErr:     false,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key\", flag=\"test-flag\", value=\"map\\[test:test\\]\", variation=\"Default\"",
		},
		{
			name: "Get true value, rule apply",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewAnonymousUser("random-key"),
				defaultValue: map[string]interface{}{"test123": "test"},
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("legacyRuleV0"),
							Query: testconvert.String("key eq \"random-key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					Variations: &map[string]*interface{}{
						"Default": testconvert.Interface(map[string]interface{}{"test": "test"}),
						"True":    testconvert.Interface(map[string]interface{}{"test2": "test"}),
						"False":   testconvert.Interface(map[string]interface{}{"test3": "test"}),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("legacyDefaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want: model.RawVarResult{
				Value: map[string]interface{}{"test2": "test"},
				VariationResult: model.VariationResult{
					VariationType: "True",
					Failed:        false,
					TrackEvents:   true,
					Reason:        flag.ReasonTargetingMatch,
				},
			},
			wantErr:     false,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key\", flag=\"test-flag\", value=\"map\\[test2:test\\]\", variation=\"True\"",
		},
		{
			name: "Get false value, rule apply",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewAnonymousUser("random-key-ssss1"),
				defaultValue: map[string]interface{}{"test123": "test"},
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("legacyRuleV0"),
							Query: testconvert.String("anonymous eq true"),
							Percentages: &map[string]float64{
								"False": 90,
								"True":  10,
							},
						},
					},
					Variations: &map[string]*interface{}{
						"Default": testconvert.Interface(map[string]interface{}{"test": "test"}),
						"True":    testconvert.Interface(map[string]interface{}{"test2": "test"}),
						"False":   testconvert.Interface(map[string]interface{}{"test3": "test"}),
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("legacyDefaultRule"),
						VariationResult: testconvert.String("Default"),
					},
				}, nil),
			},
			want: model.RawVarResult{
				Value: map[string]interface{}{"test3": "test"},
				VariationResult: model.VariationResult{
					VariationType: "False",
					Failed:        false,
					TrackEvents:   true,
					Reason:        flag.ReasonTargetingMatch,
				},
			},
			wantErr:     false,
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"random-key-ssss1\", flag=\"test-flag\", value=\"map\\[test3:test\\]\", variation=\"False\"",
		},
		{
			name: "No exported log",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewAnonymousUser("random-key"),
				defaultValue: true,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Variations: &map[string]*interface{}{
						"Default": testconvert.Interface(false),
						"True":    testconvert.Interface(true),
						"False":   testconvert.Interface(false),
					},
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("legacyRuleV0"),
							Query: testconvert.String("key eq \"random-key\""),
							Percentages: &map[string]float64{
								"False": 0,
								"True":  100,
							},
						},
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("legacyDefaultRule"),
						VariationResult: testconvert.String("Default"),
					},
					TrackEvents: testconvert.Bool(false),
				}, nil),
			},
			want: model.RawVarResult{
				Value: true,
				VariationResult: model.VariationResult{
					VariationType: "True",
					Failed:        false,
					TrackEvents:   false,
					Reason:        flag.ReasonTargetingMatch,
				},
			},
			wantErr:     false,
			expectedLog: "^$",
		},
		{
			name: "Get sdk default value if offline",
			args: args{
				offline:      true,
				flagKey:      "disable-flag",
				user:         ffuser.NewUser("random-key"),
				defaultValue: false,
				cacheMock: NewCacheMock(&flag.InternalFlag{
					Disable: testconvert.Bool(true),
				}, nil),
			},
			want: model.RawVarResult{
				Value: false,
				VariationResult: model.VariationResult{
					VariationType: flag.VariationSDKDefault,
					Failed:        true,
					TrackEvents:   false,
				},
			},
			wantErr:     false,
			expectedLog: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// init logger
			file, _ := ioutil.TempFile("", "log")
			logger := log.New(file, "", 0)

			ff = &GoFeatureFlag{
				bgUpdater: newBackgroundUpdater(5),
				cache:     tt.args.cacheMock,
				config: Config{
					PollingInterval: 0,
					Logger:          logger,
					Offline:         tt.args.offline,
				},
				dataExporter: dataexporter.NewScheduler(context.Background(), 0, 0,
					&logsexporter.Exporter{
						LogFormat: "[{{ .FormattedDate}}] user=\"{{ .UserKey}}\", flag=\"{{ .Key}}\", " +
							"value=\"{{ .Value}}\", variation=\"{{ .Variation}}\"",
					}, logger),
			}

			got, err := ff.RawVariation(tt.args.flagKey, tt.args.user, tt.args.defaultValue)

			if tt.expectedLog != "" {
				content, _ := ioutil.ReadFile(file.Name())
				assert.Regexp(t, tt.expectedLog, string(content))
			}

			if tt.wantErr {
				assert.Error(t, err, "RawVariation() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.want, got, "RawVariation() got = %v, want %v", got, tt.want)

			// clean logger
			ff = nil
			_ = file.Close()
		})
	}
}

package cache_test

import (
	"log"
	"os"
	"testing"

	"github.com/thomaspoignant/go-feature-flag/internal/flag"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/internal/cache"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"github.com/thomaspoignant/go-feature-flag/testutils/testconvert"
)

func Test_FlagCacheNotInit(t *testing.T) {
	fCache := cache.New(nil, nil)
	fCache.Close()
	_, err := fCache.GetFlag("test-flag")
	assert.Error(t, err, "We should have an error if the cache is not init")
}

func Test_GetFlagNotExist(t *testing.T) {
	fCache := cache.New(nil, nil)
	_, err := fCache.GetFlag("not-exists-flag")
	assert.Error(t, err, "We should have an error if the flag does not exists")
}

func Test_FlagCache(t *testing.T) {
	yamlFile := []byte(`test-flag:
  rule: key eq "random-key"
  percentage: 100
  true: true
  false: false
  default: false
  trackEvents: false
`)

	jsonFile := []byte(`{
	 "test-flag": {
	   "rule": "key eq \"random-key\"",
	   "percentage": 100,
	   "true": true,
	   "false": false,
	   "default": false
	 }
	}
	`)

	tomlFile := []byte(`[test-flag]
	rule = "key eq \"random-key\""
	percentage = 100.0
	true = true
	false = false
	default = false
	disable = false`)

	type args struct {
		loadedFlags []byte
	}
	tests := []struct {
		name       string
		args       args
		expected   map[string]flag.InternalFlag
		wantErr    bool
		flagFormat string
	}{
		{
			name:       "Yaml valid",
			flagFormat: "yaml",
			args: args{
				loadedFlags: yamlFile,
			},
			expected: map[string]flag.InternalFlag{
				"test-flag": {
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
				},
			},
			wantErr: false,
		},
		{
			name:       "Yaml invalid file",
			flagFormat: "yaml",
			args: args{
				loadedFlags: []byte(`test-flag:
		 rule: key eq "random-key"
		 percentage: "toot"
		 true: true
		 false: false
		 default: false
		`),
			},
			wantErr: true,
		},
		{
			name: "JSON valid",
			args: args{
				loadedFlags: jsonFile,
			},
			flagFormat: "json",
			expected: map[string]flag.InternalFlag{
				"test-flag": {
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
				},
			},
			wantErr: false,
		},
		{
			name:       "JSON invalid file",
			flagFormat: "json",
			args: args{
				loadedFlags: []byte(`{
		"test-flag": {
		  "rule": "key eq \"random-key\"",
		  "percentage": 100,
		  "true": true,
		  "false": false,
		  "default": false"
		}
		}`),
			},
			wantErr: true,
		},
		{
			name: "TOML valid",
			args: args{
				loadedFlags: tomlFile,
			},
			flagFormat: "toml",
			expected: map[string]flag.InternalFlag{
				"test-flag": {
					Disable: testconvert.Bool(false),
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
				},
			},
			wantErr: false,
		},
		{
			name: "TOML invalid file",
			args: args{
				loadedFlags: []byte(`[test-flag]
		rule = "key eq \"random-key\""
		percentage = 100.0
		true = true
		false = false
		default = false"
		disable = false`),
			},
			flagFormat: "toml",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fCache := cache.New(cache.NewNotificationService([]notifier.Notifier{}), nil)
			err := fCache.UpdateCache(tt.args.loadedFlags, tt.flagFormat, log.New(os.Stdout, "", 0))
			if tt.wantErr {
				assert.Error(t, err, "UpdateCache() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.NoError(t, err, "UpdateCache() error = %v, wantErr %v", err, tt.wantErr)
			// If no error we compare with expected
			for key, expected := range tt.expected {
				got, _ := fCache.GetFlag(key)
				assert.Equal(t, &expected, got) // nolint
			}
			fCache.Close()
		})
	}
}

func Test_AllFlags(t *testing.T) {
	yamlFile := []byte(`test-flag:
  rule: key eq "random-key"
  percentage: 100
  true: true
  false: false
  default: false
  trackEvents: false
`)

	type args struct {
		loadedFlags []byte
	}
	tests := []struct {
		name       string
		args       args
		expected   map[string]flag.InternalFlag
		wantErr    bool
		flagFormat string
	}{
		{
			name:       "Yaml valid",
			flagFormat: "yaml",
			args: args{
				loadedFlags: yamlFile,
			},
			expected: map[string]flag.InternalFlag{
				"test-flag": {
					Variations: &map[string]*interface{}{
						"Default": testconvert.Interface(false),
						"False":   testconvert.Interface(false),
						"True":    testconvert.Interface(true),
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
				},
			},
			wantErr: false,
		},
		{
			name:       "Yaml multiple flags",
			flagFormat: "yaml",
			args: args{
				loadedFlags: []byte(`test-flag:
  rule: key eq "random-key"
  percentage: 100
  true: true
  false: false
  default: false
  trackEvents: false
test-flag2:
  rule: key eq "random-key"
  percentage: 0
  true: "true"
  false: "false"
  default: "false"
  trackEvents: false
`),
			},
			expected: map[string]flag.InternalFlag{
				"test-flag": {
					Variations: &map[string]*interface{}{
						"Default": testconvert.Interface(false),
						"False":   testconvert.Interface(false),
						"True":    testconvert.Interface(true),
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
				},
				"test-flag2": {
					Variations: &map[string]*interface{}{
						"Default": testconvert.Interface("false"),
						"False":   testconvert.Interface("false"),
						"True":    testconvert.Interface("true"),
					},
					Rules: &[]flag.Rule{
						{
							Name:  testconvert.String("legacyRuleV0"),
							Query: testconvert.String("key eq \"random-key\""),
							Percentages: &map[string]float64{
								"False": 100,
								"True":  0,
							},
						},
					},
					DefaultRule: &flag.Rule{
						Name:            testconvert.String("legacyDefaultRule"),
						VariationResult: testconvert.String("Default"),
					},
					TrackEvents: testconvert.Bool(false),
				},
			},
			wantErr: false,
		},
		{
			name:       "empty",
			flagFormat: "yaml",
			args: args{
				loadedFlags: []byte(``),
			},
			expected: map[string]flag.InternalFlag{},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fCache := cache.New(cache.NewNotificationService([]notifier.Notifier{}), nil)
			_ = fCache.UpdateCache(tt.args.loadedFlags, tt.flagFormat, log.New(os.Stdout, "", 0))

			allFlags, err := fCache.AllFlags()
			if tt.wantErr {
				assert.Error(t, err, "UpdateCache() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.NoError(t, err)

			// If no error we compare with expected
			for key, expected := range tt.expected {
				got := allFlags[key]
				assert.Equal(t, &expected, got) //nolint: gosec
			}
			fCache.Close()
		})
	}
}

func Test_cacheManagerImpl_GetLatestUpdateDate(t *testing.T) {
	loadedFlags := []byte(`test-flag:
  rule: key eq "random-key"
  percentage: 100
  true: true
  false: false
  default: false
  trackEvents: false
`)

	fCache := cache.New(cache.NewNotificationService([]notifier.Notifier{}), nil)
	timeBefore := fCache.GetLatestUpdateDate()
	_ = fCache.UpdateCache(loadedFlags, "yaml", log.New(os.Stdout, "", 0))
	timeAfter := fCache.GetLatestUpdateDate()

	assert.True(t, timeBefore.Before(timeAfter))
}

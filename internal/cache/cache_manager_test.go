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
	yamlFile := []byte(`
test-flag:
  variations:
    true_var: true
    false_var: false
  targeting:
    - query: key eq "random-key"
      percentage:
        true_var: 100
        false_var: 0
  defaultRule:
    variation: false_var	
  trackEvents: false`)

	jsonFile := []byte(`{
  "test-flag": {
    "variations": {
      "true_var": true,
      "false_var": false
    },
    "targeting": [
      {
        "query": "key eq \"random-key\"",
        "percentage": {
          "true_var": 100,
          "false_var": 0
        }
      }
    ],
    "defaultRule": {
      "variation": "false_var"
    },
		"trackEvents": false
  }
}
	`)

	tomlFile := []byte(`[test-flag]
trackEvents = false

  [test-flag.variations]
  true_var = true
  false_var = false

  [[test-flag.targeting]]
  query = 'key eq "random-key"'

    [test-flag.targeting.percentage]
    true_var = 100.00
    false_var = 0.00

  [test-flag.defaultRule]
  variation = "false_var"`)

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
							Query: testconvert.String("key eq \"random-key\""),
							Percentages: &map[string]float64{
								"false_var": 0,
								"true_var":  100,
							},
						},
					},
					Variations: &map[string]*interface{}{
						"false_var": testconvert.Interface(false),
						"true_var":  testconvert.Interface(true),
					},
					DefaultRule: &flag.Rule{
						VariationResult: testconvert.String("false_var"),
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
				loadedFlags: []byte(`
test-flag:
  variations:
    true_var: true
    false_var: false
  targeting:
    - query: key eq "random-key"
      percentage: "toto"
  defaultRule:
    variation: false_var	
  trackEvents: false`),
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
							Query: testconvert.String("key eq \"random-key\""),
							Percentages: &map[string]float64{
								"false_var": 0,
								"true_var":  100,
							},
						},
					},
					Variations: &map[string]*interface{}{
						"false_var": testconvert.Interface(false),
						"true_var":  testconvert.Interface(true),
					},
					DefaultRule: &flag.Rule{
						VariationResult: testconvert.String("false_var"),
					},
					TrackEvents: testconvert.Bool(false),
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
    "variations": {
      "true_var": true,
      "false_var": false
    },
    "targeting": [
      {
        "query": "key eq \"random-key\"",
        "percentage": "toto"
      }
    ],
    "defaultRule": {
      "variation": "false_var"
    },
    "trackEvents": false
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
					Rules: &[]flag.Rule{
						{
							Query: testconvert.String("key eq \"random-key\""),
							Percentages: &map[string]float64{
								"false_var": 0,
								"true_var":  100,
							},
						},
					},
					Variations: &map[string]*interface{}{
						"false_var": testconvert.Interface(false),
						"true_var":  testconvert.Interface(true),
					},
					DefaultRule: &flag.Rule{
						VariationResult: testconvert.String("false_var"),
					},
					TrackEvents: testconvert.Bool(false),
				},
			},
			wantErr: false,
		},
		{
			name: "TOML invalid file",
			args: args{
				loadedFlags: []byte(`
[test-flag]
trackEvents = false
[test-flag.variations]
true_var = true
false_var = false
[[test-flag.targeting]]
query = 'key eq "random-key"'
percentage = "toto
[test-flag.defaultRule]
variation = "false_var"
`),
			},
			flagFormat: "toml",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fCache := cache.New(cache.NewNotificationService([]notifier.Notifier{}), nil)
			newFlags, err := fCache.ConvertToFlagStruct(tt.args.loadedFlags, tt.flagFormat)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			err = fCache.UpdateCache(newFlags, log.New(os.Stdout, "", 0))
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
	yamlFile := []byte(`
test-flag:
  variations:
    true_var: true
    false_var: false
  targeting:
    - query: key eq "random-key"
      percentage:
        true_var: 100
        false_var: 0
  defaultRule:
    variation: false_var	
  trackEvents: false`)

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
						"false_var": testconvert.Interface(false),
						"true_var":  testconvert.Interface(true),
					},
					Rules: &[]flag.Rule{
						{
							Query: testconvert.String("key eq \"random-key\""),
							Percentages: &map[string]float64{
								"false_var": 0,
								"true_var":  100,
							},
						},
					},
					DefaultRule: &flag.Rule{
						VariationResult: testconvert.String("false_var"),
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
				loadedFlags: []byte(`
test-flag:
  variations:
    true_var: true
    false_var: false
  targeting:
    - query: key eq "random-key"
      percentage:
        true_var: 100
        false_var: 0
  defaultRule:
    variation: false_var	
  trackEvents: false

test-flag2:
  variations:
    true_var: "true"
    false_var: "false"
  targeting:
    - query: key eq "random-key"
      percentage:
        true_var: 0
        false_var: 100
  defaultRule:
    variation: false_var  
  trackEvents: false
`),
			},
			expected: map[string]flag.InternalFlag{
				"test-flag": {
					Variations: &map[string]*interface{}{
						"false_var": testconvert.Interface(false),
						"true_var":  testconvert.Interface(true),
					},
					Rules: &[]flag.Rule{
						{
							Query: testconvert.String("key eq \"random-key\""),
							Percentages: &map[string]float64{
								"false_var": 0,
								"true_var":  100,
							},
						},
					},
					DefaultRule: &flag.Rule{
						VariationResult: testconvert.String("false_var"),
					},
					TrackEvents: testconvert.Bool(false),
				},
				"test-flag2": {
					Variations: &map[string]*interface{}{
						"false_var": testconvert.Interface("false"),
						"true_var":  testconvert.Interface("true"),
					},
					Rules: &[]flag.Rule{
						{
							Query: testconvert.String("key eq \"random-key\""),
							Percentages: &map[string]float64{
								"false_var": 100,
								"true_var":  0,
							},
						},
					},
					DefaultRule: &flag.Rule{
						VariationResult: testconvert.String("false_var"),
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
			newFlags, err := fCache.ConvertToFlagStruct(tt.args.loadedFlags, tt.flagFormat)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			err = fCache.UpdateCache(newFlags, log.New(os.Stdout, "", 0))
			assert.NoError(t, err)

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
  variations:
    true_var: true
    false_var: false
  targeting:
    - query: key eq "random-key"
      percentage:
        true_var: 100
        false_var: 0
  defaultRule:
    variation: false_var	
  trackEvents: false
`)

	fCache := cache.New(cache.NewNotificationService([]notifier.Notifier{}), nil)
	timeBefore := fCache.GetLatestUpdateDate()
	newFlags, _ := fCache.ConvertToFlagStruct(loadedFlags, "yaml")
	_ = fCache.UpdateCache(newFlags, log.New(os.Stdout, "", 0))
	timeAfter := fCache.GetLatestUpdateDate()

	assert.True(t, timeBefore.Before(timeAfter))
}

package cache_test

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/internal/cache"
	"github.com/thomaspoignant/go-feature-flag/internal/notification"
	"github.com/thomaspoignant/go-feature-flag/modules/core/dto"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/modules/core/testutils/testconvert"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"github.com/thomaspoignant/go-feature-flag/testutils/mock"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
	"gopkg.in/yaml.v3"
)

func Test_FlagCacheNotInit(t *testing.T) {
	fCache := cache.New(nil, "", nil)
	fCache.Close()
	_, err := fCache.GetFlag("test-flag")
	assert.Error(t, err, "We should have an error if the cache is not init")
}

func Test_GetFlagNotExist(t *testing.T) {
	fCache := cache.New(nil, "", nil)
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
					Variations: &map[string]*any{
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
					Variations: &map[string]*any{
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
					Variations: &map[string]*any{
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
			fCache := cache.New(notification.NewService([]notifier.Notifier{}), "",
				&fflog.FFLogger{LeveledLogger: slog.Default()})
			newFlags, err := cache.ConvertToFlagStruct(tt.args.loadedFlags, tt.flagFormat)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			err = fCache.UpdateCache(newFlags, nil, true)
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
					Variations: &map[string]*any{
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
					Variations: &map[string]*any{
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
					Variations: &map[string]*any{
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
			fCache := cache.New(notification.NewService([]notifier.Notifier{}), "", nil)
			newFlags, err := cache.ConvertToFlagStruct(tt.args.loadedFlags, tt.flagFormat)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			err = fCache.UpdateCache(newFlags, &fflog.FFLogger{LeveledLogger: slog.Default()}, true)
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

	fCache := cache.New(notification.NewService([]notifier.Notifier{}), "", nil)
	timeBefore := fCache.GetLatestUpdateDate()
	newFlags, _ := cache.ConvertToFlagStruct(loadedFlags, "yaml")
	_ = fCache.UpdateCache(newFlags, &fflog.FFLogger{LeveledLogger: slog.Default()}, true)
	timeAfter := fCache.GetLatestUpdateDate()

	assert.True(t, timeBefore.Before(timeAfter))
}

func Test_persistCacheAndRestartCacheWithIt(t *testing.T) {
	file, err := os.CreateTemp("", "")
	assert.NoError(t, err)
	defer func() { _ = file.Close() }()

	// We start without any cache and we persist it on disk
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
	loadedFlagsMap := map[string]dto.DTO{}
	err = yaml.Unmarshal(loadedFlags, &loadedFlagsMap)
	assert.NoError(t, err)

	fCache := cache.New(notification.NewService([]notifier.Notifier{}), file.Name(), nil)
	err = fCache.UpdateCache(loadedFlagsMap, &fflog.FFLogger{LeveledLogger: slog.Default()}, true)
	assert.NoError(t, err)
	allFlags1, err := fCache.AllFlags()
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond) // waiting to let the go routine write in the file

	// we start a new cache with the file persisted
	fCache2 := cache.New(notification.NewService([]notifier.Notifier{}), "", nil)
	content, err := os.ReadFile(file.Name())
	assert.NoError(t, err)
	loadedFlagsMap2 := map[string]dto.DTO{}
	err = yaml.Unmarshal(content, &loadedFlagsMap)
	assert.NoError(t, err)
	err = fCache2.UpdateCache(loadedFlagsMap2, &fflog.FFLogger{LeveledLogger: slog.Default()}, true)
	assert.NoError(t, err)
	allFlags2, err := fCache.AllFlags()
	assert.NoError(t, err)

	// Compare the 2 caches
	assert.Equal(t, allFlags1, allFlags2)
}

func TestCacheManager_UpdateCache(t *testing.T) {
	tests := []struct {
		name         string
		initialFlags map[string]dto.DTO
		updatedFlags map[string]dto.DTO
	}{
		{
			name: "Update existing flags",
			initialFlags: map[string]dto.DTO{
				"flag1": {
					Variations: &map[string]*any{},
					DefaultRule: &flag.Rule{
						VariationResult: testconvert.String("true"),
					},
				},
			},
			updatedFlags: map[string]dto.DTO{
				"flag1": {
					Variations: &map[string]*any{
						"true": testconvert.Interface(true),
					},
					DefaultRule: &flag.Rule{
						VariationResult: testconvert.String("true"),
					},
				},
				"flag2": {
					Variations: &map[string]*any{
						"false": testconvert.Interface(false),
					},
					DefaultRule: &flag.Rule{
						VariationResult: testconvert.String("false"),
					},
				},
			},
		},
		{
			name:         "Empty initial flags",
			initialFlags: map[string]dto.DTO{},
			updatedFlags: map[string]dto.DTO{
				"flag1": {
					Variations: &map[string]*any{
						"true": testconvert.Interface(true),
					},
					DefaultRule: &flag.Rule{
						VariationResult: testconvert.String("true"),
					},
				},
			},
		},
		{
			name: "Remove a flag",
			initialFlags: map[string]dto.DTO{
				"flag1": {
					Variations: &map[string]*any{
						"true": testconvert.Interface(true),
					},
					DefaultRule: &flag.Rule{
						VariationResult: testconvert.String("true"),
					},
				},
				"flag2": {
					Variations: &map[string]*any{
						"false": testconvert.Interface(false),
					},
					DefaultRule: &flag.Rule{
						VariationResult: testconvert.String("false"),
					},
				},
			},
			updatedFlags: map[string]dto.DTO{
				"flag1": {
					Variations: &map[string]*any{
						"true": testconvert.Interface(true),
					},
					DefaultRule: &flag.Rule{
						VariationResult: testconvert.String("true"),
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test UpdateCache without notification
			mockNotificationService := &mock.NotificationService{}
			cm := cache.New(
				mockNotificationService,
				"",
				&fflog.FFLogger{LeveledLogger: slog.Default()},
			)

			err := cm.UpdateCache(tt.initialFlags, nil, false)
			assert.NoError(t, err)

			err = cm.UpdateCache(tt.updatedFlags, nil, false)
			assert.NoError(t, err)
			assert.Equal(
				t,
				0,
				mockNotificationService.GetNotifyCalls(),
				"Notify should not be called for UpdateCache",
			)

			flags, err := cm.AllFlags()
			assert.NoError(t, err)
			assert.Len(
				t,
				flags,
				len(tt.updatedFlags),
				"Cache should be updated with correct number of flags",
			)

			// Test UpdateCacache with notification
			mockNotificationService = &mock.NotificationService{}
			cm = cache.New(
				mockNotificationService,
				"",
				&fflog.FFLogger{LeveledLogger: slog.Default()},
			)

			err = cm.UpdateCache(tt.initialFlags, nil, false)
			assert.NoError(t, err)

			err = cm.UpdateCache(tt.updatedFlags, nil, true)
			assert.NoError(t, err)
			assert.Equal(
				t,
				1,
				mockNotificationService.GetNotifyCalls(),
				"Notify should be called once for UpdateCache with notification",
			)

			flags, err = cm.AllFlags()
			assert.NoError(t, err)
			assert.Len(
				t,
				flags,
				len(tt.updatedFlags),
				"Cache should be updated with correct number of flags",
			)
		})
	}
}

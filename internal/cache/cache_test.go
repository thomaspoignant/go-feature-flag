package cache

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/thomaspoignant/go-feature-flag/internal/model"
	"github.com/thomaspoignant/go-feature-flag/internal/notifier"
)

func Test_FlagCacheNotInit(t *testing.T) {
	fCache := cacheImpl{}
	_, err := fCache.GetFlag("test-flag")
	assert.Error(t, err, "We should have an error if the cache is not init")
}

func Test_GetFlagNotExist(t *testing.T) {
	fCache := New(nil)
	_, err := fCache.GetFlag("not-exists-flag")
	assert.Error(t, err, "We should have an error if the flag does not exists")
}

func Test_FlagCache_yaml(t *testing.T) {
	exampleFile := []byte(`test-flag:
  rule: key eq "random-key"
  percentage: 100
  true: true
  false: false
  default: false
`)

	type args struct {
		loadedFlags []byte
	}
	tests := []struct {
		name     string
		args     args
		expected map[string]model.Flag
		wantErr  bool
	}{
		{
			name: "Add valid",
			args: args{
				loadedFlags: exampleFile,
			},
			expected: map[string]model.Flag{
				"test-flag": {
					Disable:    false,
					Rule:       "key eq \"random-key\"",
					Percentage: 100,
					True:       true,
					False:      false,
					Default:    false,
				},
			},
			wantErr: false,
		},
		{
			name: "Add invalid yaml file",
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fCache := New(NewNotificationService([]notifier.Notifier{}))
			err := fCache.UpdateCache(tt.args.loadedFlags, "yaml")
			if tt.wantErr {
				assert.Error(t, err, "UpdateCache() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.NoError(t, err, "UpdateCache() error = %v, wantErr %v", err, tt.wantErr)
			// If no error we compare with expected
			for key, value := range tt.expected {
				got, _ := fCache.GetFlag(key)
				assert.Equal(t, value, got)
			}
			fCache.Close()
		})
	}
}

func Test_FlagCache_json(t *testing.T) {
	exampleFile := []byte(`{
  "test-flag": {
    "rule": "key eq \"random-key\"",
    "percentage": 100,
    "true": true,
    "false": false,
    "default": false
  }
}
`)

	type args struct {
		loadedFlags []byte
	}
	tests := []struct {
		name     string
		args     args
		expected map[string]model.Flag
		wantErr  bool
	}{
		{
			name: "Add valid",
			args: args{
				loadedFlags: exampleFile,
			},
			expected: map[string]model.Flag{
				"test-flag": {
					Disable:    false,
					Rule:       "key eq \"random-key\"",
					Percentage: 100,
					True:       true,
					False:      false,
					Default:    false,
				},
			},
			wantErr: false,
		},
		{
			name: "Add invalid yaml file",
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fCache := New(NewNotificationService([]notifier.Notifier{}))
			err := fCache.UpdateCache(tt.args.loadedFlags, "json")
			if tt.wantErr {
				assert.Error(t, err, "UpdateCache() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.NoError(t, err, "UpdateCache() error = %v, wantErr %v", err, tt.wantErr)
			for key, value := range tt.expected {
				got, _ := fCache.GetFlag(key)
				assert.Equal(t, value, got)
			}
			fCache.Close()
		})
	}
}

func Test_FlagCache_toml(t *testing.T) {
	exampleFile := []byte(`[test-flag]
rule = "key eq \"random-key\""
percentage = 100.0
true = true
false = false
default = false`)

	type args struct {
		loadedFlags []byte
	}
	tests := []struct {
		name     string
		args     args
		expected map[string]model.Flag
		wantErr  bool
	}{
		{
			name: "Add valid",
			args: args{
				loadedFlags: exampleFile,
			},
			expected: map[string]model.Flag{
				"test-flag": {
					Disable:    false,
					Rule:       "key eq \"random-key\"",
					Percentage: 100,
					True:       true,
					False:      false,
					Default:    false,
				},
			},
			wantErr: false,
		},
		{
			name: "Add invalid yaml file",
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fCache := New(NewNotificationService([]notifier.Notifier{}))
			err := fCache.UpdateCache(tt.args.loadedFlags, "TOML")
			if tt.wantErr {
				assert.Error(t, err, "UpdateCache() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.NoError(t, err, "UpdateCache() error = %v, wantErr %v", err, tt.wantErr)
			for key, value := range tt.expected {
				got, _ := fCache.GetFlag(key)
				assert.Equal(t, value, got)
			}
			fCache.Close()
		})
	}
}

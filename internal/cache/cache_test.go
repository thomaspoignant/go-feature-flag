package cache

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/thomaspoignant/go-feature-flag/internal/flags"
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

func Test_FlagCache(t *testing.T) {
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
		expected map[string]flags.Flag
		wantErr  bool
	}{
		{
			name: "Add valid",
			args: args{
				loadedFlags: exampleFile,
			},
			expected: map[string]flags.Flag{
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
			fCache := New(NewService([]Notifier{}))
			err := fCache.UpdateCache(tt.args.loadedFlags)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateCache() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If no error we compare with expected
			if err == nil {
				for key, value := range tt.expected {
					got, _ := fCache.GetFlag(key)
					assert.Equal(t, value, got)
				}
			}
			fCache.Close()
		})
	}
}



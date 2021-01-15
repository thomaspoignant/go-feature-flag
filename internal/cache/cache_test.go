package cache

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"testing"

	"github.com/thomaspoignant/go-feature-flag/internal/flags"
	"github.com/thomaspoignant/go-feature-flag/testutil"
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
			fCache := New(log.New(os.Stdout, "", 0))
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

func Test_LogFlagChanges(t *testing.T) {
	exampleFile := []byte(`test-flag:
  rule: key eq "random-key"
  percentage: 100
  true: true
  false: false
  default: false
`)

	type args struct {
		oldFlag     []byte
		loadedFlags []byte
	}
	tests := []struct {
		name     string
		args     args
		expected string
	}{
		{
			name: "Update flag",
			args: args{
				oldFlag: exampleFile,
				loadedFlags: []byte(`test-flag:
  percentage: 100
  true: true
  false: false
  default: false`),
			},
			expected: "\\[" + testutil.RFC3339Regex + "\\] flag test-flag updated, old=\\[percentage=100%, rule=\"key eq \"random-key\"\", true=\"true\", false=\"false\", true=\"false\", disable=\"false\"\\], new=\\[percentage=100%, true=\"true\", false=\"false\", true=\"false\", disable=\"false\"\\]",
		},
		{
			name: "Remove flag",
			args: args{
				oldFlag:     exampleFile,
				loadedFlags: []byte(``),
			},
			expected: "\\[" + testutil.RFC3339Regex + "\\] flag test-flag removed",
		},
		{
			name: "Disable flag",
			args: args{
				oldFlag: exampleFile,
				loadedFlags: []byte(`test-flag:
  rule: key eq "random-key"
  disable: true
  percentage: 100
  true: true
  false: false
  default: false
`),
			},
			expected: "\\[" + testutil.RFC3339Regex + "\\] flag test-flag is turned OFF",
		},
		{
			name: "Add flag",
			args: args{
				oldFlag: exampleFile,
				loadedFlags: []byte(`test-flag:
  rule: key eq "random-key"
  percentage: 100
  true: true
  false: false
  default: false
add-test-flag:
  rule: key eq "random-key"
  percentage: 100
  true: true
  false: false
  default: false`),
			},
			expected: "\\[" + testutil.RFC3339Regex + "\\] flag add-test-flag added",
		},
		{
			name: "Enable flag",
			args: args{
				oldFlag: []byte(`test-flag:
  rule: key eq "random-key"
  disable: true
  percentage: 100
  true: true
  false: false
  default: false
`),
				loadedFlags: []byte(`test-flag:
  rule: key eq "random-key"
  disable: false
  percentage: 100
  true: true
  false: false
  default: false
`),
			},
			expected: "\\[" + testutil.RFC3339Regex + "\\] flag test-flag is turned ON \\(flag=\\[percentage=100%, rule=\"key eq \"random-key\"\", true=\"true\", false=\"false\", true=\"false\", disable=\"false\"\\]\\)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var oldValue map[string]flags.Flag
			_ = yaml.Unmarshal(tt.args.oldFlag, &oldValue)

			// create temp log file
			logOutput, _ := ioutil.TempFile("", "")

			fCache := cacheImpl{
				flagsCache: oldValue,
				mutex:      sync.Mutex{},
				Logger:     log.New(logOutput, "", 0),
				waitGroup:  sync.WaitGroup{},
			}

			var newValue map[string]flags.Flag
			_ = yaml.Unmarshal(tt.args.loadedFlags, &newValue)

			// update the cache file
			fCache.logFlagChanges(oldValue, newValue)

			// get the logs
			log, _ := ioutil.ReadFile(logOutput.Name())
			assert.Regexp(t, tt.expected, string(log))

			// Remove temp log file
			os.Remove(logOutput.Name())
			fCache.Close()
		})
	}
}

func Test_cacheImpl_getCacheCopy(t *testing.T) {
	type fields struct {
		flagsCache map[string]flags.Flag
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]flags.Flag
	}{
		{
			name: "Copy with values",
			fields: fields{
				flagsCache: map[string]flags.Flag{
					"test": {
						Disable:    false,
						Rule:       "key eq \"toto\"",
						Percentage: 20,
						True:       true,
						False:      false,
						Default:    true,
					},
				},
			},
		},
		{
			name: "Copy without value",
			fields: fields{
				flagsCache: map[string]flags.Flag{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &cacheImpl{
				flagsCache: tt.fields.flagsCache,
			}
			assert.Equal(t, c.flagsCache, c.getCacheCopy())
		})
	}
}

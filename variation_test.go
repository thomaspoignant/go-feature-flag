package ffclient

import (
	"errors"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"testing"

	"github.com/thomaspoignant/go-feature-flag/ffuser"
	"github.com/thomaspoignant/go-feature-flag/internal/cache"
	"github.com/thomaspoignant/go-feature-flag/internal/model"
	"github.com/thomaspoignant/go-feature-flag/testutil"
)

type cacheMock struct {
	flag model.Flag
	err  error
}

func NewCacheMock(flag model.Flag, err error) cache.Cache {
	return &cacheMock{
		flag: flag,
		err:  err,
	}
}
func (c *cacheMock) UpdateCache(loadedFlags []byte, fileFormat string) error {
	return nil
}
func (c *cacheMock) Close() {}
func (c *cacheMock) GetFlag(key string) (model.Flag, error) {
	return c.flag, c.err
}

func TestBoolVariation(t *testing.T) {
	type args struct {
		flagKey      string
		user         ffuser.User
		defaultValue bool
		cacheMock    cache.Cache
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
				cacheMock: NewCacheMock(model.Flag{
					Disable: true,
				}, nil),
			},
			want:        true,
			wantErr:     true,
			expectedLog: "\\[" + testutil.RFC3339Regex + "\\] user=\"random-key\", flag=\"disable-flag\", value=\"true\"\n",
		},
		{
			name: "Get error when not init",
			args: args{
				flagKey:      "key-not-exist",
				user:         ffuser.NewUser("random-key"),
				defaultValue: true,
				cacheMock: NewCacheMock(
					model.Flag{},
					errors.New("impossible to read the toggle before the initialisation")),
			},
			want:        true,
			wantErr:     true,
			expectedLog: "\\[" + testutil.RFC3339Regex + "\\] user=\"random-key\", flag=\"key-not-exist\", value=\"true\"\n",
		},
		{
			name: "Get default value with key not exist",
			args: args{
				flagKey:      "key-not-exist",
				user:         ffuser.NewUser("random-key"),
				defaultValue: true,
				cacheMock:    NewCacheMock(model.Flag{}, errors.New("flag [key-not-exist] does not exists")),
			},
			want:        true,
			wantErr:     true,
			expectedLog: "\\[" + testutil.RFC3339Regex + "\\] user=\"random-key\", flag=\"key-not-exist\", value=\"true\"\n",
		},
		{
			name: "Get default value, rule not apply",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewUser("random-key"),
				defaultValue: true,
				cacheMock: NewCacheMock(model.Flag{
					Rule:       "key eq \"key\"",
					Percentage: 100,
					Default:    true,
					True:       false,
					False:      false,
				}, nil),
			},
			want:        true,
			wantErr:     false,
			expectedLog: "\\[" + testutil.RFC3339Regex + "\\] user=\"random-key\", flag=\"test-flag\", value=\"true\"\n",
		},
		{
			name: "Get true value, rule apply",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewAnonymousUser("random-key"),
				defaultValue: true,
				cacheMock: NewCacheMock(model.Flag{
					Rule:       "key eq \"random-key\"",
					Percentage: 100,
					Default:    false,
					True:       true,
					False:      false,
				}, nil),
			},
			want:        true,
			wantErr:     false,
			expectedLog: "\\[" + testutil.RFC3339Regex + "\\] user=\"random-key\", flag=\"test-flag\", value=\"true\"\n",
		},
		{
			name: "Get false value, rule apply",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewAnonymousUser("random-key-ssss1"),
				defaultValue: true,
				cacheMock: NewCacheMock(model.Flag{
					Rule:       "anonymous eq true",
					Percentage: 50,
					Default:    true,
					True:       true,
					False:      false,
				}, nil),
			},
			want:        false,
			wantErr:     false,
			expectedLog: "\\[" + testutil.RFC3339Regex + "\\] user=\"random-key-ssss1\", flag=\"test-flag\", value=\"false\"\n",
		},
		{
			name: "Get default value, when rule apply and not right type",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewUser("random-key-ssss1"),
				defaultValue: true,
				cacheMock: NewCacheMock(model.Flag{
					Percentage: 100,
					Default:    "xxx",
					True:       "xxx",
					False:      "xxx",
				}, nil),
			},
			want:        true,
			wantErr:     true,
			expectedLog: "\\[" + testutil.RFC3339Regex + "\\] user=\"random-key-ssss1\", flag=\"test-flag\", value=\"true\"\n",
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
					PollInterval: 0,
					Logger:       logger,
				},
			}

			got, err := BoolVariation(tt.args.flagKey, tt.args.user, tt.args.defaultValue)

			if (err != nil) != tt.wantErr {
				t.Errorf("BoolVariation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("BoolVariation() got = %v, want %v", got, tt.want)
			}

			if tt.expectedLog != "" {
				content, _ := ioutil.ReadFile(file.Name())
				assert.Regexp(t, tt.expectedLog, string(content))
			}
			// clean logger
			ff = nil
			file.Close()
		})
	}
}

func TestFloat64Variation(t *testing.T) {
	type args struct {
		flagKey      string
		user         ffuser.User
		defaultValue float64
		cacheMock    cache.Cache
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
				cacheMock: NewCacheMock(model.Flag{
					Disable: true,
				}, nil),
			},
			want:        120.12,
			wantErr:     true,
			expectedLog: "\\[" + testutil.RFC3339Regex + "\\] user=\"random-key\", flag=\"disable-flag\", value=\"120.12\"\n",
		},
		{
			name: "Get error when not init",
			args: args{
				flagKey:      "key-not-exist",
				user:         ffuser.NewUser("random-key"),
				defaultValue: 118.12,
				cacheMock: NewCacheMock(
					model.Flag{},
					errors.New("impossible to read the toggle before the initialisation")),
			},
			want:        118.12,
			wantErr:     true,
			expectedLog: "\\[" + testutil.RFC3339Regex + "\\] user=\"random-key\", flag=\"key-not-exist\", value=\"118.12\"\n",
		},
		{
			name: "Get default value with key not exist",
			args: args{
				flagKey:      "key-not-exist",
				user:         ffuser.NewUser("random-key"),
				defaultValue: 118.12,
				cacheMock:    NewCacheMock(model.Flag{}, errors.New("flag [key-not-exist] does not exists")),
			},
			want:        118.12,
			wantErr:     true,
			expectedLog: "\\[" + testutil.RFC3339Regex + "\\] user=\"random-key\", flag=\"key-not-exist\", value=\"118.12\"\n",
		},
		{
			name: "Get default value, rule not apply",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewUser("random-key"),
				defaultValue: 118.12,
				cacheMock: NewCacheMock(model.Flag{
					Rule:       "key eq \"key\"",
					Percentage: 100,
					Default:    119.12,
					True:       120.12,
					False:      121.12,
				}, nil),
			},
			want:        119.12,
			wantErr:     false,
			expectedLog: "\\[" + testutil.RFC3339Regex + "\\] user=\"random-key\", flag=\"test-flag\", value=\"119.12\"\n",
		},
		{
			name: "Get true value, rule apply",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewAnonymousUser("random-key"),
				defaultValue: 118.12,
				cacheMock: NewCacheMock(model.Flag{
					Rule:       "key eq \"random-key\"",
					Percentage: 100,
					Default:    119.12,
					True:       120.12,
					False:      121.12,
				}, nil),
			},
			want:        120.12,
			wantErr:     false,
			expectedLog: "\\[" + testutil.RFC3339Regex + "\\] user=\"random-key\", flag=\"test-flag\", value=\"120.12\"\n",
		},
		{
			name: "Get false value, rule apply",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewAnonymousUser("random-key-ssss1"),
				defaultValue: 118.12,
				cacheMock: NewCacheMock(model.Flag{
					Rule:       "anonymous eq true",
					Percentage: 50,
					Default:    119.12,
					True:       120.12,
					False:      121.12,
				}, nil),
			},
			want:        121.12,
			wantErr:     false,
			expectedLog: "\\[" + testutil.RFC3339Regex + "\\] user=\"random-key-ssss1\", flag=\"test-flag\", value=\"121.12\"\n",
		},
		{
			name: "Get default value, when rule apply and not right type",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewUser("random-key-ssss1"),
				defaultValue: 118.12,
				cacheMock: NewCacheMock(model.Flag{
					Percentage: 100,
					Default:    "xxx",
					True:       "xxx",
					False:      "xxx",
				}, nil),
			},
			want:        118.12,
			wantErr:     true,
			expectedLog: "\\[" + testutil.RFC3339Regex + "\\] user=\"random-key-ssss1\", flag=\"test-flag\", value=\"118.12\"\n",
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
					PollInterval: 0,
					Logger:       logger,
				},
			}

			got, err := Float64Variation(tt.args.flagKey, tt.args.user, tt.args.defaultValue)

			if (err != nil) != tt.wantErr {
				t.Errorf("Float64Variation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Float64Variation() got = %v, want %v", got, tt.want)
			}

			if tt.expectedLog != "" {
				content, _ := ioutil.ReadFile(file.Name())
				assert.Regexp(t, tt.expectedLog, string(content))
			}
			// clean logger
			ff = nil
			file.Close()
		})
	}
}

func TestJSONArrayVariation(t *testing.T) {
	type args struct {
		flagKey      string
		user         ffuser.User
		defaultValue []interface{}
		cacheMock    cache.Cache
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
				cacheMock: NewCacheMock(model.Flag{
					Disable: true,
				}, nil),
			},
			want:        []interface{}{"toto"},
			wantErr:     true,
			expectedLog: "\\[" + testutil.RFC3339Regex + "\\] user=\"random-key\", flag=\"disable-flag\", value=\"\\[toto\\]\"\n",
		},
		{
			name: "Get error when not init",
			args: args{
				flagKey:      "key-not-exist",
				user:         ffuser.NewUser("random-key"),
				defaultValue: []interface{}{"toto"},
				cacheMock: NewCacheMock(
					model.Flag{},
					errors.New("impossible to read the toggle before the initialisation")),
			},
			want:        []interface{}{"toto"},
			wantErr:     true,
			expectedLog: "\\[" + testutil.RFC3339Regex + "\\] user=\"random-key\", flag=\"key-not-exist\", value=\"\\[toto\\]\"\n",
		},
		{
			name: "Get default value with key not exist",
			args: args{
				flagKey:      "key-not-exist",
				user:         ffuser.NewUser("random-key"),
				defaultValue: []interface{}{"toto"},
				cacheMock:    NewCacheMock(model.Flag{}, errors.New("flag [key-not-exist] does not exists")),
			},
			want:        []interface{}{"toto"},
			wantErr:     true,
			expectedLog: "\\[" + testutil.RFC3339Regex + "\\] user=\"random-key\", flag=\"key-not-exist\", value=\"\\[toto\\]\"\n",
		},
		{
			name: "Get default value, rule not apply",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewUser("random-key"),
				defaultValue: []interface{}{"toto"},
				cacheMock: NewCacheMock(model.Flag{
					Rule:       "key eq \"key\"",
					Percentage: 100,
					Default:    []interface{}{"default"},
					True:       []interface{}{"true"},
					False:      []interface{}{"false"},
				}, nil),
			},
			want:        []interface{}{"default"},
			wantErr:     false,
			expectedLog: "\\[" + testutil.RFC3339Regex + "\\] user=\"random-key\", flag=\"test-flag\", value=\"\\[default\\]\"\n",
		},
		{
			name: "Get true value, rule apply",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewAnonymousUser("random-key"),
				defaultValue: []interface{}{"toto"},
				cacheMock: NewCacheMock(model.Flag{
					Rule:       "key eq \"random-key\"",
					Percentage: 100,
					Default:    []interface{}{"default"},
					True:       []interface{}{"true"},
					False:      []interface{}{"false"},
				}, nil),
			},
			want:        []interface{}{"true"},
			wantErr:     false,
			expectedLog: "\\[" + testutil.RFC3339Regex + "\\] user=\"random-key\", flag=\"test-flag\", value=\"\\[true\\]\"\n",
		},
		{
			name: "Get false value, rule apply",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewAnonymousUser("random-key-ssss1"),
				defaultValue: []interface{}{"toto"},
				cacheMock: NewCacheMock(model.Flag{
					Rule:       "anonymous eq true",
					Percentage: 50,
					Default:    []interface{}{"default"},
					True:       []interface{}{"true"},
					False:      []interface{}{"false"},
				}, nil),
			},
			want:        []interface{}{"false"},
			wantErr:     false,
			expectedLog: "\\[" + testutil.RFC3339Regex + "\\] user=\"random-key-ssss1\", flag=\"test-flag\", value=\"\\[false\\]\"\n",
		},
		{
			name: "Get default value, when rule apply and not right type",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewUser("random-key-ssss1"),
				defaultValue: []interface{}{"toto"},
				cacheMock: NewCacheMock(model.Flag{
					Percentage: 100,
					Default:    "xxx",
					True:       "xxx",
					False:      "xxx",
				}, nil),
			},
			want:        []interface{}{"toto"},
			wantErr:     true,
			expectedLog: "\\[" + testutil.RFC3339Regex + "\\] user=\"random-key-ssss1\", flag=\"test-flag\", value=\"\\[toto\\]\"\n",
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
					PollInterval: 0,
					Logger:       logger,
				},
			}

			got, err := JSONArrayVariation(tt.args.flagKey, tt.args.user, tt.args.defaultValue)

			if (err != nil) != tt.wantErr {
				t.Errorf("JSONArrayVariation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want) {
				t.Errorf("JSONArrayVariation() got = %v, want %v", got, tt.want)
			}

			if tt.expectedLog != "" {
				content, _ := ioutil.ReadFile(file.Name())
				assert.Regexp(t, tt.expectedLog, string(content))
			}
			// clean logger
			ff = nil
			file.Close()
		})
	}
}

func TestJSONVariation(t *testing.T) {
	type args struct {
		flagKey      string
		user         ffuser.User
		defaultValue map[string]interface{}
		cacheMock    cache.Cache
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
				cacheMock: NewCacheMock(model.Flag{
					Disable: true,
				}, nil),
			},
			want:        map[string]interface{}{"default-notkey": true},
			wantErr:     true,
			expectedLog: "\\[" + testutil.RFC3339Regex + "\\] user=\"random-key\", flag=\"disable-flag\", value=\"map\\[default-notkey:true\\]\"\n",
		},
		{
			name: "Get error when not init",
			args: args{
				flagKey:      "key-not-exist",
				user:         ffuser.NewUser("random-key"),
				defaultValue: map[string]interface{}{"default-notkey": true},
				cacheMock: NewCacheMock(
					model.Flag{},
					errors.New("impossible to read the toggle before the initialisation")),
			},
			want:        map[string]interface{}{"default-notkey": true},
			wantErr:     true,
			expectedLog: "\\[" + testutil.RFC3339Regex + "\\] user=\"random-key\", flag=\"key-not-exist\", value=\"map\\[default-notkey:true\\]\"\n",
		},
		{
			name: "Get default value with key not exist",
			args: args{
				flagKey:      "key-not-exist",
				user:         ffuser.NewUser("random-key"),
				defaultValue: map[string]interface{}{"default-notkey": true},
				cacheMock:    NewCacheMock(model.Flag{}, errors.New("flag [key-not-exist] does not exists")),
			},
			want:        map[string]interface{}{"default-notkey": true},
			wantErr:     true,
			expectedLog: "\\[" + testutil.RFC3339Regex + "\\] user=\"random-key\", flag=\"key-not-exist\", value=\"map\\[default-notkey:true\\]\"\n",
		},
		{
			name: "Get default value, rule not apply",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewUser("random-key"),
				defaultValue: map[string]interface{}{"default-notkey": true},
				cacheMock: NewCacheMock(model.Flag{
					Rule:       "key eq \"key\"",
					Percentage: 100,
					Default:    map[string]interface{}{"default": true},
					True:       map[string]interface{}{"true": true},
					False:      map[string]interface{}{"false": true},
				}, nil),
			},
			want:        map[string]interface{}{"default": true},
			wantErr:     false,
			expectedLog: "\\[" + testutil.RFC3339Regex + "\\] user=\"random-key\", flag=\"test-flag\", value=\"map\\[default:true\\]\"\n",
		},
		{
			name: "Get true value, rule apply",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewAnonymousUser("random-key"),
				defaultValue: map[string]interface{}{"default-notkey": true},
				cacheMock: NewCacheMock(model.Flag{
					Rule:       "key eq \"random-key\"",
					Percentage: 100,
					Default:    map[string]interface{}{"default": true},
					True:       map[string]interface{}{"true": true},
					False:      map[string]interface{}{"false": true},
				}, nil),
			},
			want:        map[string]interface{}{"true": true},
			wantErr:     false,
			expectedLog: "\\[" + testutil.RFC3339Regex + "\\] user=\"random-key\", flag=\"test-flag\", value=\"map\\[true:true\\]\"\n",
		},
		{
			name: "Get false value, rule apply",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewAnonymousUser("random-key-ssss1"),
				defaultValue: map[string]interface{}{"default-notkey": true},
				cacheMock: NewCacheMock(model.Flag{
					Rule:       "anonymous eq true",
					Percentage: 50,
					Default:    map[string]interface{}{"default": true},
					True:       map[string]interface{}{"true": true},
					False:      map[string]interface{}{"false": true},
				}, nil),
			},
			want:        map[string]interface{}{"false": true},
			wantErr:     false,
			expectedLog: "\\[" + testutil.RFC3339Regex + "\\] user=\"random-key-ssss1\", flag=\"test-flag\", value=\"map\\[false:true\\]\"\n",
		},
		{
			name: "Get default value, when rule apply and not right type",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewUser("random-key-ssss1"),
				defaultValue: map[string]interface{}{"default-notkey": true},
				cacheMock: NewCacheMock(model.Flag{
					Percentage: 100,
					Default:    "xxx",
					True:       "xxx",
					False:      "xxx",
				}, nil),
			},
			want:        map[string]interface{}{"default-notkey": true},
			wantErr:     true,
			expectedLog: "\\[" + testutil.RFC3339Regex + "\\] user=\"random-key-ssss1\", flag=\"test-flag\", value=\"map\\[default-notkey:true\\]\"\n",
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
					PollInterval: 0,
					Logger:       logger,
				},
			}

			got, err := JSONVariation(tt.args.flagKey, tt.args.user, tt.args.defaultValue)

			if (err != nil) != tt.wantErr {
				t.Errorf("JSONVariation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want) {
				t.Errorf("JSONVariation() got = %v, want %v", got, tt.want)
			}

			if tt.expectedLog != "" {
				content, _ := ioutil.ReadFile(file.Name())
				assert.Regexp(t, tt.expectedLog, string(content))
			}
			// clean logger
			ff = nil
			file.Close()
		})
	}
}

func TestStringVariation(t *testing.T) {
	type args struct {
		flagKey      string
		user         ffuser.User
		defaultValue string
		cacheMock    cache.Cache
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
				cacheMock: NewCacheMock(model.Flag{
					Disable: true,
				}, nil),
			},
			want:        "default-notkey",
			wantErr:     true,
			expectedLog: "\\[" + testutil.RFC3339Regex + "\\] user=\"random-key\", flag=\"disable-flag\", value=\"default-notkey\"\n",
		},
		{
			name: "Get error when not init",
			args: args{
				flagKey:      "key-not-exist",
				user:         ffuser.NewUser("random-key"),
				defaultValue: "default-notkey",
				cacheMock: NewCacheMock(
					model.Flag{},
					errors.New("impossible to read the toggle before the initialisation")),
			},
			want:        "default-notkey",
			wantErr:     true,
			expectedLog: "\\[" + testutil.RFC3339Regex + "\\] user=\"random-key\", flag=\"key-not-exist\", value=\"default-notkey\"\n",
		},
		{
			name: "Get default value with key not exist",
			args: args{
				flagKey:      "key-not-exist",
				user:         ffuser.NewUser("random-key"),
				defaultValue: "default-notkey",
				cacheMock:    NewCacheMock(model.Flag{}, errors.New("flag [key-not-exist] does not exists")),
			},
			want:        "default-notkey",
			wantErr:     true,
			expectedLog: "\\[" + testutil.RFC3339Regex + "\\] user=\"random-key\", flag=\"key-not-exist\", value=\"default-notkey\"\n",
		},

		{
			name: "Get default value, rule not apply",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewUser("random-key"),
				defaultValue: "default-notkey",
				cacheMock: NewCacheMock(model.Flag{
					Rule:       "key eq \"key\"",
					Percentage: 100,
					Default:    "default",
					True:       "true",
					False:      "false",
				}, nil),
			},
			want:        "default",
			wantErr:     false,
			expectedLog: "\\[" + testutil.RFC3339Regex + "\\] user=\"random-key\", flag=\"test-flag\", value=\"default\"\n",
		},
		{
			name: "Get true value, rule apply",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewAnonymousUser("random-key"),
				defaultValue: "default-notkey",
				cacheMock: NewCacheMock(model.Flag{
					Rule:       "key eq \"random-key\"",
					Percentage: 100,
					Default:    "default",
					True:       "true",
					False:      "false",
				}, nil),
			},
			want:        "true",
			wantErr:     false,
			expectedLog: "\\[" + testutil.RFC3339Regex + "\\] user=\"random-key\", flag=\"test-flag\", value=\"true\"\n",
		},
		{
			name: "Get false value, rule apply",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewAnonymousUser("random-key-ssss1"),
				defaultValue: "default-notkey",
				cacheMock: NewCacheMock(model.Flag{
					Rule:       "anonymous eq true",
					Percentage: 50,
					Default:    "default",
					True:       "true",
					False:      "false",
				}, nil),
			},
			want:        "false",
			wantErr:     false,
			expectedLog: "\\[" + testutil.RFC3339Regex + "\\] user=\"random-key-ssss1\", flag=\"test-flag\", value=\"false\"\n",
		},
		{
			name: "Get default value, when rule apply and not right type",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewUser("random-key-ssss1"),
				defaultValue: "default-notkey",
				cacheMock: NewCacheMock(model.Flag{
					Rule:       "anonymous eq true",
					Percentage: 50,
					Default:    111,
					True:       112,
					False:      113,
				}, nil),
			},
			want:        "default-notkey",
			wantErr:     true,
			expectedLog: "\\[" + testutil.RFC3339Regex + "\\] user=\"random-key-ssss1\", flag=\"test-flag\", value=\"default-notkey\"\n",
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
					PollInterval: 0,
					Logger:       logger,
				},
			}
			got, err := StringVariation(tt.args.flagKey, tt.args.user, tt.args.defaultValue)

			if (err != nil) != tt.wantErr {
				t.Errorf("StringVariation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("StringVariation() got = %v, want %v", got, tt.want)
			}

			if tt.expectedLog != "" {
				content, _ := ioutil.ReadFile(file.Name())
				assert.Regexp(t, tt.expectedLog, string(content))
			}
			// clean logger
			ff = nil
			file.Close()
		})
	}
}

func TestIntVariation(t *testing.T) {
	type args struct {
		flagKey      string
		user         ffuser.User
		defaultValue int
		cacheMock    cache.Cache
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
				cacheMock: NewCacheMock(model.Flag{
					Disable: true,
				}, nil),
			},
			want:        125,
			wantErr:     true,
			expectedLog: "\\[" + testutil.RFC3339Regex + "\\] user=\"random-key\", flag=\"disable-flag\", value=\"125\"\n",
		},
		{
			name: "Get error when not init",
			args: args{
				flagKey:      "key-not-exist",
				user:         ffuser.NewUser("random-key"),
				defaultValue: 118,
				cacheMock: NewCacheMock(
					model.Flag{},
					errors.New("impossible to read the toggle before the initialisation")),
			},
			want:        118,
			wantErr:     true,
			expectedLog: "\\[" + testutil.RFC3339Regex + "\\] user=\"random-key\", flag=\"key-not-exist\", value=\"118\"\n",
		},
		{
			name: "Get default value with key not exist",
			args: args{
				flagKey:      "key-not-exist",
				user:         ffuser.NewUser("random-key"),
				defaultValue: 118,
				cacheMock:    NewCacheMock(model.Flag{}, errors.New("flag [key-not-exist] does not exists")),
			},
			want:        118,
			wantErr:     true,
			expectedLog: "\\[" + testutil.RFC3339Regex + "\\] user=\"random-key\", flag=\"key-not-exist\", value=\"118\"\n",
		},
		{
			name: "Get default value rule not apply",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewUser("random-key"),
				defaultValue: 118,
				cacheMock: NewCacheMock(model.Flag{
					Rule:       "key eq \"key\"",
					Percentage: 100,
					Default:    119,
					True:       120,
					False:      121,
				}, nil),
			},
			want:        119,
			wantErr:     false,
			expectedLog: "\\[" + testutil.RFC3339Regex + "\\] user=\"random-key\", flag=\"test-flag\", value=\"119\"\n",
		},
		{
			name: "Get true value, rule apply",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewAnonymousUser("random-key"),
				defaultValue: 118,
				cacheMock: NewCacheMock(model.Flag{
					Rule:       "key eq \"random-key\"",
					Percentage: 100,
					Default:    119,
					True:       120,
					False:      121,
				}, nil),
			},
			want:        120,
			wantErr:     false,
			expectedLog: "\\[" + testutil.RFC3339Regex + "\\] user=\"random-key\", flag=\"test-flag\", value=\"120\"\n",
		},
		{
			name: "Get false value, rule apply",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewAnonymousUser("random-key-ssss1"),
				defaultValue: 118,
				cacheMock: NewCacheMock(model.Flag{
					Rule:       "anonymous eq true",
					Percentage: 50,
					Default:    119,
					True:       120,
					False:      121,
				}, nil),
			},
			want:        121,
			wantErr:     false,
			expectedLog: "\\[" + testutil.RFC3339Regex + "\\] user=\"random-key-ssss1\", flag=\"test-flag\", value=\"121\"\n",
		},
		{
			name: "Get default value, when rule apply and not right type",
			args: args{
				flagKey:      "test-flag",
				user:         ffuser.NewUser("random-key-ssss1"),
				defaultValue: 118,
				cacheMock: NewCacheMock(model.Flag{
					Rule:       "anonymous eq true",
					Percentage: 50,
					Default:    "default",
					True:       "true",
					False:      "false",
				}, nil),
			},
			want:        118,
			wantErr:     true,
			expectedLog: "\\[" + testutil.RFC3339Regex + "\\] user=\"random-key-ssss1\", flag=\"test-flag\", value=\"118\"\n",
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
					PollInterval: 0,
					Logger:       logger,
				},
			}
			got, err := IntVariation(tt.args.flagKey, tt.args.user, tt.args.defaultValue)

			if (err != nil) != tt.wantErr {
				t.Errorf("IntVariation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IntVariation() got = %v, want %v", got, tt.want)
			}

			if tt.expectedLog != "" {
				content, _ := ioutil.ReadFile(file.Name())
				assert.Regexp(t, tt.expectedLog, string(content))
			}
			// clean logger
			ff = nil
			file.Close()
		})
	}
}

package flagv1_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	flagv1 "github.com/thomaspoignant/go-feature-flag/internal/flagv1"
	"testing"
	"time"

	"github.com/thomaspoignant/go-feature-flag/ffuser"
	"github.com/thomaspoignant/go-feature-flag/testutils/testconvert"
)

func TestFlag_value(t *testing.T) {
	type fields struct {
		Disable    bool
		Rule       string
		Percentage float64
		True       interface{}
		False      interface{}
		Default    interface{}
		Rollout    flagv1.Rollout
	}
	type args struct {
		flagName string
		user     ffuser.User
	}
	type want struct {
		value         interface{}
		variationType flagv1.VariationType
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "Rule disable get default value",
			fields: fields{
				Disable: true,
				True:    "true",
				False:   "false",
				Default: "default",
			},
			args: args{
				flagName: "test_689483",
				user:     ffuser.NewUser("test_689483"),
			},
			want: want{
				value:         "default",
				variationType: flagv1.VariationDefault,
			},
		},
		{
			name: "Get true value if rule pass",
			fields: fields{
				True:       "true",
				False:      "false",
				Default:    "default",
				Rule:       "key == \"7e50ee61-06ad-4bb0-9034-38ad7cdea9f5\"",
				Percentage: 10,
			},
			args: args{
				flagName: "test-flag",
				user:     ffuser.NewUserBuilder("7e50ee61-06ad-4bb0-9034-38ad7cdea9f5").AddCustom("name", "john").Build(), // combined hash is 9
			},
			want: want{
				value:         "true",
				variationType: flagv1.VariationTrue,
			},
		},
		{
			name: "Rollout Experimentation only start date in the past",
			fields: fields{
				True:       "true",
				False:      "false",
				Default:    "default",
				Rule:       "key == \"7e50ee61-06ad-4bb0-9034-38ad7cdea9f5\"",
				Percentage: 10,
				Rollout: flagv1.Rollout{
					Experimentation: &flagv1.Experimentation{
						Start: testconvert.Time(time.Now().Add(-1 * time.Minute)),
						End:   nil,
					},
				},
			},
			args: args{
				flagName: "test-flag",
				user:     ffuser.NewUserBuilder("7e50ee61-06ad-4bb0-9034-38ad7cdea9f5").AddCustom("name", "john").Build(),
			},
			want: want{
				value:         "true",
				variationType: flagv1.VariationTrue,
			},
		},
		{
			name: "Rollout Experimentation only start date in the future",
			fields: fields{
				True:       "true",
				False:      "false",
				Default:    "default",
				Rule:       "key == \"user66\"",
				Percentage: 10,
				Rollout: flagv1.Rollout{
					Experimentation: &flagv1.Experimentation{
						Start: testconvert.Time(time.Now().Add(1 * time.Minute)),
						End:   nil,
					},
				},
			},
			args: args{
				flagName: "test-flag",
				user:     ffuser.NewUserBuilder("user66").AddCustom("name", "john").Build(), // combined hash is 9
			},
			want: want{
				value:         "default",
				variationType: flagv1.VariationDefault,
			},
		},
		{
			name: "Rollout Experimentation between start and end date",
			fields: fields{
				True:       "true",
				False:      "false",
				Default:    "default",
				Rule:       "key == \"7e50ee61-06ad-4bb0-9034-38ad7cdea9f5\"",
				Percentage: 10,
				Rollout: flagv1.Rollout{
					Experimentation: &flagv1.Experimentation{
						Start: testconvert.Time(time.Now().Add(-1 * time.Minute)),
						End:   testconvert.Time(time.Now().Add(1 * time.Minute)),
					},
				},
			},
			args: args{
				flagName: "test-flag",
				user:     ffuser.NewUserBuilder("7e50ee61-06ad-4bb0-9034-38ad7cdea9f5").AddCustom("name", "john").Build(),
			},
			want: want{
				value:         "true",
				variationType: flagv1.VariationTrue,
			},
		},
		{
			name: "Rollout Experimentation not started yet",
			fields: fields{
				True:       "true",
				False:      "false",
				Default:    "default",
				Rule:       "key == \"user66\"",
				Percentage: 10,
				Rollout: flagv1.Rollout{
					Experimentation: &flagv1.Experimentation{
						Start: testconvert.Time(time.Now().Add(1 * time.Minute)),
						End:   testconvert.Time(time.Now().Add(2 * time.Minute)),
					},
				},
			},
			args: args{
				flagName: "test-flag",
				user:     ffuser.NewUserBuilder("user66").AddCustom("name", "john").Build(), // combined hash is 9
			},
			want: want{
				value:         "default",
				variationType: flagv1.VariationDefault,
			},
		},
		{
			name: "Rollout Experimentation finished",
			fields: fields{
				True:       "true",
				False:      "false",
				Default:    "default",
				Rule:       "key == \"user66\"",
				Percentage: 10,
				Rollout: flagv1.Rollout{
					Experimentation: &flagv1.Experimentation{
						Start: testconvert.Time(time.Now().Add(-2 * time.Minute)),
						End:   testconvert.Time(time.Now().Add(-1 * time.Minute)),
					},
				},
			},
			args: args{
				flagName: "test-flag",
				user:     ffuser.NewUserBuilder("user66").AddCustom("name", "john").Build(), // combined hash is 9
			},
			want: want{
				value:         "default",
				variationType: flagv1.VariationDefault,
			},
		},
		{
			name: "Rollout Experimentation only end date finished",
			fields: fields{
				True:       "true",
				False:      "false",
				Default:    "default",
				Rule:       "key == \"user66\"",
				Percentage: 10,
				Rollout: flagv1.Rollout{
					Experimentation: &flagv1.Experimentation{
						Start: nil,
						End:   testconvert.Time(time.Now().Add(-1 * time.Minute)),
					},
				},
			},
			args: args{
				flagName: "test-flag",
				user:     ffuser.NewUserBuilder("user66").AddCustom("name", "john").Build(), // combined hash is 9
			},
			want: want{
				value:         "default",
				variationType: flagv1.VariationDefault,
			},
		},
		{
			name: "Rollout Experimentation only end date not finished",
			fields: fields{
				True:       "true",
				False:      "false",
				Default:    "default",
				Rule:       "key == \"7e50ee61-06ad-4bb0-9034-38ad7cdea9f5\"",
				Percentage: 10,
				Rollout: flagv1.Rollout{
					Experimentation: &flagv1.Experimentation{
						Start: nil,
						End:   testconvert.Time(time.Now().Add(1 * time.Minute)),
					},
				},
			},
			args: args{
				flagName: "test-flag",
				user:     ffuser.NewUserBuilder("7e50ee61-06ad-4bb0-9034-38ad7cdea9f5").AddCustom("name", "john").Build(),
			},
			want: want{
				value:         "true",
				variationType: flagv1.VariationTrue,
			},
		},
		{
			name: "Rollout Experimentation only end date not finished",
			fields: fields{
				True:       "true",
				False:      "false",
				Default:    "default",
				Rule:       "key == \"7e50ee61-06ad-4bb0-9034-38ad7cdea9f5\"",
				Percentage: 10,
				Rollout: flagv1.Rollout{
					Experimentation: &flagv1.Experimentation{
						Start: nil,
						End:   nil,
					},
				},
			},
			args: args{
				flagName: "test-flag",
				user:     ffuser.NewUserBuilder("7e50ee61-06ad-4bb0-9034-38ad7cdea9f5").AddCustom("name", "john").Build(),
			},
			want: want{
				value:         "true",
				variationType: flagv1.VariationTrue,
			},
		},
		{
			name: "Invert start date and end date",
			fields: fields{
				True:       "true",
				False:      "false",
				Default:    "default",
				Rule:       "key == \"user66\"",
				Percentage: 10,
				Rollout: flagv1.Rollout{
					Experimentation: &flagv1.Experimentation{
						Start: testconvert.Time(time.Now().Add(1 * time.Minute)),
						End:   testconvert.Time(time.Now().Add(-1 * time.Minute)),
					},
				},
			},
			args: args{
				flagName: "test-flag",
				user:     ffuser.NewUserBuilder("user66").AddCustom("name", "john").Build(), // combined hash is 9
			},
			want: want{
				value:         "default",
				variationType: flagv1.VariationDefault,
			},
		},
		{
			name: "Get default value if does not pass",
			fields: fields{
				True:       "true",
				False:      "false",
				Default:    "default",
				Rule:       "key == \"7e50ee61-06ad-4bb0-9034-38ad7\"",
				Percentage: 10,
			},
			args: args{
				flagName: "test-flag",
				user:     ffuser.NewUserBuilder("7e50ee61-06ad-4bb0-9034-38ad7cdea9f5").AddCustom("name", "john").Build(),
			},
			want: want{
				value:         "default",
				variationType: flagv1.VariationDefault,
			},
		},
		{
			name: "Get false value if rule pass and not in the cohort",
			fields: fields{
				True:       "true",
				False:      "false",
				Default:    "default",
				Rule:       "key == \"7e50ee61-06ad-4bb0-9034-38ad7cdea9f5\"",
				Percentage: 10,
			},
			args: args{
				flagName: "test-flag2",
				user:     ffuser.NewUserBuilder("7e50ee61-06ad-4bb0-9034-38ad7cdea9f5").AddCustom("name", "john").Build(),
			},
			want: want{
				value:         "false",
				variationType: flagv1.VariationFalse,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &flagv1.FlagData{
				Disable:    testconvert.Bool(tt.fields.Disable),
				Rule:       testconvert.String(tt.fields.Rule),
				Percentage: testconvert.Float64(tt.fields.Percentage),
				True:       testconvert.Interface(tt.fields.True),
				False:      testconvert.Interface(tt.fields.False),
				Default:    testconvert.Interface(tt.fields.Default),
				Rollout:    &tt.fields.Rollout,
			}

			got, variationType := f.Value(tt.args.flagName, tt.args.user)
			assert.Equal(t, tt.want.value, got)
			assert.Equal(t, tt.want.variationType, variationType)
		})
	}
}

func TestFlag_ProgressiveRollout(t *testing.T) {
	f := &flagv1.FlagData{
		Percentage: testconvert.Float64(0),
		True:       testconvert.Interface("True"),
		False:      testconvert.Interface("False"),
		Default:    testconvert.Interface("Default"),
		Rollout: &flagv1.Rollout{Progressive: &flagv1.Progressive{
			ReleaseRamp: flagv1.ProgressiveReleaseRamp{
				Start: testconvert.Time(time.Now().Add(1 * time.Second)),
				End:   testconvert.Time(time.Now().Add(2 * time.Second)),
			},
		}},
	}

	user := ffuser.NewAnonymousUser("test")
	flagName := "test-flag"

	// We evaluate the same flag multiple time overtime.
	v, _ := f.Value(flagName, user)
	assert.Equal(t, f.GetVariationValue(flagv1.VariationFalse), v)

	time.Sleep(1 * time.Second)
	v2, _ := f.Value(flagName, user)
	assert.Equal(t, f.GetVariationValue(flagv1.VariationFalse), v2)

	time.Sleep(1 * time.Second)
	v3, _ := f.Value(flagName, user)
	assert.Equal(t, f.GetVariationValue(flagv1.VariationTrue), v3)
}

func TestFlag_ScheduledRollout(t *testing.T) {
	f := &flagv1.FlagData{
		Rule:       testconvert.String("key eq \"test\""),
		Percentage: testconvert.Float64(0),
		True:       testconvert.Interface("True"),
		False:      testconvert.Interface("False"),
		Default:    testconvert.Interface("Default"),
		Rollout: &flagv1.Rollout{
			Scheduled: &flagv1.ScheduledRollout{
				Steps: []flagv1.ScheduledStep{
					{
						FlagData: flagv1.FlagData{
							Version: testconvert.Float64(1.1),
						},
						Date: testconvert.Time(time.Now().Add(1 * time.Second)),
					},
					{
						FlagData: flagv1.FlagData{
							Percentage: testconvert.Float64(100),
						},
						Date: testconvert.Time(time.Now().Add(1 * time.Second)),
					},
					{
						FlagData: flagv1.FlagData{
							True:    testconvert.Interface("True2"),
							False:   testconvert.Interface("False2"),
							Default: testconvert.Interface("Default2"),
							Rule:    testconvert.String("key eq \"test2\""),
						},
						Date: testconvert.Time(time.Now().Add(2 * time.Second)),
					},
					{
						FlagData: flagv1.FlagData{
							True:    testconvert.Interface("True2"),
							False:   testconvert.Interface("False2"),
							Default: testconvert.Interface("Default2"),
							Rule:    testconvert.String("key eq \"test\""),
						},
						Date: testconvert.Time(time.Now().Add(3 * time.Second)),
					},
					{
						FlagData: flagv1.FlagData{
							Disable: testconvert.Bool(true),
						},
						Date: testconvert.Time(time.Now().Add(4 * time.Second)),
					},
					{
						FlagData: flagv1.FlagData{
							Percentage: testconvert.Float64(0),
						},
					},
					{
						FlagData: flagv1.FlagData{
							Disable:     testconvert.Bool(false),
							TrackEvents: testconvert.Bool(true),
							Rollout: &flagv1.Rollout{
								Experimentation: &flagv1.Experimentation{
									Start: testconvert.Time(time.Now().Add(6 * time.Second)),
									End:   testconvert.Time(time.Now().Add(7 * time.Second)),
								},
							},
						},
						Date: testconvert.Time(time.Now().Add(5 * time.Second)),
					},
				},
			},
		},
	}

	user := ffuser.NewAnonymousUser("test")
	flagName := "test-flag"

	// We evaluate the same flag multiple time overtime.
	v, _ := f.Value(flagName, user)
	assert.Equal(t, f.GetVariationValue(flagv1.VariationFalse), v)

	time.Sleep(1 * time.Second)

	v, _ = f.Value(flagName, user)
	assert.Equal(t, "True", v)
	assert.Equal(t, 1.1, f.GetVersion())

	time.Sleep(1 * time.Second)

	v, _ = f.Value(flagName, user)
	assert.Equal(t, "Default2", v)

	time.Sleep(1 * time.Second)

	v, _ = f.Value(flagName, user)
	assert.Equal(t, "True2", v)

	time.Sleep(1 * time.Second)

	v, _ = f.Value(flagName, user)
	assert.Equal(t, "Default2", v)

	time.Sleep(1 * time.Second)

	v, _ = f.Value(flagName, user)
	assert.Equal(t, "Default2", v)

	time.Sleep(1 * time.Second)

	v, _ = f.Value(flagName, user)
	assert.Equal(t, "True2", v)

	time.Sleep(1 * time.Second)

	v, _ = f.Value(flagName, user)
	assert.Equal(t, "Default2", v)
}

func TestFlag_String(t *testing.T) {
	type fields struct {
		Disable     bool
		Rule        string
		Percentage  float64
		True        interface{}
		False       interface{}
		Default     interface{}
		TrackEvents *bool
		Version     *float64
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "All fields",
			fields: fields{
				Disable:     false,
				Rule:        "key eq \"toto\"",
				Percentage:  10,
				True:        true,
				False:       false,
				Default:     false,
				TrackEvents: testconvert.Bool(true),
				Version:     testconvert.Float64(12),
			},
			want: "percentage=10%, rule=\"key eq \"toto\"\", true=\"true\", false=\"false\", default=\"false\", disable=\"false\", trackEvents=\"true\", version=12",
		},
		{
			name: "No rule",
			fields: fields{
				Disable:    false,
				Percentage: 10,
				True:       true,
				False:      false,
				Default:    false,
			},
			want: "percentage=10%, true=\"true\", false=\"false\", default=\"false\", disable=\"false\"",
		},
		{
			name: "Default values",
			fields: fields{
				True:    true,
				False:   false,
				Default: false,
			},
			want: "percentage=0%, true=\"true\", false=\"false\", default=\"false\", disable=\"false\"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &flagv1.FlagData{
				Disable:     testconvert.Bool(tt.fields.Disable),
				Rule:        testconvert.String(tt.fields.Rule),
				Percentage:  testconvert.Float64(tt.fields.Percentage),
				True:        testconvert.Interface(tt.fields.True),
				False:       testconvert.Interface(tt.fields.False),
				Default:     testconvert.Interface(tt.fields.Default),
				TrackEvents: tt.fields.TrackEvents,
				Version:     tt.fields.Version,
			}
			got := f.String()
			assert.Equal(t, tt.want, got, "String() = %v, want %v", got, tt.want)
		})
	}
}

func TestFlag_Getter(t *testing.T) {
	type expected struct {
		True        interface{}
		False       interface{}
		Default     interface{}
		Rollout     *flagv1.Rollout
		Disable     bool
		TrackEvents bool
		Percentage  float64
		Rule        string
		Version     float64
		RawValues   map[string]string
	}
	tests := []struct {
		name string
		flag flag.Flag
		want expected
	}{
		{
			name: "all default",
			flag: &flagv1.FlagData{},
			want: expected{
				True:        nil,
				False:       nil,
				Default:     nil,
				Rollout:     nil,
				Disable:     false,
				TrackEvents: true,
				Percentage:  0,
				Rule:        "",
				Version:     0,
				RawValues: map[string]string{
					"Default":     "",
					"Disable":     "false",
					"False":       "",
					"Percentage":  "0.00",
					"Rollout":     "",
					"Rule":        "",
					"TrackEvents": "true",
					"True":        "",
					"Version":     "0",
				},
			},
		},
		{
			name: "custom flag",
			flag: &flagv1.FlagData{
				Rule:        testconvert.String("test"),
				Percentage:  testconvert.Float64(90),
				True:        testconvert.Interface(12.2),
				False:       testconvert.Interface(13.2),
				Default:     testconvert.Interface(14.2),
				TrackEvents: testconvert.Bool(false),
				Disable:     testconvert.Bool(true),
				Version:     testconvert.Float64(127),
			},
			want: expected{
				True:        12.2,
				False:       13.2,
				Default:     14.2,
				Disable:     true,
				TrackEvents: false,
				Percentage:  90,
				Rule:        "test",
				Version:     127,
				RawValues: map[string]string{
					"Default":     "14.2",
					"Disable":     "true",
					"False":       "13.2",
					"Percentage":  "90.00",
					"Rollout":     "",
					"Rule":        "test",
					"TrackEvents": "false",
					"True":        "12.2",
					"Version":     "127",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want.Disable, tt.flag.GetDisable())
			assert.Equal(t, tt.want.TrackEvents, tt.flag.GetTrackEvents())
			assert.Equal(t, tt.want.Version, tt.flag.GetVersion())
			assert.Equal(t, flagv1.VariationDefault, tt.flag.GetDefaultVariation())
			fmt.Println(tt.want.Default, tt.flag.GetVariationValue(tt.flag.GetDefaultVariation()))
			assert.Equal(t, tt.want.Default, tt.flag.GetVariationValue(tt.flag.GetDefaultVariation()))
			assert.Equal(t, tt.want.RawValues, tt.flag.GetRawValues())
		})
	}
}

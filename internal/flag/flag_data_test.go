package flag_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/internal/constant"
	flag "github.com/thomaspoignant/go-feature-flag/internal/flag"
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
		Rollout    flag.DtoRollout
	}
	type args struct {
		flagName string
		user     ffuser.User
	}
	type want struct {
		value         interface{}
		variationType string
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
				variationType: constant.VariationSDKDefault,
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
				variationType: "True",
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
				Rollout: flag.DtoRollout{
					Experimentation: &flag.Experimentation{
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
				variationType: "True",
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
				Rollout: flag.DtoRollout{
					Experimentation: &flag.Experimentation{
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
				variationType: constant.VariationSDKDefault,
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
				Rollout: flag.DtoRollout{
					Experimentation: &flag.Experimentation{
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
				variationType: "True",
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
				Rollout: flag.DtoRollout{
					Experimentation: &flag.Experimentation{
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
				variationType: constant.VariationSDKDefault,
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
				Rollout: flag.DtoRollout{
					Experimentation: &flag.Experimentation{
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
				variationType: constant.VariationSDKDefault,
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
				Rollout: flag.DtoRollout{
					Experimentation: &flag.Experimentation{
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
				variationType: constant.VariationSDKDefault,
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
				Rollout: flag.DtoRollout{
					Experimentation: &flag.Experimentation{
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
				variationType: "True",
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
				Rollout: flag.DtoRollout{
					Experimentation: &flag.Experimentation{
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
				variationType: "True",
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
				Rollout: flag.DtoRollout{
					Experimentation: &flag.Experimentation{
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
				variationType: constant.VariationSDKDefault,
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
				variationType: flag.VariationDefault,
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
				variationType: "False",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dto := &flag.DtoFlag{
				Disable:    testconvert.Bool(tt.fields.Disable),
				Rule:       testconvert.String(tt.fields.Rule),
				Percentage: testconvert.Float64(tt.fields.Percentage),
				True:       testconvert.Interface(tt.fields.True),
				False:      testconvert.Interface(tt.fields.False),
				Default:    testconvert.Interface(tt.fields.Default),
				Rollout:    &tt.fields.Rollout,
			}

			f, err := dto.ConvertToFlagData(false)
			assert.NoError(t, err)

			got, variationType, _ := f.Value(tt.args.flagName, tt.args.user, tt.fields.Default)
			assert.Equal(t, tt.want.value, got)
			assert.Equal(t, tt.want.variationType, variationType)
		})
	}
}

func TestFlag_ProgressiveRollout(t *testing.T) {
	dto := &flag.DtoFlag{
		Percentage: testconvert.Float64(0),
		True:       testconvert.Interface("True"),
		False:      testconvert.Interface("False"),
		Default:    testconvert.Interface("Default"),
		Rollout: &flag.DtoRollout{Progressive: &flag.Progressive{
			ReleaseRamp: flag.ProgressiveReleaseRamp{
				Start: testconvert.Time(time.Now().Add(1 * time.Second)),
				End:   testconvert.Time(time.Now().Add(2 * time.Second)),
			},
		}},
	}

	f, err := dto.ConvertToFlagData(false)
	assert.NoError(t, err)

	user := ffuser.NewAnonymousUser("test")
	flagName := "test-flag"

	// We evaluate the same flag multiple time overtime.
	v, _, _ := f.Value(flagName, user, "sdkdefault")
	assert.Equal(t, f.GetVariationValue("False"), v)

	time.Sleep(1 * time.Second)
	v2, _, _ := f.Value(flagName, user, "sdkdefault")
	assert.Equal(t, f.GetVariationValue("False"), v2)

	time.Sleep(1 * time.Second)
	v3, _, _ := f.Value(flagName, user, "sdkdefault")
	assert.Equal(t, f.GetVariationValue("True"), v3)
}

func TestFlag_ScheduledRollout(t *testing.T) {
	dto := &flag.DtoFlag{
		Rule:       testconvert.String("key eq \"test\""),
		Percentage: testconvert.Float64(0),
		True:       testconvert.Interface("True"),
		False:      testconvert.Interface("False"),
		Default:    testconvert.Interface("Default"),
		Rollout: &flag.DtoRollout{
			Scheduled: &flag.DtoScheduledRollout{
				Steps: []flag.DtoScheduledStep{
					{
						DtoFlag: flag.DtoFlag{
							Version: testconvert.String(fmt.Sprintf("%.2f", 1.1)),
						},
						Date: testconvert.Time(time.Now().Add(1 * time.Second)),
					},
					{
						DtoFlag: flag.DtoFlag{
							Percentage: testconvert.Float64(100),
						},
						Date: testconvert.Time(time.Now().Add(1 * time.Second)),
					},
					{
						DtoFlag: flag.DtoFlag{
							True:    testconvert.Interface("True2"),
							False:   testconvert.Interface("False2"),
							Default: testconvert.Interface("Default2"),
							Rule:    testconvert.String("key eq \"test2\""),
						},
						Date: testconvert.Time(time.Now().Add(2 * time.Second)),
					},
					{
						DtoFlag: flag.DtoFlag{
							True:    testconvert.Interface("True2"),
							False:   testconvert.Interface("False2"),
							Default: testconvert.Interface("Default2"),
							Rule:    testconvert.String("key eq \"test\""),
						},
						Date: testconvert.Time(time.Now().Add(3 * time.Second)),
					},
					{
						DtoFlag: flag.DtoFlag{
							Disable: testconvert.Bool(true),
						},
						Date: testconvert.Time(time.Now().Add(4 * time.Second)),
					},
					{
						DtoFlag: flag.DtoFlag{
							Disable: testconvert.Bool(false),
						},
					},
					{
						DtoFlag: flag.DtoFlag{
							Disable:     testconvert.Bool(false),
							TrackEvents: testconvert.Bool(true),
							Rollout: &flag.DtoRollout{
								Experimentation: &flag.Experimentation{
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

	f, err := dto.ConvertToFlagData(false)
	assert.NoError(t, err)

	user := ffuser.NewAnonymousUser("test")
	flagName := "test-flag"

	// We evaluate the same flag multiple time overtime.
	v, _ := f.Value(flagName, user, "sdkdefault")
	assert.Equal(t, f.GetVariationValue("False"), v)

	time.Sleep(1 * time.Second)

	// Change the version of the flag + rollout the flag to 100% of the users with the filter
	v, _ = f.Value(flagName, user, "sdkdefault")
	assert.Equal(t, "True", v)
	assert.Equal(t, "1.10", f.GetVersion())

	time.Sleep(1 * time.Second)

	// Change the query to unmatch user + value of variations
	v, _ = f.Value(flagName, user, "sdkdefault")
	assert.Equal(t, "Default2", v)

	time.Sleep(1 * time.Second)

	// Change the query to match user + value of variations
	v, _ = f.Value(flagName, user, "sdkdefault")
	assert.Equal(t, "True2", v)

	time.Sleep(1 * time.Second)

	// Disable the flag
	v, _ = f.Value(flagName, user, "sdkdefault")
	assert.Equal(t, "sdkdefault", v)

	time.Sleep(1 * time.Second)

	// Enable flag without date (should be ignored)
	v, _ = f.Value(flagName, user, "sdkdefault")
	assert.Equal(t, "sdkdefault", v)

	time.Sleep(1 * time.Second)

	// enable flag + add progressive rollout
	v, _ = f.Value(flagName, user, "sdkdefault")
	assert.Equal(t, "True2", v)

	time.Sleep(1 * time.Second)

	// experimentation should be finished so we serve default value
	v, _ = f.Value(flagName, user, "sdkdefault")
	assert.Equal(t, "sdkdefault", v)
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
			want: "Variations:[Default=false,False=false,True=true], Rules:[[query:[key eq \"toto\"], percentages:[False=90.00,True=10.00]]], DefaultRule:[variation:[Default]], TrackEvents:[true], Disable:[false], Version:[12.00]",
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
			want: "Variations:[Default=false,False=false,True=true], Rules:[[percentages:[False=90.00,True=10.00]]], DefaultRule:[variation:[Default]], TrackEvents:[true], Disable:[false]",
		},
		{
			name: "Default values",
			fields: fields{
				True:    true,
				False:   false,
				Default: false,
			},
			want: "Variations:[Default=false,False=false,True=true], Rules:[[percentages:[False=100.00,True=0.00]]], DefaultRule:[variation:[Default]], TrackEvents:[true], Disable:[false]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version := ""
			if tt.fields.Version != nil {
				version = fmt.Sprintf("%.2f", *tt.fields.Version)
			}
			dto := &flag.DtoFlag{
				Disable:     testconvert.Bool(tt.fields.Disable),
				Rule:        testconvert.String(tt.fields.Rule),
				Percentage:  testconvert.Float64(tt.fields.Percentage),
				True:        testconvert.Interface(tt.fields.True),
				False:       testconvert.Interface(tt.fields.False),
				Default:     testconvert.Interface(tt.fields.Default),
				TrackEvents: tt.fields.TrackEvents,
				Version:     testconvert.String(version),
			}

			f, err := dto.ConvertToFlagData(false)
			assert.NoError(t, err)

			got := f.String()
			assert.Equal(t, tt.want, got, "String() = %v, want %v", got, tt.want)
		})
	}
}

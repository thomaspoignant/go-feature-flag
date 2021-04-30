package model_test

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"github.com/thomaspoignant/go-feature-flag/ffuser"
	"github.com/thomaspoignant/go-feature-flag/internal/model"
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
		Rollout    model.Rollout
	}
	type args struct {
		flagName string
		user     ffuser.User
	}
	type want struct {
		value         interface{}
		variationType model.VariationType
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
				variationType: model.VariationDefault,
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
				variationType: model.VariationTrue,
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
				Rollout: model.Rollout{
					Experimentation: &model.Experimentation{
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
				variationType: model.VariationTrue,
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
				Rollout: model.Rollout{
					Experimentation: &model.Experimentation{
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
				variationType: model.VariationDefault,
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
				Rollout: model.Rollout{
					Experimentation: &model.Experimentation{
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
				variationType: model.VariationTrue,
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
				Rollout: model.Rollout{
					Experimentation: &model.Experimentation{
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
				variationType: model.VariationDefault,
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
				Rollout: model.Rollout{
					Experimentation: &model.Experimentation{
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
				variationType: model.VariationDefault,
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
				Rollout: model.Rollout{
					Experimentation: &model.Experimentation{
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
				variationType: model.VariationDefault,
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
				Rollout: model.Rollout{
					Experimentation: &model.Experimentation{
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
				variationType: model.VariationTrue,
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
				Rollout: model.Rollout{
					Experimentation: &model.Experimentation{
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
				variationType: model.VariationTrue,
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
				Rollout: model.Rollout{
					Experimentation: &model.Experimentation{
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
				variationType: model.VariationDefault,
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
				variationType: model.VariationDefault,
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
				variationType: model.VariationFalse,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &model.Flag{
				Disable:    tt.fields.Disable,
				Rule:       tt.fields.Rule,
				Percentage: tt.fields.Percentage,
				True:       tt.fields.True,
				False:      tt.fields.False,
				Default:    tt.fields.Default,
				Rollout:    &tt.fields.Rollout,
			}

			got, variationType := f.Value(tt.args.flagName, tt.args.user)
			assert.Equal(t, tt.want.value, got)
			assert.Equal(t, tt.want.variationType, variationType)
		})
	}
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
			},
			want: "percentage=10%, rule=\"key eq \"toto\"\", true=\"true\", false=\"false\", true=\"false\", disable=\"false\", trackEvents=\"true\"",
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
			want: "percentage=10%, true=\"true\", false=\"false\", true=\"false\", disable=\"false\"",
		},
		{
			name: "Default values",
			fields: fields{
				True:    true,
				False:   false,
				Default: false,
			},
			want: "percentage=0%, true=\"true\", false=\"false\", true=\"false\", disable=\"false\"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := model.Flag{
				Disable:     tt.fields.Disable,
				Rule:        tt.fields.Rule,
				Percentage:  tt.fields.Percentage,
				True:        tt.fields.True,
				False:       tt.fields.False,
				Default:     tt.fields.Default,
				TrackEvents: tt.fields.TrackEvents,
			}
			got := f.String()
			assert.Equal(t, tt.want, got, "String() = %v, want %v", got, tt.want)
		})
	}
}

func TestFlag_ProgressiveRollout(t *testing.T) {
	f := &model.Flag{
		Percentage: 0,
		True:       "True",
		False:      "False",
		Default:    "Default",
		Rollout: &model.Rollout{Progressive: &model.Progressive{
			ReleaseRamp: model.ProgressiveReleaseRamp{
				Start: testconvert.Time(time.Now().Add(1 * time.Second)),
				End:   testconvert.Time(time.Now().Add(2 * time.Second)),
			},
		}},
	}

	user := ffuser.NewAnonymousUser("test")
	flagName := "test-flag"

	// We evaluate the same flag multiple time overtime.
	v, _ := f.Value(flagName, user)
	assert.Equal(t, f.False, v)

	time.Sleep(1 * time.Second)
	v2, _ := f.Value(flagName, user)
	assert.Equal(t, f.False, v2)

	time.Sleep(1 * time.Second)
	v3, _ := f.Value(flagName, user)
	assert.Equal(t, f.True, v3)
}

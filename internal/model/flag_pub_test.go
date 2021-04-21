package model_test

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"github.com/thomaspoignant/go-feature-flag/ffuser"
	"github.com/thomaspoignant/go-feature-flag/internal/model"
	"github.com/thomaspoignant/go-feature-flag/testutils"
)

func TestFlag_value(t *testing.T) {
	type fields struct {
		Disable         bool
		Rule            string
		Percentage      float64
		True            interface{}
		False           interface{}
		Default         interface{}
		Experimentation model.Experimentation
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
			name: "Experimentation only start date in the past",
			fields: fields{
				True:       "true",
				False:      "false",
				Default:    "default",
				Rule:       "key == \"7e50ee61-06ad-4bb0-9034-38ad7cdea9f5\"",
				Percentage: 10,
				Experimentation: model.Experimentation{
					StartDate: testutils.Time(time.Now().Add(-1 * time.Minute)),
					EndDate:   nil,
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
			name: "Experimentation only start date in the future",
			fields: fields{
				True:       "true",
				False:      "false",
				Default:    "default",
				Rule:       "key == \"user66\"",
				Percentage: 10,
				Experimentation: model.Experimentation{
					StartDate: testutils.Time(time.Now().Add(1 * time.Minute)),
					EndDate:   nil,
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
			name: "Experimentation between start and end date",
			fields: fields{
				True:       "true",
				False:      "false",
				Default:    "default",
				Rule:       "key == \"7e50ee61-06ad-4bb0-9034-38ad7cdea9f5\"",
				Percentage: 10,
				Experimentation: model.Experimentation{
					StartDate: testutils.Time(time.Now().Add(-1 * time.Minute)),
					EndDate:   testutils.Time(time.Now().Add(1 * time.Minute)),
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
			name: "Experimentation not started yet",
			fields: fields{
				True:       "true",
				False:      "false",
				Default:    "default",
				Rule:       "key == \"user66\"",
				Percentage: 10,
				Experimentation: model.Experimentation{
					StartDate: testutils.Time(time.Now().Add(1 * time.Minute)),
					EndDate:   testutils.Time(time.Now().Add(2 * time.Minute)),
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
			name: "Experimentation finished",
			fields: fields{
				True:       "true",
				False:      "false",
				Default:    "default",
				Rule:       "key == \"user66\"",
				Percentage: 10,
				Experimentation: model.Experimentation{
					StartDate: testutils.Time(time.Now().Add(-2 * time.Minute)),
					EndDate:   testutils.Time(time.Now().Add(-1 * time.Minute)),
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
			name: "Experimentation only end date finished",
			fields: fields{
				True:       "true",
				False:      "false",
				Default:    "default",
				Rule:       "key == \"user66\"",
				Percentage: 10,
				Experimentation: model.Experimentation{
					StartDate: nil,
					EndDate:   testutils.Time(time.Now().Add(-1 * time.Minute)),
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
			name: "Experimentation only end date not finished",
			fields: fields{
				True:       "true",
				False:      "false",
				Default:    "default",
				Rule:       "key == \"7e50ee61-06ad-4bb0-9034-38ad7cdea9f5\"",
				Percentage: 10,
				Experimentation: model.Experimentation{
					StartDate: nil,
					EndDate:   testutils.Time(time.Now().Add(1 * time.Minute)),
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
			name: "Experimentation only end date not finished",
			fields: fields{
				True:       "true",
				False:      "false",
				Default:    "default",
				Rule:       "key == \"7e50ee61-06ad-4bb0-9034-38ad7cdea9f5\"",
				Percentage: 10,
				Experimentation: model.Experimentation{
					StartDate: nil,
					EndDate:   nil,
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
				Experimentation: model.Experimentation{
					StartDate: testutils.Time(time.Now().Add(1 * time.Minute)),
					EndDate:   testutils.Time(time.Now().Add(-1 * time.Minute)),
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
				Disable:         tt.fields.Disable,
				Rule:            tt.fields.Rule,
				Percentage:      tt.fields.Percentage,
				True:            tt.fields.True,
				False:           tt.fields.False,
				Default:         tt.fields.Default,
				Experimentation: &tt.fields.Experimentation,
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
				TrackEvents: testutils.Bool(true),
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

func TestExperimentation_String(t *testing.T) {
	type fields struct {
		StartDate *time.Time
		EndDate   *time.Time
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "both dates",
			fields: fields{
				StartDate: testutils.Time(time.Unix(1095379400, 0)),
				EndDate:   testutils.Time(time.Unix(1095379500, 0)),
			},
			want: "start:[2004-09-17T00:03:20Z] end:[2004-09-17T00:05:00Z]",
		},
		{
			name: "only start date",
			fields: fields{
				StartDate: testutils.Time(time.Unix(1095379400, 0)),
			},
			want: "start:[2004-09-17T00:03:20Z]",
		},
		{
			name: "only end date",
			fields: fields{
				EndDate: testutils.Time(time.Unix(1095379500, 0)),
			},
			want: "end:[2004-09-17T00:05:00Z]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := model.Experimentation{
				StartDate: tt.fields.StartDate,
				EndDate:   tt.fields.EndDate,
			}
			got := e.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

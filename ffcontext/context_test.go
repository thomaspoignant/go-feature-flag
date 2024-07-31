package ffcontext_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/testutils/testconvert"
	"testing"
	"time"
)

func TestUser_AddCustomAttribute(t *testing.T) {
	type args struct {
		name  string
		value interface{}
	}
	tests := []struct {
		name string
		user ffcontext.EvaluationContext
		args args
		want map[string]interface{}
	}{
		{
			name: "trying to add nil value",
			user: ffcontext.NewEvaluationContext("123"),
			args: args{},
			want: map[string]interface{}{},
		},
		{
			name: "add valid element",
			user: ffcontext.NewEvaluationContext("123"),
			args: args{
				name:  "test",
				value: "test",
			},
			want: map[string]interface{}{
				"test": "test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.user.AddCustomAttribute(tt.args.name, tt.args.value)
			assert.Equal(t, tt.want, tt.user.GetCustom())
		})
	}
}

func Test_ExtractGOFFProtectedFields(t *testing.T) {
	tests := []struct {
		name string
		ctx  ffcontext.EvaluationContext
		want ffcontext.GoffContextSpecifics
	}{
		{
			name: "context goff specifics as map[string]string",
			ctx: ffcontext.NewEvaluationContextBuilder("my-key").AddCustom("gofeatureflag", map[string]string{
				"currentDateTime": time.Date(2022, 8, 1, 0, 0, 10, 0, time.UTC).Format(time.RFC3339),
			}).Build(),
			want: ffcontext.GoffContextSpecifics{
				CurrentDateTime: testconvert.Time(time.Date(2022, 8, 1, 0, 0, 10, 0, time.UTC)),
			},
		},
		{
			name: "context goff specifics as map[string]interface and date as time.Time",
			ctx: ffcontext.NewEvaluationContextBuilder("my-key").AddCustom("gofeatureflag", map[string]interface{}{
				"currentDateTime": time.Date(2022, 8, 1, 0, 0, 10, 0, time.UTC),
			}).Build(),
			want: ffcontext.GoffContextSpecifics{
				CurrentDateTime: testconvert.Time(time.Date(2022, 8, 1, 0, 0, 10, 0, time.UTC)),
			},
		},
		{
			name: "context goff specifics as map[string]interface and date as *time.Time",
			ctx: ffcontext.NewEvaluationContextBuilder("my-key").AddCustom("gofeatureflag", map[string]interface{}{
				"currentDateTime": testconvert.Time(time.Date(2022, 8, 1, 0, 0, 10, 0, time.UTC)),
			}).Build(),
			want: ffcontext.GoffContextSpecifics{
				CurrentDateTime: testconvert.Time(time.Date(2022, 8, 1, 0, 0, 10, 0, time.UTC)),
			},
		},
		{
			name: "context goff specifics as map[string]interface",
			ctx: ffcontext.NewEvaluationContextBuilder("my-key").AddCustom("gofeatureflag", map[string]interface{}{
				"currentDateTime": time.Date(2022, 8, 1, 0, 0, 10, 0, time.UTC).Format(time.RFC3339),
			}).Build(),
			want: ffcontext.GoffContextSpecifics{
				CurrentDateTime: testconvert.Time(time.Date(2022, 8, 1, 0, 0, 10, 0, time.UTC)),
			},
		},
		{
			name: "context goff specifics nil",
			ctx:  ffcontext.NewEvaluationContextBuilder("my-key").AddCustom("gofeatureflag", nil).Build(),
			want: ffcontext.GoffContextSpecifics{
				CurrentDateTime: nil,
			},
		},
		{
			name: "no context goff specifics",
			ctx:  ffcontext.NewEvaluationContextBuilder("my-key").Build(),
			want: ffcontext.GoffContextSpecifics{
				CurrentDateTime: nil,
			},
		},
		{
			name: "context goff specifics as GoffContextSpecifics type",
			ctx: ffcontext.NewEvaluationContextBuilder("my-key").AddCustom("gofeatureflag", ffcontext.GoffContextSpecifics{
				CurrentDateTime: testconvert.Time(time.Date(2022, 8, 1, 0, 0, 10, 0, time.UTC)),
			}).Build(),
			want: ffcontext.GoffContextSpecifics{
				CurrentDateTime: testconvert.Time(time.Date(2022, 8, 1, 0, 0, 10, 0, time.UTC)),
			},
		},
		{
			name: "context goff specifics as GoffContextSpecifics type contains flagList",
			ctx: ffcontext.NewEvaluationContextBuilder("my-key").AddCustom("gofeatureflag", ffcontext.GoffContextSpecifics{
				CurrentDateTime: testconvert.Time(time.Date(2022, 8, 1, 0, 0, 10, 0, time.UTC)),
				FlagList:        []string{"flag1", "flag2"},
			}).Build(),
			want: ffcontext.GoffContextSpecifics{
				CurrentDateTime: testconvert.Time(time.Date(2022, 8, 1, 0, 0, 10, 0, time.UTC)),
				FlagList:        []string{"flag1", "flag2"},
			},
		},
		{
			name: "context goff specifics as map[string]interface type contains flagList",
			ctx: ffcontext.NewEvaluationContextBuilder("my-key").AddCustom("gofeatureflag", map[string]interface{}{
				"currentDateTime": testconvert.Time(time.Date(2022, 8, 1, 0, 0, 10, 0, time.UTC)).Format(time.RFC3339),
				"flagList":        []string{"flag1", "flag2"},
			}).Build(),
			want: ffcontext.GoffContextSpecifics{
				CurrentDateTime: testconvert.Time(time.Date(2022, 8, 1, 0, 0, 10, 0, time.UTC)),
				FlagList:        []string{"flag1", "flag2"},
			},
		},
		{
			name: "context goff specifics only flagList",
			ctx: ffcontext.NewEvaluationContextBuilder("my-key").AddCustom("gofeatureflag", map[string]interface{}{
				"flagList": []string{"flag1", "flag2"},
			}).Build(),
			want: ffcontext.GoffContextSpecifics{
				FlagList: []string{"flag1", "flag2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ctx.ExtractGOFFProtectedFields()
			assert.Equal(t, tt.want, got)
		})
	}
}

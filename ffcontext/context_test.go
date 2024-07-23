package ffcontext_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
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
	ctx := ffcontext.NewEvaluationContext("my-key")
	ctx.AddCustomAttribute("toto", "tata")
	ctx.AddCustomAttribute("gofeatureflag", map[string]interface{}{
		"currentDateTime": "2022-08-01T00:00:00.1+02:00",
	})
	p := ctx.ExtractGOFFProtectedFields()
	want := time.Date(2022, 8, 1, 0, 0, 0, 100000000, time.Local)
	assert.Equal(t, want, *p.CurrentDateTime)
}

func Test_ExtractGOFFProtectedFields_nil(t *testing.T) {
	ctx := ffcontext.NewEvaluationContext("my-key")
	ctx.AddCustomAttribute("toto", "tata")
	p := ctx.ExtractGOFFProtectedFields()
	assert.Nil(t, p.CurrentDateTime)
}

package flags

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/thomaspoignant/go-feature-flag/ffuser"
)

func TestFlag_evaluateRule(t *testing.T) {
	type fields struct {
		Disable    bool
		Rule       string
		Percentage int
		True       interface{}
		False      interface{}
	}
	type args struct {
		user ffuser.User
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "Disabled toggle",
			fields: fields{
				Disable: true,
			},
			args: args{
				user: ffuser.NewAnonymousUser("random-key"),
			},
			want: false,
		},
		{
			name: "Toggle enabled and no rule",
			fields: fields{
				Disable: false,
			},
			args: args{
				user: ffuser.NewAnonymousUser("random-key"),
			},
			want: true,
		},
		{
			name: "Toggle enabled with rule success",
			fields: fields{
				Rule: "key == \"random-key\"",
			},
			args: args{
				user: ffuser.NewAnonymousUser("random-key"),
			},
			want: true,
		},
		{
			name: "Toggle enabled with rule failure",
			fields: fields{
				Rule: "key == \"incorrect-key\"",
			},
			args: args{
				user: ffuser.NewAnonymousUser("random-key"),
			},
			want: false,
		},
		{
			name: "Toggle enabled with no key",
			fields: fields{
				Rule: "key == \"random-key\"",
			},
			args: args{
				user: ffuser.NewAnonymousUser(""),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Flag{
				Disable:    tt.fields.Disable,
				Rule:       tt.fields.Rule,
				Percentage: tt.fields.Percentage,
				True:       tt.fields.True,
				False:      tt.fields.False,
			}

			got := f.evaluateRule(tt.args.user)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFlag_isInPercentage(t *testing.T) {
	type fields struct {
		Disable    bool
		Rule       string
		Percentage int
		True       interface{}
		False      interface{}
	}
	type args struct {
		flagName string
		user     ffuser.User
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "Anything should work at 100%",
			fields: fields{
				Percentage: 100,
			},
			args: args{
				flagName: "test_689025",
				user:     ffuser.NewUser("test_689053"),
			},
			want: true,
		},
		{
			name: "Anything should work at 0%",
			fields: fields{
				Percentage: 0,
			},
			args: args{
				flagName: "test_689025",
				user:     ffuser.NewUser("test_689053"),
			},
			want: false,
		},
		{
			name: "User flag in the range",
			fields: fields{
				Percentage: 10,
			},
			args: args{
				flagName: "test_689025",                 // hash is 20
				user:     ffuser.NewUser("test_689053"), // hash of the key is 29
			},
			want: true,
		},
		{
			name: "User toggle in the range, after the modulo",
			fields: fields{
				Percentage: 10,
			},
			args: args{
				flagName: "test_689054",                 // hash is 96
				user:     ffuser.NewUser("test_689061"), // hash of the key is 4
			},
			want: true,
		},
		{
			name: "User toggle same hash as the toggle",
			fields: fields{
				Percentage: 10,
			},
			args: args{
				flagName: "test_689054",                 // hash is 96
				user:     ffuser.NewUser("test_689371"), // hash of the key is 96
			},
			want: true,
		},
		{
			name: "User toggle not in the range",
			fields: fields{
				Percentage: 10,
			},
			args: args{
				flagName: "test_689372",                 // hash is 53
				user:     ffuser.NewUser("test_689373"), // hash of the key is 54
			},
			want: false,
		},
		{
			name: "User toggle equals higher range",
			fields: fields{
				Percentage: 10,
			},
			args: args{
				flagName: "test_689372",                 // hash is 53
				user:     ffuser.NewUser("test_689470"), // hash of the key is 62
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Flag{
				Disable:    tt.fields.Disable,
				Rule:       tt.fields.Rule,
				Percentage: tt.fields.Percentage,
				True:       tt.fields.True,
				False:      tt.fields.False,
			}

			got := f.isInPercentage(tt.args.flagName, tt.args.user)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFlag_value(t *testing.T) {
	type fields struct {
		Disable    bool
		Rule       string
		Percentage int
		True       interface{}
		False      interface{}
		Default    interface{}
	}
	type args struct {
		flagName string
		user     ffuser.User
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   interface{}
	}{
		{
			name: "Rule disable get default Value",
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
			want: "default",
		},
		{
			name: "Get true Value if rule pass",
			fields: fields{
				True:       "true",
				False:      "false",
				Default:    "default",
				Rule:       "key == \"test_689483\"",
				Percentage: 10,
			},
			args: args{
				flagName: "test_689483",
				user:     ffuser.NewUser("test_689483"),
			},
			want: "true",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Flag{
				Disable:    tt.fields.Disable,
				Rule:       tt.fields.Rule,
				Percentage: tt.fields.Percentage,
				True:       tt.fields.True,
				False:      tt.fields.False,
				Default:    tt.fields.Default,
			}

			got := f.Value(tt.args.flagName, tt.args.user)
			assert.Equal(t, tt.want, got)
		})
	}
}

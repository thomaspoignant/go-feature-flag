package model

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/thomaspoignant/go-feature-flag/ffuser"
)

func TestFlag_evaluateRule(t *testing.T) {
	type fields struct {
		Disable    bool
		Rule       string
		Percentage float64
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
		Percentage float64
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
				flagName: "test-flag",
				user:     ffuser.NewUser("user2"), // combined hash is 1
			},
			want: true,
		},
		{
			name: "High limit of the percentage",
			fields: fields{
				Percentage: 10,
			},
			args: args{
				flagName: "test-flag",
				user:     ffuser.NewUser("user66"), // combined hash is 9
			},
			want: true,
		},
		{
			name: "Limit +1",
			fields: fields{
				Percentage: 10,
			},
			args: args{
				flagName: "test-flag",
				user:     ffuser.NewUser("user40"), // combined hash is 10
			},
			want: false,
		},
		{
			name: "Low limit of the percentage",
			fields: fields{
				Percentage: 10,
			},
			args: args{
				flagName: "test-flag",
				user:     ffuser.NewUser("user135"), // hash of the key is 0
			},
			want: true,
		},
		{
			name: "Flag not in the percentage",
			fields: fields{
				Percentage: 10,
			},
			args: args{
				flagName: "test-flag",
				user:     ffuser.NewUser("user134"), // hash of the key is 19
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

			got := f.isInPercentage(tt.args.flagName, tt.args.user)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFlag_value(t *testing.T) {
	type fields struct {
		Disable    bool
		Rule       string
		Percentage float64
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
			want: "default",
		},
		{
			name: "Get true value if rule pass",
			fields: fields{
				True:       "true",
				False:      "false",
				Default:    "default",
				Rule:       "key == \"user66\"",
				Percentage: 10,
			},
			args: args{
				flagName: "test-flag",
				user:     ffuser.NewUserBuilder("user66").AddCustom("name", "john").Build(), // combined hash is 9
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

func TestFlag_String(t *testing.T) {
	type fields struct {
		Disable    bool
		Rule       string
		Percentage float64
		True       interface{}
		False      interface{}
		Default    interface{}
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "All fields",
			fields: fields{
				Disable:    false,
				Rule:       "key eq \"toto\"",
				Percentage: 10,
				True:       true,
				False:      false,
				Default:    false,
			},
			want: "percentage=10%, rule=\"key eq \"toto\"\", true=\"true\", false=\"false\", true=\"false\", disable=\"false\"",
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
			f := Flag{
				Disable:    tt.fields.Disable,
				Rule:       tt.fields.Rule,
				Percentage: tt.fields.Percentage,
				True:       tt.fields.True,
				False:      tt.fields.False,
				Default:    tt.fields.Default,
			}
			if got := f.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

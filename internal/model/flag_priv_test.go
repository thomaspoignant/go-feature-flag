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
			name: "105% should work as 100%",
			fields: fields{
				Percentage: 105,
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
			name: "-1% should work like 0%",
			fields: fields{
				Percentage: -1,
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
				user:     ffuser.NewUser("86fe0fd9-d19c-4c35-bd05-07b434a21c04"),
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
				user:     ffuser.NewUser("7e50ee61-06ad-4bb0-9034-38ad7cdea9f5"),
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
				user:     ffuser.NewUser("a287f16a-b50b-4151-a50f-a97fe334a4bf"),
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
				user:     ffuser.NewUser("a4599f14-f7a3-4c14-b3b9-0c0d728224ff"),
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
				user:     ffuser.NewUser("ffc35559-bc1d-4cf3-8e21-7f95c432d1c2"),
			},
			want: false,
		},
		{
			name: "float percentage",
			fields: fields{
				Percentage: 10.123,
			},
			args: args{
				flagName: "test-flag",
				user:     ffuser.NewUser("ffc35559-bc1d-4cf3-8e21-7f95c432d1c2"),
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

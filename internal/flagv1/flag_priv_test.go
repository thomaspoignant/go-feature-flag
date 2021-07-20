package flagv1

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"github.com/thomaspoignant/go-feature-flag/ffuser"
	"github.com/thomaspoignant/go-feature-flag/testutils/testconvert"
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
			f := &FlagData{
				Disable:    testconvert.Bool(tt.fields.Disable),
				Rule:       testconvert.String(tt.fields.Rule),
				Percentage: testconvert.Float64(tt.fields.Percentage),
				True:       testconvert.Interface(tt.fields.True),
				False:      testconvert.Interface(tt.fields.False),
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
			f := &FlagData{
				Disable:    testconvert.Bool(tt.fields.Disable),
				Rule:       testconvert.String(tt.fields.Rule),
				Percentage: testconvert.Float64(tt.fields.Percentage),
				True:       testconvert.Interface(tt.fields.True),
				False:      testconvert.Interface(tt.fields.False),
			}

			got := f.isInPercentage(tt.args.flagName, tt.args.user)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFlag_getPercentage(t *testing.T) {
	tests := []struct {
		name string
		flag FlagData
		want float64
	}{
		{
			name: "No rollout strategy 100",
			flag: FlagData{
				Percentage: testconvert.Float64(100),
			},
			want: 100 * percentageMultiplier,
		},
		{
			name: "No rollout strategy 0",
			flag: FlagData{
				Percentage: testconvert.Float64(0),
			},
			want: 0 * percentageMultiplier,
		},
		{
			name: "No rollout strategy 50",
			flag: FlagData{
				Percentage: testconvert.Float64(50),
			},
			want: 50 * percentageMultiplier,
		},
		{
			name: "Progressive rollout no explicit percentage",
			flag: FlagData{
				Rollout: &Rollout{
					Progressive: &Progressive{
						ReleaseRamp: ProgressiveReleaseRamp{
							Start: testconvert.Time(time.Now().Add(-1 * time.Minute)),
							End:   testconvert.Time(time.Now().Add(1 * time.Minute)),
						},
					},
				},
			},
			want: 50 * percentageMultiplier,
		},
		{
			name: "Progressive rollout explicit initial percentage",
			flag: FlagData{
				Rollout: &Rollout{
					Progressive: &Progressive{
						Percentage: ProgressivePercentage{
							Initial: 20,
						},
						ReleaseRamp: ProgressiveReleaseRamp{
							Start: testconvert.Time(time.Now().Add(-1 * time.Minute)),
							End:   testconvert.Time(time.Now().Add(1 * time.Minute)),
						},
					},
				},
			},
			want: float64(60000),
		},
		{
			name: "Progressive rollout explicit end percentage",
			flag: FlagData{
				Rollout: &Rollout{
					Progressive: &Progressive{
						Percentage: ProgressivePercentage{
							End: 20,
						},
						ReleaseRamp: ProgressiveReleaseRamp{
							Start: testconvert.Time(time.Now().Add(-1 * time.Minute)),
							End:   testconvert.Time(time.Now().Add(1 * time.Minute)),
						},
					},
				},
			},
			want: float64(10000),
		},
		{
			name: "Progressive rollout explicit initial and end percentage",
			flag: FlagData{
				Rollout: &Rollout{
					Progressive: &Progressive{
						Percentage: ProgressivePercentage{
							Initial: 10,
							End:     20,
						},
						ReleaseRamp: ProgressiveReleaseRamp{
							Start: testconvert.Time(time.Now().Add(-1 * time.Minute)),
							End:   testconvert.Time(time.Now().Add(1 * time.Minute)),
						},
					},
				},
			},
			want: float64(15000),
		},
		{
			name: "Progressive rollout before date",
			flag: FlagData{
				Rollout: &Rollout{
					Progressive: &Progressive{
						Percentage: ProgressivePercentage{
							Initial: 10,
							End:     20,
						},
						ReleaseRamp: ProgressiveReleaseRamp{
							Start: testconvert.Time(time.Now().Add(1 * time.Minute)),
							End:   testconvert.Time(time.Now().Add(2 * time.Minute)),
						},
					},
				},
			},
			want: float64(10000),
		},
		{
			name: "Progressive rollout after date",
			flag: FlagData{
				Rollout: &Rollout{
					Progressive: &Progressive{
						Percentage: ProgressivePercentage{
							Initial: 10,
							End:     80,
						},
						ReleaseRamp: ProgressiveReleaseRamp{
							Start: testconvert.Time(time.Now().Add(-1 * time.Minute)),
							End:   testconvert.Time(time.Now().Add(-2 * time.Minute)),
						},
					},
				},
			},
			want: float64(80000),
		},
		{
			name: "End percentage lower than start use top level percentage",
			flag: FlagData{
				Rollout: &Rollout{
					Progressive: &Progressive{
						Percentage: ProgressivePercentage{
							Initial: 80,
							End:     10,
						},
						ReleaseRamp: ProgressiveReleaseRamp{
							Start: testconvert.Time(time.Now().Add(-1 * time.Minute)),
							End:   testconvert.Time(time.Now().Add(1 * time.Minute)),
						},
					},
				},
			},
			want: float64(0),
		},
		{
			name: "Missing end date",
			flag: FlagData{
				Rollout: &Rollout{
					Progressive: &Progressive{
						ReleaseRamp: ProgressiveReleaseRamp{
							Start: testconvert.Time(time.Now().Add(-1 * time.Minute)),
						},
					},
				},
			},
			want: float64(0),
		},
		{
			name: "Missing start date",
			flag: FlagData{
				Rollout: &Rollout{
					Progressive: &Progressive{
						ReleaseRamp: ProgressiveReleaseRamp{
							End: testconvert.Time(time.Now().Add(-1 * time.Minute)),
						},
					},
				},
			},
			want: float64(0),
		},
		{
			name: "Missing date use default percentage",
			flag: FlagData{
				Percentage: testconvert.Float64(46),
				Rollout: &Rollout{
					Progressive: &Progressive{
						ReleaseRamp: ProgressiveReleaseRamp{
							End: testconvert.Time(time.Now().Add(-1 * time.Minute)),
						},
					},
				},
			},
			want: float64(46000),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.flag.getActualPercentage()
			assert.Equal(t, tt.want, got)
		})
	}
}

package flag_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/ffuser"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	"github.com/thomaspoignant/go-feature-flag/internal/rollout"
	"github.com/thomaspoignant/go-feature-flag/testutils/testconvert"
	"testing"
	"time"
)

func TestFlag_value(t *testing.T) {

	type want struct {
		value         interface{}
		variationType string
	}
	type args struct {
		flagName   string
		user       ffuser.User
		sdkDefault interface{}
	}

	tests := []struct {
		name      string
		inputFlag flag.Flag
		args      args
		want      want
	}{
		{
			name: "Rule disable get sdk default value",
			inputFlag: &flag.FlagData{
				Variations: &map[string]*interface{}{
					"True":    testconvert.Interface("true"),
					"False":   testconvert.Interface("false"),
					"Default": testconvert.Interface("default"),
				},
				DefaultRule: nil,
				Disable:     testconvert.Bool(true),
			},
			args: args{
				flagName:   "test_689483",
				user:       ffuser.NewUser("test_689483"),
				sdkDefault: "sdk-Default",
			},
			want: want{
				value:         "sdk-Default",
				variationType: "SdkDefault",
			},
		},
		{
			name: "Get true value if rule pass",
			inputFlag: &flag.FlagData{
				Variations: &map[string]*interface{}{
					"True":    testconvert.Interface("true"),
					"False":   testconvert.Interface("false"),
					"Default": testconvert.Interface("default"),
				},
				Rules: &[]flag.Rule{{
					Query:       testconvert.String("key == \"7e50ee61-06ad-4bb0-9034-38ad7cdea9f5\""),
					Percentages: &map[string]float64{"True": 10, "False": 90},
				}},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("Default"),
				},
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
			inputFlag: &flag.FlagData{
				Variations: &map[string]*interface{}{
					"True":    testconvert.Interface("true"),
					"False":   testconvert.Interface("false"),
					"Default": testconvert.Interface("default"),
				},
				Rules: &[]flag.Rule{{
					Query:       testconvert.String("key == \"7e50ee61-06ad-4bb0-9034-38ad7cdea9f5\""),
					Percentages: &map[string]float64{"True": 10, "False": 90},
				}},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("Default"),
				},
				Rollout: &flag.Rollout{
					Experimentation: &rollout.Experimentation{
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
			inputFlag: &flag.FlagData{
				Variations: &map[string]*interface{}{
					"True":    testconvert.Interface("true"),
					"False":   testconvert.Interface("false"),
					"Default": testconvert.Interface("default"),
				},
				Rules: &[]flag.Rule{{
					Query:       testconvert.String("key == \"7e50ee61-06ad-4bb0-9034-38ad7cdea9f5\""),
					Percentages: &map[string]float64{"True": 10, "False": 90},
				}},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("Default"),
				},
				Rollout: &flag.Rollout{
					Experimentation: &rollout.Experimentation{
						Start: testconvert.Time(time.Now().Add(1 * time.Minute)),
						End:   nil,
					},
				},
			},
			args: args{
				flagName:   "test-flag",
				user:       ffuser.NewUserBuilder("user66").AddCustom("name", "john").Build(), // combined hash is 9
				sdkDefault: "default",
			},
			want: want{
				value:         "default",
				variationType: "SdkDefault",
			},
		},
		{
			name: "Rollout Experimentation between start and end date",
			inputFlag: &flag.FlagData{
				Variations: &map[string]*interface{}{
					"True":    testconvert.Interface("true"),
					"False":   testconvert.Interface("false"),
					"Default": testconvert.Interface("default"),
				},
				Rules: &[]flag.Rule{{
					Query:       testconvert.String("key == \"7e50ee61-06ad-4bb0-9034-38ad7cdea9f5\""),
					Percentages: &map[string]float64{"True": 10, "False": 90},
				}},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("Default"),
				},
				Rollout: &flag.Rollout{
					Experimentation: &rollout.Experimentation{
						Start: testconvert.Time(time.Now().Add(-1 * time.Minute)),
						End:   testconvert.Time(time.Now().Add(1 * time.Minute)),
					},
				},
			},
			args: args{
				flagName:   "test-flag",
				user:       ffuser.NewUserBuilder("7e50ee61-06ad-4bb0-9034-38ad7cdea9f5").AddCustom("name", "john").Build(),
				sdkDefault: "default",
			},
			want: want{
				value:         "true",
				variationType: "True",
			},
		},
		{
			name: "Rollout Experimentation not started yet",
			inputFlag: &flag.FlagData{
				Variations: &map[string]*interface{}{
					"True":    testconvert.Interface("true"),
					"False":   testconvert.Interface("false"),
					"Default": testconvert.Interface("default"),
				},
				Rules: &[]flag.Rule{{
					Query:       testconvert.String("key == \"user66\""),
					Percentages: &map[string]float64{"True": 10, "False": 90},
				}},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("Default"),
				},
				Rollout: &flag.Rollout{
					Experimentation: &rollout.Experimentation{
						Start: testconvert.Time(time.Now().Add(1 * time.Minute)),
						End:   testconvert.Time(time.Now().Add(2 * time.Minute)),
					},
				},
			},
			args: args{
				flagName:   "test-flag",
				user:       ffuser.NewUserBuilder("user66").AddCustom("name", "john").Build(), // combined hash is 9
				sdkDefault: "default",
			},
			want: want{
				value:         "default",
				variationType: "SdkDefault",
			},
		},
		{
			name: "Rollout Experimentation finished",
			inputFlag: &flag.FlagData{
				Variations: &map[string]*interface{}{
					"True":    testconvert.Interface("true"),
					"False":   testconvert.Interface("false"),
					"Default": testconvert.Interface("default"),
				},
				Rules: &[]flag.Rule{{
					Query:       testconvert.String("key == \"user66\""),
					Percentages: &map[string]float64{"True": 10, "False": 90},
				}},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("Default"),
				},
				Rollout: &flag.Rollout{
					Experimentation: &rollout.Experimentation{
						Start: testconvert.Time(time.Now().Add(-2 * time.Minute)),
						End:   testconvert.Time(time.Now().Add(-1 * time.Minute)),
					},
				},
			},
			args: args{
				flagName:   "test-flag",
				user:       ffuser.NewUserBuilder("user66").AddCustom("name", "john").Build(), // combined hash is 9
				sdkDefault: "default",
			},
			want: want{
				value:         "default",
				variationType: "SdkDefault",
			},
		},
		{
			name: "Rollout Experimentation only end date finished",
			inputFlag: &flag.FlagData{
				Variations: &map[string]*interface{}{
					"True":    testconvert.Interface("true"),
					"False":   testconvert.Interface("false"),
					"Default": testconvert.Interface("default"),
				},
				Rules: &[]flag.Rule{{
					Query:       testconvert.String("key == \"user66\""),
					Percentages: &map[string]float64{"True": 10, "False": 90},
				}},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("Default"),
				},
				Rollout: &flag.Rollout{
					Experimentation: &rollout.Experimentation{
						Start: nil,
						End:   testconvert.Time(time.Now().Add(-1 * time.Minute)),
					},
				},
			},
			args: args{
				flagName:   "test-flag",
				user:       ffuser.NewUserBuilder("user66").AddCustom("name", "john").Build(), // combined hash is 9
				sdkDefault: "default",
			},
			want: want{
				value:         "default",
				variationType: "SdkDefault",
			},
		},
		{
			name: "Rollout Experimentation only end date not finished",
			inputFlag: &flag.FlagData{
				Variations: &map[string]*interface{}{
					"True":    testconvert.Interface("true"),
					"False":   testconvert.Interface("false"),
					"Default": testconvert.Interface("default"),
				},
				Rules: &[]flag.Rule{{
					Query:       testconvert.String("key == \"7e50ee61-06ad-4bb0-9034-38ad7cdea9f5\""),
					Percentages: &map[string]float64{"True": 10, "False": 90},
				}},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("Default"),
				},
				Rollout: &flag.Rollout{
					Experimentation: &rollout.Experimentation{
						Start: nil,
						End:   testconvert.Time(time.Now().Add(1 * time.Minute)),
					},
				},
			},
			args: args{
				flagName:   "test-flag",
				user:       ffuser.NewUserBuilder("7e50ee61-06ad-4bb0-9034-38ad7cdea9f5").AddCustom("name", "john").Build(),
				sdkDefault: "default",
			},
			want: want{
				value:         "true",
				variationType: "True",
			},
		},
		{
			name: "Rollout Experimentation both date nil",
			inputFlag: &flag.FlagData{
				Variations: &map[string]*interface{}{
					"True":    testconvert.Interface("true"),
					"False":   testconvert.Interface("false"),
					"Default": testconvert.Interface("default"),
				},
				Rules: &[]flag.Rule{{
					Query:       testconvert.String("key == \"7e50ee61-06ad-4bb0-9034-38ad7cdea9f5\""),
					Percentages: &map[string]float64{"True": 10, "False": 90},
				}},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("Default"),
				},
				Rollout: &flag.Rollout{
					Experimentation: &rollout.Experimentation{
						Start: nil,
						End:   nil,
					},
				},
			},
			args: args{
				flagName:   "test-flag",
				user:       ffuser.NewUserBuilder("7e50ee61-06ad-4bb0-9034-38ad7cdea9f5").AddCustom("name", "john").Build(),
				sdkDefault: "default",
			},
			want: want{
				value:         "true",
				variationType: "True",
			},
		},
		{
			name: "Invert start date and end date",
			inputFlag: &flag.FlagData{
				Variations: &map[string]*interface{}{
					"True":    testconvert.Interface("true"),
					"False":   testconvert.Interface("false"),
					"Default": testconvert.Interface("default"),
				},
				Rules: &[]flag.Rule{{
					Query:       testconvert.String("key == \"user66\""),
					Percentages: &map[string]float64{"True": 10, "False": 90},
				}},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("Default"),
				},
				Rollout: &flag.Rollout{
					Experimentation: &rollout.Experimentation{
						Start: testconvert.Time(time.Now().Add(1 * time.Minute)),
						End:   testconvert.Time(time.Now().Add(-1 * time.Minute)),
					},
				},
			},
			args: args{
				flagName:   "test-flag",
				user:       ffuser.NewUserBuilder("user66").AddCustom("name", "john").Build(),
				sdkDefault: "default",
			},
			want: want{
				value:         "default",
				variationType: "SdkDefault",
			},
		},
		{
			name: "Get default value if does not pass",
			inputFlag: &flag.FlagData{
				Variations: &map[string]*interface{}{
					"True":    testconvert.Interface("true"),
					"False":   testconvert.Interface("false"),
					"Default": testconvert.Interface("default"),
				},
				Rules: &[]flag.Rule{{
					Query:       testconvert.String("key == \"7e50ee61-06ad-4bb0-9034-38ad7\""),
					Percentages: &map[string]float64{"True": 10, "False": 90},
				}},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("Default"),
				},
			},
			args: args{
				flagName:   "test-flag",
				user:       ffuser.NewUserBuilder("7e50ee61-06ad-4bb0-9034-38ad7cdea9f5").AddCustom("name", "john").Build(),
				sdkDefault: "default",
			},
			want: want{
				value:         "default",
				variationType: "Default",
			},
		},
		{
			name: "Get false value if rule pass and not in the cohort",
			inputFlag: &flag.FlagData{
				Variations: &map[string]*interface{}{
					"True":    testconvert.Interface("true"),
					"False":   testconvert.Interface("false"),
					"Default": testconvert.Interface("default"),
				},
				Rules: &[]flag.Rule{{
					Query:       testconvert.String("key == \"7e50ee61-06ad-4bb0-9034-38ad7cdea9f5\""),
					Percentages: &map[string]float64{"True": 10, "False": 90},
				}},
				DefaultRule: &flag.Rule{
					VariationResult: testconvert.String("Default"),
				},
			},
			args: args{
				flagName:   "test-flag2",
				user:       ffuser.NewUserBuilder("7e50ee61-06ad-4bb0-9034-38ad7cdea9f5").AddCustom("name", "john").Build(),
				sdkDefault: "default",
			},
			want: want{
				value:         "false",
				variationType: "False",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, variationType := tt.inputFlag.Value(tt.args.flagName, tt.args.user, tt.args.sdkDefault)
			assert.Equal(t, tt.want.value, got)
			assert.Equal(t, tt.want.variationType, variationType)
		})
	}
}

func TestFlag_ProgressiveRollout(t *testing.T) {
	sdkDefaultValue := "toto"
	f := flag.FlagData{
		Variations: &map[string]*interface{}{
			"True":    testconvert.Interface("True"),
			"False":   testconvert.Interface("False"),
			"Default": testconvert.Interface("Default"),
		},
		DefaultRule: &flag.Rule{
			ProgressiveRollout: &flag.ProgressiveRollout{
				Initial: &flag.ProgressiveRolloutStep{
					Date:      testconvert.Time(time.Now().Add(1 * time.Second)),
					Variation: testconvert.String("False"),
				},
				End: &flag.ProgressiveRolloutStep{
					Date:      testconvert.Time(time.Now().Add(2 * time.Second)),
					Variation: testconvert.String("True"),
				},
			},
		},
	}

	user := ffuser.NewAnonymousUser("test")
	flagName := "test-flag"

	// We Evaluate the same flag multiple time overtime.
	v, _ := f.Value(flagName, user, sdkDefaultValue)
	assert.Equal(t, f.GetVariationValue("False"), v)

	time.Sleep(1 * time.Second)
	v2, _ := f.Value(flagName, user, sdkDefaultValue)
	assert.Equal(t, f.GetVariationValue("False"), v2)

	time.Sleep(1 * time.Second)
	v3, _ := f.Value(flagName, user, sdkDefaultValue)
	assert.Equal(t, f.GetVariationValue("True"), v3)
}

func TestFlag_ScheduledRollout(t *testing.T) {
	f := &flag.FlagData{
		Variations: &map[string]*interface{}{
			"True":    testconvert.Interface("True"),
			"False":   testconvert.Interface("False"),
			"Default": testconvert.Interface("Default"),
		},
		Rules: &[]flag.Rule{{
			Query: testconvert.String("key eq \"test\""),
			Percentages: &map[string]float64{
				"True": 0, "False": 100,
			},
		}},
		DefaultRule: &flag.Rule{
			VariationResult: testconvert.String("Default"),
		},
		Rollout: &flag.Rollout{
			Scheduled: &flag.ScheduledRollout{
				Steps: []flag.ScheduledStep{
					{
						FlagData: flag.FlagData{
							Version: testconvert.String("1.1"),
						},
						Date: testconvert.Time(time.Now().Add(1 * time.Second)),
					},
					{
						FlagData: flag.FlagData{
							Rules: &[]flag.Rule{{
								Query: testconvert.String("key eq \"test\""),
								Percentages: &map[string]float64{
									"True": 100, "False": 0,
								},
							}},
						},
						Date: testconvert.Time(time.Now().Add(1 * time.Second)),
					},
					{
						Date: testconvert.Time(time.Now().Add(2 * time.Second)),
						FlagData: flag.FlagData{
							Variations: &map[string]*interface{}{
								"True":    testconvert.Interface("True2"),
								"False":   testconvert.Interface("False2"),
								"Default": testconvert.Interface("Default2"),
							},
							Rules: &[]flag.Rule{{
								Query: testconvert.String("key eq \"test2\""),
								Percentages: &map[string]float64{
									"True": 100, "False": 0,
								},
							}},
						},
					},
					{
						Date: testconvert.Time(time.Now().Add(3 * time.Second)),
						FlagData: flag.FlagData{
							Rules: &[]flag.Rule{{
								Query: testconvert.String("key eq \"test\""),
								Percentages: &map[string]float64{
									"True": 100, "False": 0,
								},
							}},
						},
					},
					{
						FlagData: flag.FlagData{
							Disable: testconvert.Bool(true),
						},
						Date: testconvert.Time(time.Now().Add(4 * time.Second)),
					},
					{
						FlagData: flag.FlagData{
							Rules: &[]flag.Rule{{
								Query: testconvert.String("key eq \"test\""),
								Percentages: &map[string]float64{
									"True": 0, "False": 100,
								},
							}},
						},
					},
					{
						FlagData: flag.FlagData{
							Disable:     testconvert.Bool(false),
							TrackEvents: testconvert.Bool(true),
							Rollout: &flag.Rollout{
								Experimentation: &rollout.Experimentation{
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
	sdkDefault := "sdkDefault"
	// We Evaluate the same flag multiple time overtime.
	v, _ := f.Value(flagName, user, sdkDefault)
	assert.Equal(t, f.GetVariationValue("False"), v)

	time.Sleep(1 * time.Second)

	v, _ = f.Value(flagName, user, sdkDefault)
	assert.Equal(t, "True", v)
	assert.Equal(t, "1.1", f.GetVersion())

	time.Sleep(1 * time.Second)

	v, _ = f.Value(flagName, user, sdkDefault)
	assert.Equal(t, "Default2", v)

	time.Sleep(1 * time.Second)

	v, _ = f.Value(flagName, user, sdkDefault)
	assert.Equal(t, "True2", v)

	time.Sleep(1 * time.Second)

	v, _ = f.Value(flagName, user, sdkDefault)
	assert.Equal(t, "sdkDefault", v)

	time.Sleep(1 * time.Second)

	v, _ = f.Value(flagName, user, sdkDefault)
	assert.Equal(t, "sdkDefault", v)

	time.Sleep(1 * time.Second)

	v, _ = f.Value(flagName, user, sdkDefault)
	assert.Equal(t, "True2", v)

	time.Sleep(1 * time.Second)

	v, _ = f.Value(flagName, user, sdkDefault)
	assert.Equal(t, "sdkDefault", v)
}

//func TestFlag_Getter(t *testing.T) {
//	type expected struct {
//		True        interface{}
//		False       interface{}
//		Default     interface{}
//		Rollout     *flagv1.Rollout
//		Disable     bool
//		TrackEvents bool
//		Percentage  float64
//		Rule        string
//		Version     string
//		RawValues   map[string]string
//	}
//	tests := []struct {
//		name string
//		flag flag.Flag
//		want expected
//	}{
//		{
//			name: "all default",
//			flag: &flag.FlagData{},
//			want: expected{
//				True:        nil,
//				False:       nil,
//				Default:     nil,
//				Rollout:     nil,
//				Disable:     false,
//				TrackEvents: true,
//				Percentage:  0,
//				Rule:        "",
//				RawValues: map[string]string{
//					"Default":     "",
//					"Disable":     "false",
//					"False":       "",
//					"Percentage":  "0.00",
//					"Rollout":     "",
//					"Rule":        "",
//					"TrackEvents": "true",
//					"True":        "",
//					"Version":     "",
//				},
//			},
//		},
//		{
//			name: "custom flag",
//			flag: &flagv1.FlagData{
//				Rule:        testconvert.String("test"),
//				Percentage:  testconvert.Float64(90),
//				True:        testconvert.Interface(12.2),
//				False:       testconvert.Interface(13.2),
//				Default:     testconvert.Interface(14.2),
//				TrackEvents: testconvert.Bool(false),
//				Disable:     testconvert.Bool(true),
//				Version:     testconvert.String("127"),
//			},
//			want: expected{
//				True:        12.2,
//				False:       13.2,
//				Default:     14.2,
//				Disable:     true,
//				TrackEvents: false,
//				Percentage:  90,
//				Rule:        "test",
//				Version:     "127",
//				RawValues: map[string]string{
//					"Default":     "14.2",
//					"Disable":     "true",
//					"False":       "13.2",
//					"Percentage":  "90.00",
//					"Rollout":     "",
//					"Rule":        "test",
//					"TrackEvents": "false",
//					"True":        "12.2",
//					"Version":     "127",
//				},
//			},
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			assert.Equal(t, tt.want.Disable, tt.flag.IsDisable())
//			assert.Equal(t, tt.want.TrackEvents, tt.flag.IsTrackEvents())
//			assert.Equal(t, tt.want.Version, tt.flag.GetVersion())
//			assert.Equal(t, flagv1.VariationDefault, tt.flag.GetDefaultVariation())
//			assert.Equal(t, tt.want.Default, tt.flag.GetVariationValue(tt.flag.GetDefaultVariation()))
//			assert.Equal(t, tt.want.RawValues, tt.flag.GetRawValues())
//		})
//	}
//}

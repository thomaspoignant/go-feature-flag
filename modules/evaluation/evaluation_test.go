package evaluation_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/modules/core/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/modules/core/model"
	"github.com/thomaspoignant/go-feature-flag/modules/core/testutils/testconvert"
	"github.com/thomaspoignant/go-feature-flag/modules/evaluation"
)

func Test_evaluate(t *testing.T) {
	type args struct {
		flag            flag.Flag
		flagKey         string
		evaluationCtx   ffcontext.Context
		flagCtx         flag.Context
		expectedType    string
		sdkDefaultValue interface{}
	}
	tests := []struct {
		name         string
		args         args
		want         model.VariationResult[interface{}]
		errAssertion assert.ErrorAssertionFunc
	}{
		{
			name: "Test with a rule name",
			args: args{
				flag: &flag.InternalFlag{
					Variations: &map[string]*interface{}{
						"on":  testconvert.Interface(true),
						"off": testconvert.Interface(false),
					},
					DefaultRule: &flag.Rule{
						VariationResult: testconvert.String("on"),
					},
				},
				flagKey:         "my-flag",
				evaluationCtx:   ffcontext.NewEvaluationContext("e22a40e0-28fa-413c-aef8-4dc6a27593b7"),
				flagCtx:         flag.Context{},
				expectedType:    "bool",
				sdkDefaultValue: false,
			},
			want: model.VariationResult[interface{}]{
				TrackEvents:   true,
				VariationType: "on",
				Failed:        false,
				Reason:        flag.ReasonStatic,
				Value:         true,
				Cacheable:     true,
			},
			errAssertion: assert.NoError,
		},
		{
			name: "Test with nil value",
			args: args{
				flag: &flag.InternalFlag{
					Variations: &map[string]*interface{}{
						"on":  nil,
						"off": nil,
					},
					DefaultRule: &flag.Rule{
						VariationResult: testconvert.String("on"),
					},
				},
				flagKey:         "my-flag",
				evaluationCtx:   ffcontext.NewEvaluationContext("e22a40e0-28fa-413c-aef8-4dc6a27593b7"),
				flagCtx:         flag.Context{},
				expectedType:    "bool",
				sdkDefaultValue: false,
			},
			want: model.VariationResult[interface{}]{
				VariationType: "on",
				Failed:        false,
				Reason:        flag.ReasonStatic,
				Value:         nil,
				Cacheable:     true,
				TrackEvents:   true,
			},
			errAssertion: assert.NoError,
		},
		{
			name: "Test with float64 value and int expected",
			args: args{
				flag: &flag.InternalFlag{
					Variations: &map[string]*interface{}{
						"on":  testconvert.Interface(float64(3)),
						"off": testconvert.Interface(float64(0)),
					},
					DefaultRule: &flag.Rule{
						VariationResult: testconvert.String("on"),
					},
				},
				flagKey:         "my-flag",
				evaluationCtx:   ffcontext.NewEvaluationContext("e22a40e0-28fa-413c-aef8-4dc6a27593b7"),
				flagCtx:         flag.Context{},
				expectedType:    "int",
				sdkDefaultValue: 0,
			},
			want: model.VariationResult[interface{}]{
				VariationType: "on",
				Failed:        false,
				Reason:        flag.ReasonStatic,
				Value:         3,
				Cacheable:     true,
				TrackEvents:   true,
			},
			errAssertion: assert.NoError,
		},
		{
			name: "Test with float64 value and int expected",
			args: args{
				flag: &flag.InternalFlag{
					Variations: &map[string]*interface{}{
						"on":  testconvert.Interface(float64(3)),
						"off": testconvert.Interface(float64(0)),
					},
					DefaultRule: &flag.Rule{
						VariationResult: testconvert.String("on"),
					},
				},
				flagKey:         "my-flag",
				evaluationCtx:   ffcontext.NewEvaluationContext("e22a40e0-28fa-413c-aef8-4dc6a27593b7"),
				flagCtx:         flag.Context{},
				expectedType:    "int",
				sdkDefaultValue: 0,
			},
			want: model.VariationResult[interface{}]{
				VariationType: "on",
				Failed:        false,
				Reason:        flag.ReasonStatic,
				Value:         3,
				Cacheable:     true,
				TrackEvents:   true,
			},
			errAssertion: assert.NoError,
		},
		{
			name: "Test with float64 value and float expected",
			args: args{
				flag: &flag.InternalFlag{
					Variations: &map[string]*interface{}{
						"on":  testconvert.Interface(float64(3.5)),
						"off": testconvert.Interface(float64(0.3)),
					},
					DefaultRule: &flag.Rule{
						VariationResult: testconvert.String("on"),
					},
				},
				flagKey:         "my-flag",
				evaluationCtx:   ffcontext.NewEvaluationContext("e22a40e0-28fa-413c-aef8-4dc6a27593b7"),
				flagCtx:         flag.Context{},
				expectedType:    "float",
				sdkDefaultValue: 0,
			},
			want: model.VariationResult[interface{}]{
				VariationType: "on",
				Failed:        false,
				Reason:        flag.ReasonStatic,
				Value:         3.5,
				Cacheable:     true,
				TrackEvents:   true,
			},
			errAssertion: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := evaluation.Evaluate(tt.args.flag, tt.args.flagKey, tt.args.evaluationCtx, tt.args.flagCtx, tt.args.expectedType, tt.args.sdkDefaultValue)
			tt.errAssertion(t, err)
			if err == nil {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func Test_evaluate_typeMisMatch(t *testing.T) {
	f := &flag.InternalFlag{
		Variations: &map[string]*interface{}{
			"on":  testconvert.Interface(true),
			"off": testconvert.Interface(false),
		},
		DefaultRule: &flag.Rule{
			VariationResult: testconvert.String("on"),
		},
	}
	got, err := evaluation.Evaluate[int](f, "my-flag", ffcontext.NewEvaluationContext("e22a40e0-28fa-413c-aef8-4dc6a27593b7"), flag.Context{}, "int", 42)
	assert.Error(t, err)
	want := model.VariationResult[int]{
		VariationType: flag.VariationSDKDefault,
		Failed:        true,
		Reason:        flag.ReasonError,
		ErrorCode:     flag.ErrorCodeTypeMismatch,
		Value:         42,
		TrackEvents:   true,
	}
	assert.Equal(t, want, got)
}

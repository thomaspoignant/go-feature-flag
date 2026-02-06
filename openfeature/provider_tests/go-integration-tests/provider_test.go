package go_integration_tests

import (
	"context"
	"fmt"
	"maps"
	"testing"

	gofeatureflag "github.com/open-feature/go-sdk-contrib/providers/go-feature-flag/pkg"
	of "github.com/open-feature/go-sdk/openfeature"
	"github.com/stretchr/testify/assert"
)

func defaultEvaluationCtx() of.EvaluationContext {
	return of.NewEvaluationContext(
		"d45e303a-38c2-11ed-a261-0242ac120002",
		map[string]any{
			"email":        "john.doe@gofeatureflag.org",
			"firstname":    "john",
			"lastname":     "doe",
			"anonymous":    false,
			"professional": true,
			"rate":         3.14,
			"age":          30,
			"admin":        true,
			"version":      "10.0.0-10",
			"company_info": map[string]any{
				"name": "my_company",
				"size": 120,
			},
			"labels": []string{
				"pro", "beta",
			},
		},
	)
}

func withOverrides(base of.EvaluationContext, overrides map[string]any) of.EvaluationContext {
	attributes := base.Attributes()
	if attributes == nil {
		attributes = make(map[string]any)
	}

	maps.Copy(attributes, overrides)
	return of.NewEvaluationContext("d45e303a-38c2-11ed-a261-0242ac120002", attributes)
}

func TestProvider_module_BooleanEvaluation(t *testing.T) {
	type args struct {
		flag         string
		defaultValue bool
		evalCtx      of.EvaluationContext
	}
	tests := []struct {
		name string
		args args
		want of.BooleanEvaluationDetails
	}{
		{
			name: "should resolve a valid boolean flag with TARGETING_MATCH reason",
			args: args{
				flag:         "bool_targeting_match",
				defaultValue: false,
				evalCtx:      defaultEvaluationCtx(),
			},
			want: of.BooleanEvaluationDetails{
				Value: true,
				EvaluationDetails: of.EvaluationDetails{
					FlagKey:  "bool_targeting_match",
					FlagType: of.Boolean,
					ResolutionDetail: of.ResolutionDetail{
						Variant:      "True",
						Reason:       of.TargetingMatchReason,
						ErrorCode:    "",
						ErrorMessage: "",
						FlagMetadata: map[string]any{
							"description":             "this is a test",
							"gofeatureflag_cacheable": true,
							"pr_link":                 "https://github.com/thomaspoignant/go-feature-flag/pull/916",
						},
					},
				},
			},
		},
		{
			name: "should use boolean default value if the flag is disabled",
			args: args{
				flag:         "disabled_bool",
				defaultValue: false,
				evalCtx:      defaultEvaluationCtx(),
			},
			want: of.BooleanEvaluationDetails{
				Value: false,
				EvaluationDetails: of.EvaluationDetails{
					FlagKey:  "disabled_bool",
					FlagType: of.Boolean,
					ResolutionDetail: of.ResolutionDetail{
						Variant:      "SdkDefault",
						Reason:       of.DisabledReason,
						ErrorCode:    "",
						ErrorMessage: "",
						FlagMetadata: map[string]any{
							"description":             "this is a test",
							"gofeatureflag_cacheable": true,
							"pr_link":                 "https://github.com/thomaspoignant/go-feature-flag/pull/916",
						},
					},
				},
			},
		},
		{
			name: "should error if we expect a boolean and got another type",
			args: args{
				flag:         "string_key",
				defaultValue: false,
				evalCtx:      defaultEvaluationCtx(),
			},
			want: of.BooleanEvaluationDetails{
				Value: false,
				EvaluationDetails: of.EvaluationDetails{
					FlagKey:  "string_key",
					FlagType: of.Boolean,
					ResolutionDetail: of.ResolutionDetail{
						Variant:      "",
						Reason:       of.ErrorReason,
						ErrorCode:    of.TypeMismatchCode,
						ErrorMessage: "resolved value CC0000 is not of boolean type",
						FlagMetadata: map[string]any{},
					},
				},
			},
		},
		{
			name: "should error if flag does not exists",
			args: args{
				flag:         "does_not_exists",
				defaultValue: false,
				evalCtx:      defaultEvaluationCtx(),
			},
			want: of.BooleanEvaluationDetails{
				Value: false,
				EvaluationDetails: of.EvaluationDetails{
					FlagKey:  "does_not_exists",
					FlagType: of.Boolean,
					ResolutionDetail: of.ResolutionDetail{
						Variant:      "",
						Reason:       of.ErrorReason,
						ErrorCode:    of.FlagNotFoundCode,
						ErrorMessage: "flag for key 'does_not_exists' does not exist",
						FlagMetadata: map[string]any{},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := gofeatureflag.NewProvider(gofeatureflag.ProviderOptions{
				Endpoint: "http://localhost:1031/",
			})
			assert.NoError(t, err)
			err = of.SetProviderAndWait(provider)
			assert.NoError(t, err)
			client := of.NewClient("test-app")
			value, err := client.BooleanValueDetails(context.TODO(), tt.args.flag, tt.args.defaultValue, tt.args.evalCtx)

			if tt.want.ErrorCode != "" {
				assert.Error(t, err)
				want := fmt.Sprintf("error code: %s: %s", tt.want.ErrorCode, tt.want.ErrorMessage)
				assert.Equal(t, want, err.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.want, value)
		})
	}
}

func TestProvider_module_StringEvaluation(t *testing.T) {
	type args struct {
		flag         string
		defaultValue string
		evalCtx      of.EvaluationContext
	}
	tests := []struct {
		name string
		args args
		want of.StringEvaluationDetails
	}{
		{
			name: "should resolve a valid string flag with TARGETING_MATCH reason",
			args: args{
				flag:         "string_key",
				defaultValue: "default",
				evalCtx:      defaultEvaluationCtx(),
			},
			want: of.StringEvaluationDetails{
				Value: "CC0000",
				EvaluationDetails: of.EvaluationDetails{
					FlagKey:  "string_key",
					FlagType: of.String,
					ResolutionDetail: of.ResolutionDetail{
						Variant:      "True",
						Reason:       of.TargetingMatchReason,
						ErrorCode:    "",
						ErrorMessage: "",
						FlagMetadata: map[string]any{
							"description":             "this is a test",
							"gofeatureflag_cacheable": true,
							"pr_link":                 "https://github.com/thomaspoignant/go-feature-flag/pull/916",
						},
					},
				},
			},
		},
		{
			name: "should use string default value if the flag is disabled",
			args: args{
				flag:         "disabled_string",
				defaultValue: "default",
				evalCtx:      defaultEvaluationCtx(),
			},
			want: of.StringEvaluationDetails{
				Value: "default",
				EvaluationDetails: of.EvaluationDetails{
					FlagKey:  "disabled_string",
					FlagType: of.String,
					ResolutionDetail: of.ResolutionDetail{
						Variant:      "SdkDefault",
						Reason:       of.DisabledReason,
						ErrorCode:    "",
						ErrorMessage: "",
						FlagMetadata: map[string]any{
							"description":             "this is a test",
							"gofeatureflag_cacheable": true,
							"pr_link":                 "https://github.com/thomaspoignant/go-feature-flag/pull/916",
						},
					},
				},
			},
		},
		{
			name: "should error if we expect a string and got another type",
			args: args{
				flag:         "bool_targeting_match",
				defaultValue: "default",
				evalCtx:      defaultEvaluationCtx(),
			},
			want: of.StringEvaluationDetails{
				Value: "default",
				EvaluationDetails: of.EvaluationDetails{
					FlagKey:  "bool_targeting_match",
					FlagType: of.String,
					ResolutionDetail: of.ResolutionDetail{
						Variant:      "",
						Reason:       of.ErrorReason,
						ErrorCode:    of.TypeMismatchCode,
						ErrorMessage: "resolved value true is not of string type",
						FlagMetadata: map[string]any{},
					},
				},
			},
		},
		{
			name: "should error if flag does not exists",
			args: args{
				flag:         "does_not_exists",
				defaultValue: "default",
				evalCtx:      defaultEvaluationCtx(),
			},
			want: of.StringEvaluationDetails{
				Value: "default",
				EvaluationDetails: of.EvaluationDetails{
					FlagKey:  "does_not_exists",
					FlagType: of.String,
					ResolutionDetail: of.ResolutionDetail{
						Variant:      "",
						Reason:       of.ErrorReason,
						ErrorCode:    of.FlagNotFoundCode,
						ErrorMessage: "flag for key 'does_not_exists' does not exist",
						FlagMetadata: map[string]any{},
					},
				},
			},
		},
		{
			name: "should resolve a targeting match if rule contains filed from evaluationContextEnrichment",
			args: args{
				flag:         "flag-use-evaluation-context-enrichment",
				defaultValue: "default",
				evalCtx:      defaultEvaluationCtx(),
			},
			want: of.StringEvaluationDetails{
				Value: "A",
				EvaluationDetails: of.EvaluationDetails{
					FlagKey:  "flag-use-evaluation-context-enrichment",
					FlagType: of.String,
					ResolutionDetail: of.ResolutionDetail{
						Variant: "A",
						Reason:  of.TargetingMatchReason,
						FlagMetadata: map[string]any{
							"gofeatureflag_cacheable": true,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := gofeatureflag.NewProvider(gofeatureflag.ProviderOptions{
				Endpoint: "http://localhost:1031/",
			})
			assert.NoError(t, err)
			err = of.SetProviderAndWait(provider)
			assert.NoError(t, err)
			client := of.NewClient("test-app")
			value, err := client.StringValueDetails(context.TODO(), tt.args.flag, tt.args.defaultValue, tt.args.evalCtx)

			if tt.want.ErrorCode != "" {
				assert.Error(t, err)
				want := fmt.Sprintf("error code: %s: %s", tt.want.ErrorCode, tt.want.ErrorMessage)
				assert.Equal(t, want, err.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.want, value)
		})
	}
}

func TestProvider_module_FloatEvaluation(t *testing.T) {
	type args struct {
		flag         string
		defaultValue float64
		evalCtx      of.EvaluationContext
	}
	tests := []struct {
		name string
		args args
		want of.FloatEvaluationDetails
	}{
		{
			name: "should resolve a valid float flag with TARGETING_MATCH reason",
			args: args{
				flag:         "double_key",
				defaultValue: 123.45,
				evalCtx:      defaultEvaluationCtx(),
			},
			want: of.FloatEvaluationDetails{
				Value: 100.25,
				EvaluationDetails: of.EvaluationDetails{
					FlagKey:  "double_key",
					FlagType: of.Float,
					ResolutionDetail: of.ResolutionDetail{
						Variant:      "True",
						Reason:       of.TargetingMatchReason,
						ErrorCode:    "",
						ErrorMessage: "",
						FlagMetadata: map[string]any{
							"description":             "this is a test",
							"gofeatureflag_cacheable": true,
							"pr_link":                 "https://github.com/thomaspoignant/go-feature-flag/pull/916",
						},
					},
				},
			},
		},
		{
			name: "should use float default value if the flag is disabled",
			args: args{
				flag:         "disabled_float",
				defaultValue: 123.45,
				evalCtx:      defaultEvaluationCtx(),
			},
			want: of.FloatEvaluationDetails{
				Value: 123.45,
				EvaluationDetails: of.EvaluationDetails{
					FlagKey:  "disabled_float",
					FlagType: of.Float,
					ResolutionDetail: of.ResolutionDetail{
						Variant:      "SdkDefault",
						Reason:       of.DisabledReason,
						ErrorCode:    "",
						ErrorMessage: "",
						FlagMetadata: map[string]any{
							"description":             "this is a test",
							"gofeatureflag_cacheable": true,
							"pr_link":                 "https://github.com/thomaspoignant/go-feature-flag/pull/916",
						},
					},
				},
			},
		},
		{
			name: "should error if we expect a string and got another type",
			args: args{
				flag:         "bool_targeting_match",
				defaultValue: 123.45,
				evalCtx:      defaultEvaluationCtx(),
			},
			want: of.FloatEvaluationDetails{
				Value: 123.45,
				EvaluationDetails: of.EvaluationDetails{
					FlagKey:  "bool_targeting_match",
					FlagType: of.Float,
					ResolutionDetail: of.ResolutionDetail{
						Variant:      "",
						Reason:       of.ErrorReason,
						ErrorCode:    of.TypeMismatchCode,
						ErrorMessage: "resolved value true is not of float type",
						FlagMetadata: map[string]any{},
					},
				},
			},
		},
		{
			name: "should error if flag does not exists",
			args: args{
				flag:         "does_not_exists",
				defaultValue: 123.45,
				evalCtx:      defaultEvaluationCtx(),
			},
			want: of.FloatEvaluationDetails{
				Value: 123.45,
				EvaluationDetails: of.EvaluationDetails{
					FlagKey:  "does_not_exists",
					FlagType: of.Float,
					ResolutionDetail: of.ResolutionDetail{
						Variant:      "",
						Reason:       of.ErrorReason,
						ErrorCode:    of.FlagNotFoundCode,
						ErrorMessage: "flag for key 'does_not_exists' does not exist",
						FlagMetadata: map[string]any{},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := gofeatureflag.NewProvider(gofeatureflag.ProviderOptions{
				Endpoint: "http://localhost:1031/",
			})
			assert.NoError(t, err)
			err = of.SetProviderAndWait(provider)
			assert.NoError(t, err)
			client := of.NewClient("test-app")
			value, err := client.FloatValueDetails(context.TODO(), tt.args.flag, tt.args.defaultValue, tt.args.evalCtx)

			if tt.want.ErrorCode != "" {
				assert.Error(t, err)
				want := fmt.Sprintf("error code: %s: %s", tt.want.ErrorCode, tt.want.ErrorMessage)
				assert.Equal(t, want, err.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.want, value)
		})
	}
}

func TestProvider_module_IntEvaluation(t *testing.T) {
	type args struct {
		flag         string
		defaultValue int64
		evalCtx      of.EvaluationContext
	}
	tests := []struct {
		name string
		args args
		want of.IntEvaluationDetails
	}{
		{
			name: "should resolve a valid float flag with TARGETING_MATCH reason",
			args: args{
				flag:         "integer_key",
				defaultValue: 123,
				evalCtx:      defaultEvaluationCtx(),
			},
			want: of.IntEvaluationDetails{
				Value: 100,
				EvaluationDetails: of.EvaluationDetails{
					FlagKey:  "integer_key",
					FlagType: of.Int,
					ResolutionDetail: of.ResolutionDetail{
						Variant:      "True",
						Reason:       of.TargetingMatchReason,
						ErrorCode:    "",
						ErrorMessage: "",
						FlagMetadata: map[string]any{
							"description":             "this is a test",
							"gofeatureflag_cacheable": true,
							"pr_link":                 "https://github.com/thomaspoignant/go-feature-flag/pull/916",
						},
					},
				},
			},
		},
		{
			name: "should use float default value if the flag is disabled",
			args: args{
				flag:         "disabled_int",
				defaultValue: 123,
				evalCtx:      defaultEvaluationCtx(),
			},
			want: of.IntEvaluationDetails{
				Value: 123,
				EvaluationDetails: of.EvaluationDetails{
					FlagKey:  "disabled_int",
					FlagType: of.Int,
					ResolutionDetail: of.ResolutionDetail{
						Variant:      "SdkDefault",
						Reason:       of.DisabledReason,
						ErrorCode:    "",
						ErrorMessage: "",
						FlagMetadata: map[string]any{
							"description":             "this is a test",
							"gofeatureflag_cacheable": true,
							"pr_link":                 "https://github.com/thomaspoignant/go-feature-flag/pull/916",
						},
					},
				},
			},
		},
		{
			name: "should error if we expect a string and got another type",
			args: args{
				flag:         "bool_targeting_match",
				defaultValue: 123,
				evalCtx:      defaultEvaluationCtx(),
			},
			want: of.IntEvaluationDetails{
				Value: 123,
				EvaluationDetails: of.EvaluationDetails{
					FlagKey:  "bool_targeting_match",
					FlagType: of.Int,
					ResolutionDetail: of.ResolutionDetail{
						Variant:      "",
						Reason:       of.ErrorReason,
						ErrorCode:    of.TypeMismatchCode,
						ErrorMessage: "resolved value true is not of integer type",
						FlagMetadata: map[string]any{},
					},
				},
			},
		},
		{
			name: "should error if flag does not exists",
			args: args{
				flag:         "does_not_exists",
				defaultValue: 123,
				evalCtx:      defaultEvaluationCtx(),
			},
			want: of.IntEvaluationDetails{
				Value: 123,
				EvaluationDetails: of.EvaluationDetails{
					FlagKey:  "does_not_exists",
					FlagType: of.Int,
					ResolutionDetail: of.ResolutionDetail{
						Variant:      "",
						Reason:       of.ErrorReason,
						ErrorCode:    of.FlagNotFoundCode,
						ErrorMessage: "flag for key 'does_not_exists' does not exist",
						FlagMetadata: map[string]any{},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := gofeatureflag.NewProvider(gofeatureflag.ProviderOptions{
				Endpoint: "http://localhost:1031/",
			})
			assert.NoError(t, err)
			err = of.SetProviderAndWait(provider)
			assert.NoError(t, err)
			client := of.NewClient("test-app")
			value, err := client.IntValueDetails(context.TODO(), tt.args.flag, tt.args.defaultValue, tt.args.evalCtx)

			if tt.want.ErrorCode != "" {
				assert.Error(t, err)
				want := fmt.Sprintf("error code: %s: %s", tt.want.ErrorCode, tt.want.ErrorMessage)
				assert.Equal(t, want, err.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.want, value)
		})
	}
}

func TestProvider_module_ObjectEvaluation(t *testing.T) {
	type args struct {
		flag         string
		defaultValue any
		evalCtx      of.EvaluationContext
	}
	tests := []struct {
		name string
		args args
		want of.InterfaceEvaluationDetails
	}{
		{
			name: "should resolve a valid interface flag with TARGETING_MATCH reason",
			args: args{
				flag:         "object_key",
				defaultValue: nil,
				evalCtx:      defaultEvaluationCtx(),
			},
			want: of.InterfaceEvaluationDetails{
				Value: map[string]any{
					"test":  "test1",
					"test2": false,
					"test3": 123.3,
					"test4": float64(1),
				},
				EvaluationDetails: of.EvaluationDetails{
					FlagKey:  "object_key",
					FlagType: of.Object,
					ResolutionDetail: of.ResolutionDetail{
						Variant:      "True",
						Reason:       of.TargetingMatchReason,
						ErrorCode:    "",
						ErrorMessage: "",
						FlagMetadata: map[string]any{
							"description":             "this is a test",
							"gofeatureflag_cacheable": true,
							"pr_link":                 "https://github.com/thomaspoignant/go-feature-flag/pull/916",
						},
					},
				},
			},
		},
		{
			name: "should use interface default value if the flag is disabled",
			args: args{
				flag:         "disabled_interface",
				defaultValue: nil,
				evalCtx:      defaultEvaluationCtx(),
			},
			want: of.InterfaceEvaluationDetails{
				Value: nil,
				EvaluationDetails: of.EvaluationDetails{
					FlagKey:  "disabled_interface",
					FlagType: of.Object,
					ResolutionDetail: of.ResolutionDetail{
						Variant:      "SdkDefault",
						Reason:       of.DisabledReason,
						ErrorCode:    "",
						ErrorMessage: "",
						FlagMetadata: map[string]any{
							"description":             "this is a test",
							"gofeatureflag_cacheable": true,
							"pr_link":                 "https://github.com/thomaspoignant/go-feature-flag/pull/916",
						},
					},
				},
			},
		},
		{
			name: "should error if flag does not exists",
			args: args{
				flag:         "does_not_exists",
				defaultValue: nil,
				evalCtx:      defaultEvaluationCtx(),
			},
			want: of.InterfaceEvaluationDetails{
				Value: nil,
				EvaluationDetails: of.EvaluationDetails{
					FlagKey:  "does_not_exists",
					FlagType: of.Object,
					ResolutionDetail: of.ResolutionDetail{
						Variant:      "",
						Reason:       of.ErrorReason,
						ErrorCode:    of.FlagNotFoundCode,
						ErrorMessage: "flag for key 'does_not_exists' does not exist",
						FlagMetadata: map[string]any{},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := gofeatureflag.NewProvider(gofeatureflag.ProviderOptions{
				Endpoint: "http://localhost:1031/",
			})
			assert.NoError(t, err)
			err = of.SetProviderAndWait(provider)
			assert.NoError(t, err)
			client := of.NewClient("test-app")
			value, err := client.ObjectValueDetails(context.TODO(), tt.args.flag, tt.args.defaultValue, tt.args.evalCtx)

			if tt.want.ErrorCode != "" {
				assert.Error(t, err)
				want := fmt.Sprintf("error code: %s: %s", tt.want.ErrorCode, tt.want.ErrorMessage)
				assert.Equal(t, want, err.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.want, value)
		})
	}
}

func TestProvider_apikey_relay_proxy(t *testing.T) {
	type args struct {
		apiKey string
	}
	tests := []struct {
		name string
		args args
		want of.BooleanEvaluationDetails
	}{
		{
			name: "should resolve a valid flag with an apiKey",
			args: args{
				apiKey: "authorized_token",
			},
			want: of.BooleanEvaluationDetails{
				Value: true,
				EvaluationDetails: of.EvaluationDetails{
					FlagKey:  "bool_targeting_match",
					FlagType: of.Boolean,
					ResolutionDetail: of.ResolutionDetail{
						Variant:      "True",
						Reason:       of.TargetingMatchReason,
						ErrorCode:    "",
						ErrorMessage: "",
						FlagMetadata: map[string]any{
							"description":             "this is a test",
							"gofeatureflag_cacheable": true,
							"pr_link":                 "https://github.com/thomaspoignant/go-feature-flag/pull/916",
						},
					},
				},
			},
		},
		{
			name: "should resolve a default value with an invalid apiKey",
			args: args{
				apiKey: "invalid_token",
			},
			want: of.BooleanEvaluationDetails{
				Value: false,
				EvaluationDetails: of.EvaluationDetails{
					FlagKey:  "bool_targeting_match",
					FlagType: of.Boolean,
					ResolutionDetail: of.ResolutionDetail{
						Variant:      "",
						Reason:       of.ErrorReason,
						ErrorCode:    of.GeneralCode,
						ErrorMessage: "authentication/authorization error",
						FlagMetadata: map[string]any{},
					},
				},
			},
		},
		{
			name: "should resolve a default value with no apiKey",
			args: args{
				apiKey: "",
			},
			want: of.BooleanEvaluationDetails{
				Value: false,
				EvaluationDetails: of.EvaluationDetails{
					FlagKey:  "bool_targeting_match",
					FlagType: of.Boolean,
					ResolutionDetail: of.ResolutionDetail{
						Variant:      "",
						Reason:       of.ErrorReason,
						ErrorCode:    of.GeneralCode,
						ErrorMessage: "authentication/authorization error",
						FlagMetadata: map[string]any{},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := gofeatureflag.NewProvider(gofeatureflag.ProviderOptions{
				Endpoint: "http://localhost:1032/",
				APIKey:   tt.args.apiKey,
			})
			assert.NoError(t, err)
			err = of.SetProviderAndWait(provider)
			assert.NoError(t, err)
			client := of.NewClient("test-app")
			value, err := client.BooleanValueDetails(context.TODO(), "bool_targeting_match", false, defaultEvaluationCtx())

			if tt.want.ErrorCode != "" {
				assert.Error(t, err)
				want := fmt.Sprintf("error code: %s: %s", tt.want.ErrorCode, tt.want.ErrorMessage)
				assert.Equal(t, want, err.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.want, value)
		})
	}
}

func TestProvider_rules_semverEvaluation(t *testing.T) {
	type args struct {
		flag         string
		defaultValue bool
		evalCtx      of.EvaluationContext
	}
	tests := []struct {
		name string
		args args
		want of.BooleanEvaluationDetails
	}{
		{
			name: "should resolve with TARGETING_MATCH reason for valid semver match",
			args: args{
				flag:         "boolean_semver_targeting_match",
				evalCtx:      defaultEvaluationCtx(),
			},
			want: of.BooleanEvaluationDetails{
				Value: true,
				EvaluationDetails: of.EvaluationDetails{
					FlagKey:  "boolean_semver_targeting_match",
					FlagType: of.Boolean,
					ResolutionDetail: of.ResolutionDetail{
						Variant:      "True",
						Reason:       of.TargetingMatchReason,
						ErrorCode:    "",
						ErrorMessage: "",
						FlagMetadata: map[string]any{
							"description":             "this is a semver matching test",
							"gofeatureflag_cacheable": true,
							"pr_link":                 "https://github.com/thomaspoignant/go-feature-flag/pull/4764",
						},
					},
				},
			},
		},
		{
			name: "should resolve flag with DEFAULT reason for invalid semver match",
			args: args{
				flag:         "boolean_semver_targeting_match",
				evalCtx: withOverrides(defaultEvaluationCtx(), map[string]any{
					"version": "10.0.0-2",
				}),
			},
			want: of.BooleanEvaluationDetails{
				Value: false,
				EvaluationDetails: of.EvaluationDetails{
					FlagKey:  "boolean_semver_targeting_match",
					FlagType: of.Boolean,
					ResolutionDetail: of.ResolutionDetail{
						Variant:      "False",
						Reason:       of.DefaultReason,
						ErrorCode:    "",
						ErrorMessage: "",
						FlagMetadata: map[string]any{
							"description":             "this is a semver matching test",
							"gofeatureflag_cacheable": true,
							"pr_link":                 "https://github.com/thomaspoignant/go-feature-flag/pull/4764",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := gofeatureflag.NewProvider(gofeatureflag.ProviderOptions{
				Endpoint: "http://localhost:1031/",
			})
			assert.NoError(t, err)
			err = of.SetProviderAndWait(provider)
			assert.NoError(t, err)
			client := of.NewClient("test-app")
			value, err := client.BooleanValueDetails(context.TODO(), tt.args.flag, tt.args.defaultValue, tt.args.evalCtx)

			if tt.want.ErrorCode != "" {
				assert.Error(t, err)
				want := fmt.Sprintf("error code: %s: %s", tt.want.ErrorCode, tt.want.ErrorMessage)
				assert.Equal(t, want, err.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.want, value)
		})
	}
}

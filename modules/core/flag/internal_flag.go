package flag

import (
	"encoding/json"
	"fmt"
	"maps"
	"time"

	"github.com/thomaspoignant/go-feature-flag/modules/core/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/modules/core/internalerror"
	"github.com/thomaspoignant/go-feature-flag/modules/core/utils"
)

const (
	PercentageMultiplier = float64(1000)
)

// InternalFlag is the internal representation of a flag when using go-feature-flag.
// All the flags in your configuration files can have different format but will be
// converted into an InternalFlag to be used in the library.
type InternalFlag struct {
	// Variations are all the variations available for this flag. You can have as many variation as needed.
	Variations *map[string]*interface{} `json:"variations,omitempty" yaml:"variations,omitempty" toml:"variations,omitempty"` // nolint:lll

	// Rules is the list of Rule for this flag.
	// This an optional field.
	Rules *[]Rule `json:"targeting,omitempty" yaml:"targeting,omitempty" toml:"targeting,omitempty"`

	// BucketingKey defines a source for a dynamic targeting key
	BucketingKey *string `json:"bucketingKey,omitempty" yaml:"bucketingKey,omitempty" toml:"bucketingKey,omitempty"`

	// DefaultRule is the originalRule applied after checking that any other rules
	// matched the user.
	DefaultRule *Rule `json:"defaultRule,omitempty" yaml:"defaultRule,omitempty" toml:"defaultRule,omitempty"`

	// Experimentation is your struct to configure an experimentation, it will allow you to configure a start date and
	// an end date for your flag.
	// When the experimentation is not running, the flag will serve the default value.
	Experimentation *ExperimentationRollout `json:"experimentation,omitempty" yaml:"experimentation,omitempty" toml:"experimentation,omitempty"` // nolint: lll

	// Scheduled is your struct to configure an update on some fields of your flag over time.
	// You can add several steps that updates the flag, this is typically used if you want to gradually add more user
	// in your flag.
	Scheduled *[]ScheduledStep `json:"scheduledRollout,omitempty" yaml:"scheduledRollout,omitempty" toml:"scheduledRollout,omitempty"` // nolint: lll

	// TrackEvents is false if you don't want to export the data in your data exporter.
	// Default value is true
	TrackEvents *bool `json:"trackEvents,omitempty" yaml:"trackEvents,omitempty" toml:"trackEvents,omitempty"`

	// Disable is true if the flag is disabled.
	Disable *bool `json:"disable,omitempty" yaml:"disable,omitempty" toml:"disable,omitempty"`

	// Version (optional) This field contains the version of the flag.
	// The version is manually managed when you configure your flags, and it is used to display the information
	// in the notifications and data collection.
	Version *string `json:"version,omitempty" yaml:"version,omitempty" toml:"version,omitempty"`

	// Metadata is a field containing information about your flag such as an issue tracker link, a description, etc ...
	Metadata *map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty" toml:"metadata,omitempty"`
}

// Value is returning the Value associate to the flag
func (f *InternalFlag) Value(
	flagName string,
	evaluationCtx ffcontext.Context,
	flagContext Context,
) (interface{}, ResolutionDetails) {
	// if the evaluation context is nil, we create a new one with an empty key
	// this is to avoid any nil pointer exception.
	if evaluationCtx == nil {
		evaluationCtx = ffcontext.NewEvaluationContext("")
	}

	evaluationDate := DateFromContextOrDefault(evaluationCtx, time.Now())
	flag, err := f.applyScheduledRolloutSteps(evaluationDate)
	if err != nil {
		return flagContext.DefaultSdkValue, ResolutionDetails{
			Variant:      VariationSDKDefault,
			Reason:       ReasonError,
			ErrorCode:    ErrorCodeGeneral,
			ErrorMessage: err.Error(),
		}
	}

	if flagContext.EvaluationContextEnrichment != nil {
		maps.Copy(evaluationCtx.GetCustom(), flagContext.EvaluationContextEnrichment)
	}

	key, keyError := flag.GetBucketingKeyValue(evaluationCtx)
	if keyError != nil {
		return flagContext.DefaultSdkValue, ResolutionDetails{
			Variant:      VariationSDKDefault,
			Reason:       ReasonError,
			ErrorCode:    ErrorCodeTargetingKeyMissing,
			ErrorMessage: keyError.Error(),
			Metadata:     f.GetMetadata(),
		}
	}

	if flag.IsDisable() || flag.isExperimentationOver(evaluationDate) {
		return flagContext.DefaultSdkValue, ResolutionDetails{
			Variant:   VariationSDKDefault,
			Reason:    ReasonDisabled,
			Cacheable: f.isCacheable(),
			Metadata:  f.GetMetadata(),
		}
	}

	variationSelection, err := flag.selectVariation(flagName, key, evaluationCtx)
	if err != nil {
		return flagContext.DefaultSdkValue,
			ResolutionDetails{
				Variant:      VariationSDKDefault,
				Reason:       ReasonError,
				ErrorCode:    ErrorFlagConfiguration,
				ErrorMessage: err.Error(),
				Metadata:     flag.GetMetadata(),
			}
	}

	return flag.GetVariationValue(variationSelection.name), ResolutionDetails{
		Variant:   variationSelection.name,
		Reason:    variationSelection.reason,
		RuleIndex: variationSelection.ruleIndex,
		RuleName:  variationSelection.ruleName,
		Cacheable: variationSelection.cacheable,
		Metadata:  constructMetadata(flag.GetMetadata(), variationSelection.ruleName),
	}
}

// selectEvaluationReason is choosing which reason has been chosen for the evaluation.
func selectEvaluationReason(
	hasRule bool,
	targetingMatch bool,
	isDynamic bool,
	isDefaultRule bool,
) ResolutionReason {
	if hasRule && targetingMatch {
		if isDynamic {
			return ReasonTargetingMatchSplit
		}
		return ReasonTargetingMatch
	}

	if isDefaultRule {
		if isDynamic {
			return ReasonSplit
		}
		if hasRule {
			return ReasonDefault
		}
		return ReasonStatic
	}
	return ReasonUnknown
}

func (f *InternalFlag) isCacheable() bool {
	isDynamic := (f.Scheduled != nil && len(*f.Scheduled) > 0) || f.Experimentation != nil
	return !isDynamic
}

// selectVariation is doing the magic to select the variation that should be used for this specific user
// to always affect the user to the same segment we are using a hash of the flag name + key
func (f *InternalFlag) selectVariation(
	flagName string, key string, ctx ffcontext.Context) (*variationSelection, error) {
	hasRule := len(f.GetRules()) != 0
	// Check all targeting in order, the first to match will be the one used.
	if hasRule {
		for ruleIndex, target := range f.GetRules() {
			variationName, err := target.Evaluate(key, ctx, flagName, false)
			if err != nil {
				// the targeting does not apply
				if _, ok := err.(*internalerror.RuleNotApplyError); ok {
					continue
				}
				return nil, err
			}
			reason := selectEvaluationReason(hasRule, true, target.IsDynamic(), false)
			return &variationSelection{
				name:      variationName,
				reason:    reason,
				ruleIndex: &ruleIndex,
				ruleName:  f.GetRules()[ruleIndex].Name,
				cacheable: f.isCacheable() && target.ProgressiveRollout == nil,
			}, err
		}
	}

	if f.DefaultRule == nil {
		return nil, fmt.Errorf("no default targeting for the flag")
	}

	variationName, err := f.GetDefaultRule().Evaluate(key, ctx, flagName, true)
	if err != nil {
		return nil, err
	}

	reason := selectEvaluationReason(hasRule, false, f.GetDefaultRule().IsDynamic(), true)
	return &variationSelection{
		name:      variationName,
		reason:    reason,
		cacheable: f.isCacheable() && f.GetDefaultRule().ProgressiveRollout == nil,
	}, nil
}

// nolint: gocognit
// applyScheduledRolloutSteps is checking if the flag has a scheduled rollout configured.
// If yes we merge the changes to the current flag.
func (f *InternalFlag) applyScheduledRolloutSteps(evaluationDate time.Time) (*InternalFlag, error) {
	if f.Scheduled == nil {
		return f, nil
	}

	// We are doing a deep copy the flag to avoid modifying the original flag.
	// The deep copy is done to fix this issue https://github.com/thomaspoignant/go-feature-flag/issues/2256
	data, err := json.Marshal(f)
	if err != nil {
		return &InternalFlag{}, err
	}
	var flagCopy *InternalFlag
	if err := json.Unmarshal(data, &flagCopy); err != nil {
		return &InternalFlag{}, err
	}

	// We apply the scheduled rollout
	for _, steps := range *f.Scheduled {
		if steps.Date != nil &&
			(steps.Date.Before(evaluationDate) || steps.Date.Equal(evaluationDate)) {
			flagCopy.Rules = MergeSetOfRules(f.GetRules(), steps.GetRules())
			if steps.Disable != nil {
				flagCopy.Disable = steps.Disable
			}

			if steps.TrackEvents != nil {
				flagCopy.TrackEvents = steps.TrackEvents
			}

			if steps.DefaultRule != nil {
				flagCopy.DefaultRule.MergeRules(*steps.DefaultRule)
			}

			if steps.Variations != nil {
				for key, value := range steps.GetVariations() {
					flagCopy.GetVariations()[key] = value
				}
			}

			if steps.Version != nil {
				flagCopy.Version = steps.Version
			}

			if steps.Experimentation != nil {
				if flagCopy.Experimentation == nil {
					flagCopy.Experimentation = &ExperimentationRollout{}
				}
				if steps.Experimentation.Start != nil {
					flagCopy.Experimentation.Start = steps.Experimentation.Start
				}
				if steps.Experimentation.End != nil {
					flagCopy.Experimentation.End = steps.Experimentation.End
				}
			}
		}
	}
	return flagCopy, nil
}

// isExperimentationOver checks if we are in an experimentation or not
func (f *InternalFlag) isExperimentationOver(evaluationDate time.Time) bool {
	return f.Experimentation != nil &&
		((f.Experimentation.Start != nil && evaluationDate.Before(*f.Experimentation.Start)) ||
			(f.Experimentation.End != nil && evaluationDate.After(*f.Experimentation.End)))
}

// IsValid is checking if the current flag is valid.
func (f *InternalFlag) IsValid() error {
	if len(f.GetVariations()) == 0 {
		return fmt.Errorf("no variation available")
	}

	// Check that all variation has the same types
	expectedVarType := ""
	for name, value := range f.GetVariations() {
		if value == nil {
			return fmt.Errorf("nil value for variation: %s", name)
		}
		if expectedVarType != "" {
			currentType, err := utils.JSONTypeExtractor(*value)
			if err != nil {
				return err
			}
			if currentType != expectedVarType {
				return fmt.Errorf("all variations should have the same type")
			}
		} else {
			var err error
			expectedVarType, err = utils.JSONTypeExtractor(*value)
			if err != nil {
				return err
			}
		}
	}

	// Validate that we have a default Rule
	if f.GetDefaultRule() == nil {
		return fmt.Errorf("missing default rule")
	}

	const isDefaultRule = true
	// Validate rules
	if err := f.GetDefaultRule().IsValid(isDefaultRule, f.GetVariations()); err != nil {
		return err
	}

	ruleNames := map[string]interface{}{}
	for _, rule := range f.GetRules() {
		if err := rule.IsValid(!isDefaultRule, f.GetVariations()); err != nil {
			return err
		}

		// Check if we have duplicated rule name
		if _, ok := ruleNames[rule.GetName()]; ok && rule.GetName() != "" {
			return fmt.Errorf("duplicated rule name: %s", rule.GetName())
		} else if rule.GetName() != "" {
			ruleNames[rule.GetName()] = nil
		}
	}

	return nil
}

// GetVariations is the getter of the field Variations
func (f *InternalFlag) GetVariations() map[string]*interface{} {
	if f.Variations == nil {
		return map[string]*interface{}{}
	}
	return *f.Variations
}

func (f *InternalFlag) GetRules() []Rule {
	if f.Rules == nil {
		return []Rule{}
	}
	return *f.Rules
}

func (f *InternalFlag) GetRuleIndexByName(name string) *int {
	for index, rule := range f.GetRules() {
		if rule.GetName() == name {
			return &index
		}
	}
	return nil
}

// GetDefaultRule is the getter of the field DefaultRule
func (f *InternalFlag) GetDefaultRule() *Rule {
	return f.DefaultRule
}

// IsTrackEvents is the getter of the field TrackEvents
func (f *InternalFlag) IsTrackEvents() bool {
	if f.TrackEvents == nil {
		return true
	}
	return *f.TrackEvents
}

// IsDisable is the getter for the field Disable
func (f *InternalFlag) IsDisable() bool {
	if f.Disable == nil {
		return false
	}
	return *f.Disable
}

// GetVersion is the getter for the field Version
func (f *InternalFlag) GetVersion() string {
	if f.Version == nil {
		return ""
	}
	return *f.Version
}

// GetVariationValue return the value of variation from his name
func (f *InternalFlag) GetVariationValue(name string) interface{} {
	if v, exists := f.GetVariations()[name]; exists && v != nil {
		return *v
	}
	return nil
}

// GetBucketingKey return the name of the custom bucketing key if we are using one.
func (f *InternalFlag) GetBucketingKey() string {
	if f.BucketingKey == nil {
		return ""
	}
	return *f.BucketingKey
}

// RequiresBucketing checks if the flag requires a bucketing key for evaluation
// A flag requires bucketing if it has percentage-based rules or progressive rollouts,
// including those introduced by scheduled rollout steps
func (f *InternalFlag) RequiresBucketing() bool {
	// Check if default rule requires bucketing
	if f.DefaultRule != nil && f.DefaultRule.RequiresBucketing() {
		return true
	}

	// Check if any targeting rule requires bucketing
	for _, rule := range f.GetRules() {
		if rule.RequiresBucketing() {
			return true
		}
	}

	// Check if any scheduled rollout steps introduce bucketing requirements
	if f.Scheduled != nil {
		for _, step := range *f.Scheduled {
			// Check if the scheduled step's default rule requires bucketing
			if step.DefaultRule != nil && step.DefaultRule.RequiresBucketing() {
				return true
			}

			// Check if any of the scheduled step's rules require bucketing
			for _, rule := range step.GetRules() {
				if rule.RequiresBucketing() {
					return true
				}
			}
		}
	}

	return false
}

// GetBucketingKeyValue return the value of the bucketing key from the context
// If requiresBucketing is false, it allows empty keys for flags that don't need them
func (f *InternalFlag) GetBucketingKeyValue(ctx ffcontext.Context) (string, error) {
	// Cache the bucketing requirement check to avoid multiple calls
	requiresBucketing := f.RequiresBucketing()

	// Check if custom bucketing key is provided
	if f.BucketingKey != nil {
		key := f.GetBucketingKey()
		if key == "" {
			return ctx.GetKey(), nil
		}

		value, err := utils.GetNestedFieldValue(ctx.GetCustom(), key)
		if err != nil {
			return "", fmt.Errorf("impossible to find bucketingKey in context: %w", err)
		}

		switch v := value.(type) {
		case string:
			if v == "" {
				if requiresBucketing {
					return "", &internalerror.EmptyBucketingKeyError{Message: "Empty bucketing key"}
				}
				// Return empty key if bucketing not required
				return "", nil
			}
			return v, nil
		default:
			return "", fmt.Errorf("invalid bucketing key")
		}
	}

	// Check if targeting key is required for this flag
	if ctx.GetKey() == "" {
		if requiresBucketing {
			return "", &internalerror.EmptyBucketingKeyError{Message: "Empty targeting key"}
		}
		// Return empty key if bucketing not required
		return "", nil
	}

	return ctx.GetKey(), nil
}

// GetMetadata return the metadata associated to the flag
func (f *InternalFlag) GetMetadata() map[string]interface{} {
	if f.Metadata == nil {
		return nil
	}
	return *f.Metadata
}

func DateFromContextOrDefault(ctx ffcontext.Context, defaultDate time.Time) time.Time {
	if ctx == nil || ctx.ExtractGOFFProtectedFields().CurrentDateTime == nil {
		return defaultDate
	}
	return *ctx.ExtractGOFFProtectedFields().CurrentDateTime
}

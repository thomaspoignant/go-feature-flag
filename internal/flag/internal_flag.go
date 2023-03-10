package flag

import (
	"fmt"
	"time"

	"github.com/thomaspoignant/go-feature-flag/ffuser"
	"github.com/thomaspoignant/go-feature-flag/internal/internalerror"
	"github.com/thomaspoignant/go-feature-flag/internal/utils"
)

const (
	PercentageMultiplier = float64(1000)
	MaxPercentage        = uint32(100 * PercentageMultiplier)
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
}

// Value is returning the Value associate to the flag
func (f *InternalFlag) Value(
	flagName string,
	user ffuser.User,
	evaluationCtx EvaluationContext,
) (interface{}, ResolutionDetails) {
	f.applyScheduledRolloutSteps()

	if f.IsDisable() || f.isExperimentationOver() {
		return evaluationCtx.DefaultSdkValue, ResolutionDetails{
			Variant:   VariationSDKDefault,
			Reason:    ReasonDisabled,
			Cacheable: f.isCacheable(),
		}
	}

	variationSelection, err := f.selectVariation(flagName, user)
	if err != nil {
		return evaluationCtx.DefaultSdkValue,
			ResolutionDetails{
				Variant:   VariationSDKDefault,
				Reason:    ReasonError,
				ErrorCode: ErrorFlagConfiguration,
			}
	}

	return f.GetVariationValue(variationSelection.name), ResolutionDetails{
		Variant:   variationSelection.name,
		Reason:    variationSelection.reason,
		RuleIndex: variationSelection.ruleIndex,
		RuleName:  variationSelection.ruleName,
		Cacheable: variationSelection.cacheable,
	}
}

// selectEvaluationReason is choosing which reason has been chosen for the evaluation.
func selectEvaluationReason(hasRule bool, targetingMatch bool, isDynamic bool, isDefaultRule bool) ResolutionReason {
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
func (f *InternalFlag) selectVariation(flagName string, user ffuser.User) (*variationSelection, error) {
	hashID := utils.Hash(flagName+user.GetKey()) % MaxPercentage
	hasRule := len(f.GetRules()) != 0
	// Check all targeting in order, the first to match will be the one used.
	if hasRule {
		for ruleIndex, target := range f.GetRules() {
			variationName, err := target.Evaluate(user, hashID, false)
			if err != nil {
				// the targeting does not apply
				if _, ok := err.(*internalerror.RuleNotApply); ok {
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

	variationName, err := f.GetDefaultRule().Evaluate(user, hashID, true)
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
func (f *InternalFlag) applyScheduledRolloutSteps() {
	evaluationDate := time.Now()
	if f.Scheduled != nil {
		for _, steps := range *f.Scheduled {
			if steps.Date != nil && steps.Date.Before(evaluationDate) {
				f.Rules = MergeSetOfRules(f.GetRules(), steps.GetRules())
				if steps.Disable != nil {
					f.Disable = steps.Disable
				}

				if steps.TrackEvents != nil {
					f.TrackEvents = steps.TrackEvents
				}

				if steps.DefaultRule != nil {
					f.DefaultRule.MergeRules(*steps.DefaultRule)
				}

				if steps.Variations != nil {
					for key, value := range steps.GetVariations() {
						f.GetVariations()[key] = value
					}
				}

				if steps.Version != nil {
					f.Version = steps.Version
				}

				if steps.Experimentation != nil {
					if f.Experimentation == nil {
						f.Experimentation = &ExperimentationRollout{}
					}
					if steps.Experimentation.Start != nil {
						f.Experimentation.End = steps.Experimentation.End
					}
					if steps.Experimentation.End != nil {
						f.Experimentation.End = steps.Experimentation.End
					}
				}
			}
		}
	}
}

// isExperimentationOver checks if we are in an experimentation or not
func (f *InternalFlag) isExperimentationOver() bool {
	now := time.Now()
	return f.Experimentation != nil &&
		((f.Experimentation.Start != nil && now.Before(*f.Experimentation.Start)) ||
			(f.Experimentation.End != nil && now.After(*f.Experimentation.End)))
}

// IsValid is checking if the current flag is valid.
func (f *InternalFlag) IsValid() error {
	if len(f.GetVariations()) == 0 {
		return fmt.Errorf("no variation available")
	}

	// Check that all variation have the same types
	expectedVarType := ""
	for _, value := range f.GetVariations() {
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
	if err := f.GetDefaultRule().IsValid(isDefaultRule); err != nil {
		return err
	}

	ruleNames := map[string]interface{}{}
	for _, rule := range f.GetRules() {
		if err := rule.IsValid(!isDefaultRule); err != nil {
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
	for k, v := range f.GetVariations() {
		if k == name && v != nil {
			return *v
		}
	}
	return nil
}

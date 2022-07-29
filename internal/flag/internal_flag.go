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
	// Variations are all the variations available for this flag. The minimum is 2 variations and, we don't have any max
	// limit except if the variationValue is a bool, the max is 2.
	Variations *map[string]*interface{} `json:"variations,omitempty" yaml:"variations,omitempty" toml:"variations,omitempty"` // nolint:lll

	// Rules is the list of Rule for this flag.
	// This an optional field.
	Rules *[]Rule `json:"targeting,omitempty" yaml:"targeting,omitempty" toml:"targeting,omitempty"`

	// DefaultRule is the originalRule applied after checking that any other rules
	// matched the user.
	DefaultRule *Rule `json:"defaultRule,omitempty" yaml:"defaultRule,omitempty" toml:"defaultRule,omitempty"`

	// Rollout is how we roll out the flag
	Rollout *Rollout `json:"rollout,omitempty" yaml:"rollout,omitempty" toml:"rollout,omitempty"`

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

func (f *InternalFlag) String() string {
	panic("implement me")
}

// Value is returning the Value associate to the flag
func (f *InternalFlag) Value(
	flagName string,
	user ffuser.User,
	evaluationCtx EvaluationContext,
) (interface{}, ResolutionDetails) {
	f.applyScheduledRolloutSteps()

	if f.IsDisable() || f.isExperimentationOver() {
		return evaluationCtx.DefaultSdkValue, ResolutionDetails{Variant: VariationSDKDefault, Reason: ReasonDisabled}
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
	}
}

// selectVariation is doing the magic to select the variation that should be used for this specific user
// to always affect the user to the same segment we are using a hash of the flag name + key
func (f *InternalFlag) selectVariation(flagName string, user ffuser.User) (*variationSelection, error) {
	hashID := utils.Hash(flagName+user.GetKey()) % MaxPercentage

	// Check all targeting in order, the first to match will be the one used.
	if len(f.GetRules()) != 0 {
		for ruleIndex, target := range f.GetRules() {
			variationName, err := target.Evaluate(user, hashID, false)
			if err != nil {
				// the targeting does not apply
				if _, ok := err.(*internalerror.RuleNotApply); ok {
					continue
				}
				return nil, err
			}
			return &variationSelection{
				name:      variationName,
				reason:    ReasonTargetingMatch,
				ruleIndex: &ruleIndex,
				ruleName:  f.GetRules()[ruleIndex].Name,
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

	return &variationSelection{name: variationName, reason: ReasonDefault}, nil
}

// nolint: gocognit
// applyScheduledRolloutSteps is checking if the flag has a scheduled rollout configured.
// If yes we merge the changes to the current flag.
func (f *InternalFlag) applyScheduledRolloutSteps() {
	evaluationDate := time.Now()
	if f.Rollout != nil && f.Rollout.Scheduled != nil {
		for _, steps := range *f.Rollout.Scheduled {
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

				if steps.Rollout != nil && steps.Rollout.Experimentation != nil {
					if f.Rollout.Experimentation == nil {
						f.Rollout.Experimentation = &ExperimentationRollout{}
					}
					if steps.Rollout.Experimentation.Start != nil {
						f.Rollout.Experimentation.End = steps.Rollout.Experimentation.End
					}
					if steps.Rollout.Experimentation.End != nil {
						f.Rollout.Experimentation.End = steps.Rollout.Experimentation.End
					}
				}
			}
		}
	}
}

// isExperimentationOver checks if we are in an experimentation or not
func (f *InternalFlag) isExperimentationOver() bool {
	now := time.Now()
	return f.Rollout != nil &&
		f.Rollout.Experimentation != nil &&
		((f.Rollout.Experimentation.Start != nil && now.Before(*f.Rollout.Experimentation.Start)) ||
			(f.Rollout.Experimentation.End != nil && now.After(*f.Rollout.Experimentation.End)))
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

// GetRollout is the getter for the field Rollout
func (f *InternalFlag) GetRollout() *Rollout {
	return f.Rollout
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

package flag

import (
	"fmt"
	"github.com/thomaspoignant/go-feature-flag/ffuser"
	"github.com/thomaspoignant/go-feature-flag/internal/constant"
	"github.com/thomaspoignant/go-feature-flag/internal/utils"
	"sort"
	"strings"
	"time"
)

const PercentageMultiplier = float64(1000)
const MaxPercentage = uint32(100 * PercentageMultiplier)

type FlagData struct { // nolint:revive
	// Variations are all the variations available for this flag. The minimum is 2 variations and, we don't have any max
	// limit except if the variationValue is a bool, the max is 2.
	Variations *map[string]*interface{} `json:"variations,omitempty" yaml:"variations,omitempty" toml:"variations,omitempty"` // nolint:lll

	// Rules is the list of Rule for this flag.
	// This an optional field.
	Rules *map[string]Rule `json:"targeting,omitempty" yaml:"targeting,omitempty" toml:"targeting,omitempty"`

	// DefaultRule is the rule applied after checking that any other rules
	// matched the user.
	DefaultRule *Rule `json:"defaultRule,omitempty" yaml:"defaultRule,omitempty" toml:"defaultRule,omitempty"`

	// Rollout is how we rollout the flag
	Rollout *Rollout `json:"rollout,omitempty" yaml:"rollout,omitempty" toml:"rollout,omitempty"`

	// TrackEvents is false if you don't want to export the data in your data exporter.
	// Default value is true
	TrackEvents *bool `json:"trackEvents,omitempty" yaml:"trackEvents,omitempty" toml:"trackEvents,omitempty"`

	// Disable is true if the flag is disabled.
	Disable *bool `json:"disable,omitempty" yaml:"disable,omitempty" toml:"disable,omitempty"`

	// Version (optional) This field contains the version of the flag.
	// The version is manually managed when you configure your flags and it is used to display the information
	// in the notifications and data collection.
	Version *string `json:"version,omitempty" yaml:"version,omitempty" toml:"version,omitempty"`
}

func (f *FlagData) GetVariationValue(variationName string) interface{} {
	for k, v := range f.GetVariations() {
		if k == variationName {
			if v == nil {
				return nil
			}
			return *v
		}
	}
	return nil
}

// Value is returning the Value associate to the flag
func (f *FlagData) Value(flagName string, user ffuser.User, sdkDefaultValue interface{}) (interface{}, string, error) {
	f.applyScheduledRolloutSteps()

	if f.isExperimentationOver() || f.IsDisable() {
		return sdkDefaultValue, constant.VariationSDKDefault, nil
	}

	variationName, err := f.getVariation(flagName, user)
	if err != nil {
		return sdkDefaultValue, constant.VariationSDKDefault, err
	}

	// If we have no value for a variation we consider that this variation does not exist.
	variations := *f.Variations
	if variations[variationName] == nil {
		return nil, variationName, fmt.Errorf("variation %s does not exist for the flag %s", variationName, flagName)
	}

	return f.GetVariationValue(variationName), variationName, nil
}

func (f *FlagData) getVariation(flagName string, user ffuser.User) (string, error) {
	hashID := utils.Hash(flagName+user.GetKey()) % MaxPercentage

	// Targeting rules
	if f.Rules != nil && len(*f.Rules) != 0 {
		rules := *f.Rules
		for _, rule := range rules {
			apply, varName, err := rule.Evaluate(user, hashID, false)
			if err != nil {
				return varName, err
			}
			if apply {
				return varName, nil
			}
		}
	}

	// Default rule
	if f.DefaultRule == nil {
		return "", fmt.Errorf("no default rule for the flag")
	}

	defaultRule := *f.DefaultRule
	_, varName, err := defaultRule.Evaluate(user, hashID, true)
	return varName, err
}

func (f *FlagData) isExperimentationOver() bool {
	now := time.Now()
	return f.Rollout != nil &&
		f.Rollout.Experimentation != nil &&
		((f.Rollout.Experimentation.Start != nil && now.Before(*f.Rollout.Experimentation.Start)) ||
			(f.Rollout.Experimentation.End != nil && now.After(*f.Rollout.Experimentation.End)))
}

// applyScheduledRolloutSteps is checking if the flag has a scheduled rollout configured.
// If yes we merge the changes to the current flag.
func (f *FlagData) applyScheduledRolloutSteps() {
	if f.Rollout == nil || f.Rollout.Scheduled == nil || len(f.Rollout.Scheduled.Steps) == 0 {
		// no update required because no scheduled rollout configuration
		return
	}

	now := time.Now()
	for _, step := range f.Rollout.Scheduled.Steps {
		// if the step has no date we ignore it
		if step.Date == nil {
			continue
		}

		// as soon as we have a step in the future we stop the updates
		if now.Before(*step.Date) {
			break
		}

		if step.Date != nil && now.After(*step.Date) {
			f.mergeScheduledStep(step)
		}
	}
}

// mergeScheduledStep will check every changes on the flag and apply them to the current configuration.
func (f *FlagData) mergeScheduledStep(stepFlag ScheduledStep) {
	if stepFlag.Disable != nil {
		f.Disable = stepFlag.Disable
	}

	if stepFlag.Rollout != nil {
		f.Rollout = stepFlag.Rollout
	}

	f.mergeVariationScheduledStep(stepFlag)
	f.mergeRulesScheduledStep(stepFlag)

	if stepFlag.DefaultRule != nil {
		f.DefaultRule.mergeChanges(*stepFlag.DefaultRule)
	}

	if stepFlag.TrackEvents != nil {
		f.TrackEvents = stepFlag.TrackEvents
	}

	if stepFlag.Version != nil {
		f.Version = stepFlag.Version
	}
}

// mergeRulesScheduledStep is used to merge the rules from a ScheduledStep.
// If we have a rule with an existing name we are overriding his content,
// if the rule is new we are adding it to the collection of rules.
func (f *FlagData) mergeRulesScheduledStep(stepFlag ScheduledStep) {
	if stepFlag.Rules != nil {
		rulesBeforeMerge := f.GetRules()
		for key, val := range *stepFlag.Rules {
			currentRule := f.GetRule(key)
			if (Rule{}) == currentRule {
				rulesBeforeMerge[key] = val
				continue
			}
			currentRule.mergeChanges(val)
			rulesBeforeMerge[key] = currentRule
		}
	}
}

// mergeVariationScheduledStep is used to merge the variations from a ScheduledStep.
// if the new value of an existing variation is nil we are deleting the variation.
// if the variation name already exist or is new we upsert the new value
func (f *FlagData) mergeVariationScheduledStep(stepFlag ScheduledStep) {
	if stepFlag.Variations != nil {
		variations := f.GetVariations()
		for variationName := range stepFlag.GetVariations() {
			// if the new value of an existing variation is nil we are deleting the variation.
			variationValue := stepFlag.GetVariationValue(variationName)
			if variationValue == nil {
				delete(variations, variationName)
				continue
			}

			// if the variation name already exist or is new we upsert the new value
			variations[variationName] = &variationValue
		}
		f.Variations = &variations
	}
}

func (f FlagData) String() string {
	var toString []string

	// Variations
	var variationString = make([]string, 0)
	for key, val := range f.GetVariations() {
		variationString = append(variationString, fmt.Sprintf("%s=%v", key, *val))
	}
	sort.Strings(variationString)
	toString = appendIfHasValue(toString, "Variations", strings.Join(variationString, ","))

	// Rules
	var rulesString = make([]string, 0)
	for _, rule := range f.GetRules() {
		rulesString = append(rulesString, fmt.Sprintf("[%v]", rule))
	}
	toString = appendIfHasValue(toString, "Rules", strings.Join(rulesString, ","))

	if f.GetDefaultRule() != nil {
		toString = appendIfHasValue(toString, "DefaultRule", f.GetDefaultRule().String())
	}

	// Others
	if f.GetRollout() != nil {
		toString = appendIfHasValue(toString, "Rollout", fmt.Sprintf("%v", *f.GetRollout()))
	}

	if f.TrackEvents != nil {
		toString = appendIfHasValue(toString, "TrackEvents", fmt.Sprintf("%t", f.IsTrackEvents()))
	}
	if f.Disable != nil {
		toString = appendIfHasValue(toString, "Disable", fmt.Sprintf("%t", f.IsDisable()))
	}
	toString = appendIfHasValue(toString, "Version", f.GetVersion())

	return strings.Join(toString, ", ")
}

func (f *FlagData) GetDefaultRule() *Rule {
	return f.DefaultRule
}

func (f *FlagData) GetRules() map[string]Rule {
	if f.Rules == nil {
		return map[string]Rule{}
	}
	return *f.Rules
}

func (f *FlagData) GetRule(name string) Rule {
	for k, v := range f.GetRules() {
		if k == name {
			return v
		}
	}
	return Rule{}
}

func (f *FlagData) GetVariations() map[string]*interface{} {
	if f.Variations == nil {
		return map[string]*interface{}{}
	}
	return *f.Variations
}

// IsTrackEvents is the getter of the field TrackEvents
func (f *FlagData) IsTrackEvents() bool {
	if f.TrackEvents == nil {
		return true
	}
	return *f.TrackEvents
}

// IsDisable is the getter for the field Disable
func (f *FlagData) IsDisable() bool {
	if f.Disable == nil {
		return false
	}
	return *f.Disable
}

// GetVersion is the getter for the field Version
func (f *FlagData) GetVersion() string {
	if f.Version == nil {
		return ""
	}
	return *f.Version
}

// GetRollout is the getter for the field Rollout
func (f *FlagData) GetRollout() *Rollout {
	return f.Rollout
}

// rulePercentageBucket is an internal representation of the limits of the
// bucket for a variation.
type rulePercentageBucket struct {
	start float64
	end   float64
}

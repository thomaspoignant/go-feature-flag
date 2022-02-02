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
func (f *FlagData) Value(flagName string, user ffuser.User, sdkDefaultValue interface{}) (interface{}, string) {
	f.updateFlagStage()

	variations := *f.Variations
	if f.isExperimentationOver() || f.IsDisable() {
		return sdkDefaultValue, constant.VariationSDKDefault
	}

	variationName, err := f.getVariation(flagName, user)
	if err != nil {
		// TODO log something + check that default variation exists
		return sdkDefaultValue, constant.VariationSDKDefault
	}

	if variations[variationName] == nil {
		return nil, variationName
	}

	return f.GetVariationValue(variationName), variationName
}

func (f *FlagData) getVariation(flagName string, user ffuser.User) (string, error) {
	hashID := utils.Hash(flagName+user.GetKey()) % MaxPercentage

	// Targeting rules
	if f.Rules != nil && len(*f.Rules) != 0 {
		rules := *f.Rules
		for _, rule := range rules {
			apply, varName, err := rule.Evaluate(user, hashID, false)
			if err != nil {
				// TODO log + continue to next rule
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

func (f *FlagData) updateFlagStage() {
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
			f.mergeChanges(step)
		}
	}
}

// mergeChanges will check every changes on the flag and apply them to the current configuration.
func (f *FlagData) mergeChanges(stepFlag ScheduledStep) {
	if stepFlag.Disable != nil {
		f.Disable = stepFlag.Disable
	}

	if stepFlag.Rollout != nil {
		f.Rollout = stepFlag.Rollout
	}

	// Replace all variations
	if stepFlag.Variations != nil {
		for variation := range stepFlag.GetVariations() {
			vVar := stepFlag.GetVariationValue(variation)
			if vVar != nil {
				err := f.replaceVariation(variation, vVar)
				if err != nil {
					// TODO: please write a log here
				}
			}
		}
	}

	// TODO: please write a comment to explain why we are doing this
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
	toString = appendIfHasValue(toString, "DefaultRule", f.GetDefaultRule().String())

	// Others
	if f.GetRollout() != nil {
		toString = appendIfHasValue(toString, "Rollout", fmt.Sprintf("%v", *f.GetRollout()))
	}
	toString = appendIfHasValue(toString, "TrackEvents", fmt.Sprintf("%t", f.IsTrackEvents()))
	toString = appendIfHasValue(toString, "Disable", fmt.Sprintf("%t", f.IsDisable()))
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

func (f *FlagData) replaceVariation(variationName string, variationValue interface{}) error {
	_, ok := f.GetVariations()[variationName]
	if ok {
		variations := f.GetVariations()
		variations[variationName] = &variationValue
		f.Variations = &variations
		return nil
	}
	return fmt.Errorf("impossible to update the variation %s, unknow variation name", variationName)
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

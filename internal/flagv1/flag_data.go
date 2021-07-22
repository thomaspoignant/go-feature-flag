package flagv1

import (
	"fmt"
	"github.com/nikunjy/rules/parser"
	"github.com/thomaspoignant/go-feature-flag/ffuser"
	"github.com/thomaspoignant/go-feature-flag/internal/utils"
	"math"
	"strconv"
	"strings"
	"time"
)

// percentageMultiplier is the multiplier used to have a bigger range of possibility.
const percentageMultiplier = float64(1000)

// FlagData describe the fields of a flag.
type FlagData struct {
	// Rule is the query use to select on which user the flag should apply.
	// Rule format is based on the nikunjy/rules module.
	// If no rule set, the flag apply to all users (percentage still apply).
	Rule *string `json:"rule,omitempty" yaml:"rule,omitempty" toml:"rule,omitempty"`

	// Percentage of the users affect by the flag.
	// Default value is 0
	Percentage *float64 `json:"percentage,omitempty" yaml:"percentage,omitempty" toml:"percentage,omitempty"`

	// True is the value return by the flag if apply to the user (rule is evaluated to true)
	// and user is in the active percentage.
	True *interface{} `json:"true,omitempty" yaml:"true,omitempty" toml:"true,omitempty"`

	// False is the value return by the flag if apply to the user (rule is evaluated to true)
	// and user is not in the active percentage.
	False *interface{} `json:"false,omitempty" yaml:"false,omitempty" toml:"false,omitempty"`

	// Default is the value return by the flag if not apply to the user (rule is evaluated to false).
	Default *interface{} `json:"default,omitempty" yaml:"default,omitempty" toml:"default,omitempty"`

	// TrackEvents is false if you don't want to export the data in your data exporter.
	// Default value is true
	TrackEvents *bool `json:"trackEvents,omitempty" yaml:"trackEvents,omitempty" toml:"trackEvents,omitempty"`

	// Disable is true if the flag is disabled.
	Disable *bool `json:"disable,omitempty" yaml:"disable,omitempty" toml:"disable,omitempty"`

	// Rollout is the object to configure how the flag is rollout.
	// You have different rollout strategy available but only one is used at a time.
	Rollout *Rollout `json:"rollout,omitempty" yaml:"rollout,omitempty" toml:"rollout,omitempty"`

	// Version (optional) This field contains the version of the flag.
	// The version is manually managed when you configure your flags and it is used to display the information
	// in the notifications and data collection.
	Version *float64 `json:"version,omitempty" yaml:"version,omitempty" toml:"version,omitempty"`
}

// Value is returning the Value associate to the flag (True / False / Default ) based
// if the toggle apply to the user or not.
func (f *FlagData) Value(flagName string, user ffuser.User) (interface{}, string) {
	f.updateFlagStage()
	if f.isExperimentationOver() {
		// if we have an experimentation that has not started or that is finished we use the default value.
		return f.getDefault(), VariationDefault
	}

	if f.evaluateRule(user) {
		if f.isInPercentage(flagName, user) {
			// Rule applied and user in the cohort.
			return f.getTrue(), VariationTrue
		}
		// Rule applied and user not in the cohort.
		return f.getFalse(), VariationFalse
	}

	// Default value is used if the rule does not applied to the user.
	return f.getDefault(), VariationDefault
}

func (f *FlagData) isExperimentationOver() bool {
	now := time.Now()
	return f.Rollout != nil && f.Rollout.Experimentation != nil && (
		(f.Rollout.Experimentation.Start != nil && now.Before(*f.Rollout.Experimentation.Start)) ||
			(f.Rollout.Experimentation.End != nil && now.After(*f.Rollout.Experimentation.End)))
}

// isInPercentage check if the user is in the cohort for the toggle.
func (f *FlagData) isInPercentage(flagName string, user ffuser.User) bool {
	percentage := int32(f.getActualPercentage())
	maxPercentage := uint32(100 * percentageMultiplier)

	// <= 0%
	if percentage <= 0 {
		return false
	}
	// >= 100%
	if uint32(percentage) >= maxPercentage {
		return true
	}

	hashID := utils.Hash(flagName+user.GetKey()) % maxPercentage
	return hashID < uint32(percentage)
}

// evaluateRule is checking if the rule can apply to a specific user.
func (f *FlagData) evaluateRule(user ffuser.User) bool {
	// Flag disable we cannot apply it.
	if f.GetDisable() {
		return false
	}

	// No rule means that all user can be impacted.
	if f.getRule() == "" {
		return true
	}

	// Evaluate the rule on the user.
	return parser.Evaluate(f.getRule(), utils.UserToMap(user))
}

// string display correctly a flag
func (f FlagData) String() string {
	toString := []string{}
	toString = append(toString, fmt.Sprintf("percentage=%d%%", int64(math.Round(f.getPercentage()))))
	if f.getRule() != "" {
		toString = append(toString, fmt.Sprintf("rule=\"%s\"", f.getRule()))
	}
	toString = append(toString, fmt.Sprintf("true=\"%v\"", f.getTrue()))
	toString = append(toString, fmt.Sprintf("false=\"%v\"", f.getFalse()))
	toString = append(toString, fmt.Sprintf("default=\"%v\"", f.getDefault()))
	toString = append(toString, fmt.Sprintf("disable=\"%v\"", f.GetDisable()))

	if f.TrackEvents != nil {
		toString = append(toString, fmt.Sprintf("trackEvents=\"%v\"", f.GetTrackEvents()))
	}

	if f.Version != nil {
		toString = append(toString, fmt.Sprintf("version=%s", strconv.FormatFloat(f.GetVersion(), 'f', -1, 64)))
	}

	return strings.Join(toString, ", ")
}

// getActualPercentage return the the actual percentage of the flag.
// the result value is the version with the percentageMultiplier.
func (f *FlagData) getActualPercentage() float64 {
	flagPercentage := f.getPercentage() * percentageMultiplier
	if f.Rollout == nil || f.Rollout.Progressive == nil {
		return flagPercentage
	}

	// compute progressive rollout percentage
	now := time.Now()

	// Missing date we ignore the progressive rollout
	if f.Rollout.Progressive.ReleaseRamp.Start == nil || f.Rollout.Progressive.ReleaseRamp.End == nil {
		return flagPercentage
	}
	// Expand percentage with the percentageMultiplier
	initialPercentage := f.Rollout.Progressive.Percentage.Initial * percentageMultiplier
	if f.Rollout.Progressive.Percentage.End == 0 {
		f.Rollout.Progressive.Percentage.End = 100
	}
	endPercentage := f.Rollout.Progressive.Percentage.End * percentageMultiplier

	if f.Rollout.Progressive.Percentage.Initial > f.Rollout.Progressive.Percentage.End {
		return flagPercentage
	}

	// Not in the range of the progressive rollout
	if now.Before(*f.Rollout.Progressive.ReleaseRamp.Start) {
		return initialPercentage
	}
	if now.After(*f.Rollout.Progressive.ReleaseRamp.End) {
		return endPercentage
	}

	// during the rollout ramp we compute the percentage
	nbSec := f.Rollout.Progressive.ReleaseRamp.End.Unix() - f.Rollout.Progressive.ReleaseRamp.Start.Unix()
	percentage := endPercentage - initialPercentage
	percentPerSec := percentage / float64(nbSec)

	c := now.Unix() - f.Rollout.Progressive.ReleaseRamp.Start.Unix()
	currentPercentage := float64(c)*percentPerSec + initialPercentage
	return currentPercentage
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
	if stepFlag.False != nil {
		f.False = stepFlag.False
	}
	if stepFlag.True != nil {
		f.True = stepFlag.True
	}
	if stepFlag.Default != nil {
		f.Default = stepFlag.Default
	}
	if stepFlag.TrackEvents != nil {
		f.TrackEvents = stepFlag.TrackEvents
	}
	if stepFlag.Percentage != nil {
		f.Percentage = stepFlag.Percentage
	}
	if stepFlag.Rule != nil {
		f.Rule = stepFlag.Rule
	}
	if stepFlag.Rollout != nil {
		f.Rollout = stepFlag.Rollout
	}
	if stepFlag.Version != nil {
		f.Version = stepFlag.Version
	}
}

// GetRule is the getter of the field Rule
func (f *FlagData) getRule() string {
	if f.Rule == nil {
		return ""
	}
	return *f.Rule
}

// GetPercentage is the getter of the field Percentage
func (f *FlagData) getPercentage() float64 {
	if f.Percentage == nil {
		return 0
	}
	return *f.Percentage
}

// GetTrue is the getter of the field True
func (f *FlagData) getTrue() interface{} {
	if f.True == nil {
		return nil
	}
	return *f.True
}

// f.getFalse is the getter of the field False
func (f *FlagData) getFalse() interface{} {
	if f.False == nil {
		return nil
	}
	return *f.False
}

// GetDefault is the getter of the field Default
func (f *FlagData) getDefault() interface{} {
	if f.Default == nil {
		return nil
	}
	return *f.Default
}

// GetTrackEvents is the getter of the field TrackEvents
func (f *FlagData) GetTrackEvents() bool {
	if f.TrackEvents == nil {
		return true
	}
	return *f.TrackEvents
}

// GetDisable is the getter for the field Disable
func (f *FlagData) GetDisable() bool {
	if f.Disable == nil {
		return false
	}
	return *f.Disable
}

// GetRollout is the getter for the field Rollout
func (f *FlagData) getRollout() *Rollout {
	return f.Rollout
}

// GetVersion is the getter for the field Version
func (f *FlagData) GetVersion() float64 {
	if f.Version == nil {
		return 0
	}
	return *f.Version
}

// GetVariationValue return the value of variation from his name
func (f *FlagData) GetVariationValue(variationName string) interface{} {
	switch variationName {
	case VariationDefault:
		return f.getDefault()
	case VariationTrue:
		return f.getTrue()
	case VariationFalse:
		return f.getFalse()
	default:
		return nil
	}
}

func (f *FlagData) GetDefaultVariation() string {
	return VariationDefault
}

func (f *FlagData) GetRawValues() map[string]string {
	var rawValues = make(map[string]string)
	rawValues["Rule"] = f.getRule()
	rawValues["Percentage"] = fmt.Sprintf("%.2f", f.getPercentage())

	if f.getRollout() == nil {
		rawValues["Rollout"] = ""
	} else {
		rawValues["Rollout"] = fmt.Sprintf("%v", f.getRollout())
	}
	rawValues["True"] = convertNilEmpty(f.getTrue())
	rawValues["False"] = convertNilEmpty(f.getFalse())
	rawValues["Default"] = convertNilEmpty(f.getDefault())
	rawValues["TrackEvents"] = fmt.Sprintf("%t", f.GetTrackEvents())
	rawValues["Disable"] = fmt.Sprintf("%t", f.GetDisable())
	rawValues["Version"] = fmt.Sprintf("%v", f.GetVersion())
	return rawValues
}

func convertNilEmpty(input interface{}) string {
	if input == nil {
		return ""
	}
	return fmt.Sprintf("%v", input)
}

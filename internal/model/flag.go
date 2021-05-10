package model

import (
	"fmt"
	"github.com/nikunjy/rules/parser"
	"hash/fnv"
	"math"
	"strings"
	"time"

	"github.com/thomaspoignant/go-feature-flag/ffuser"
)

// VariationType enum which describe the decision taken
type VariationType string

const (
	VariationTrue       VariationType = "True"
	VariationFalse      VariationType = "False"
	VariationDefault    VariationType = "Default"
	VariationSDKDefault VariationType = "SdkDefault"
)

// percentageMultiplier is the multiplier used to have a bigger range of possibility.
const percentageMultiplier = 1000

type Flag interface {
	// Value is returning the Value associate to the flag (True / False / Default ) based
	// if the flag apply to the user or not.
	Value(flagName string, user ffuser.User) (interface{}, VariationType)

	// String display correctly a flag with the right formatting
	String() string

	// GetRule is the getter of the field Rule
	// Default: empty string
	GetRule() string

	// GetPercentage is the getter of the field Rule
	// Default: 0.0
	GetPercentage() float64

	// GetTrue is the getter of the field True
	// Default: nil
	GetTrue() interface{}

	// GetFalse is the getter of the field False
	// Default: nil
	GetFalse() interface{}

	// GetDefault is the getter of the field Default
	// Default: nil
	GetDefault() interface{}

	// GetTrackEvents is the getter of the field TrackEvents
	// Default: true
	GetTrackEvents() bool

	// GetDisable is the getter of the field Disable
	// Default: false
	GetDisable() bool

	// GetRollout is the getter of the field Rollout
	// Default: nil
	GetRollout() *Rollout
}

// FlagData describe the fields of a flag.
type FlagData struct {
	// Rule is the query use to select on which user the flag should apply.
	// Rule format is based on the nikunjy/rules module.
	// If no rule set, the flag apply to all users (percentage still apply).
	Rule *string `json:"rule,omitempty" yaml:"rule,omitempty" toml:"rule,omitempty" slack_short:"false"`

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
	Rollout *Rollout `json:"rollout,omitempty" yaml:"rollout,omitempty" toml:"rollout,omitempty" slack_short:"false"` // nolint: lll
}

// Value is returning the Value associate to the flag (True / False / Default ) based
// if the toggle apply to the user or not.
func (f *FlagData) Value(flagName string, user ffuser.User) (interface{}, VariationType) {
	if f.isExperimentationOver() {
		// if we have an experimentation that has not started or that is finished we use the default value.
		return f.GetDefault(), VariationDefault
	}

	if f.evaluateRule(user) {
		if f.isInPercentage(flagName, user) {
			// Rule applied and user in the cohort.
			return f.GetTrue(), VariationTrue
		}
		// Rule applied and user not in the cohort.
		return f.GetFalse(), VariationFalse
	}

	// Default value is used if the rule does not applied to the user.
	return f.GetDefault(), VariationDefault
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

	hashID := Hash(flagName+user.GetKey()) % maxPercentage
	return hashID < uint32(percentage)
}

// evaluateRule is checking if the rule can apply to a specific user.
func (f *FlagData) evaluateRule(user ffuser.User) bool {
	// Flag disable we cannot apply it.
	if f.GetDisable() {
		return false
	}

	// No rule means that all user can be impacted.
	if f.GetRule() == "" {
		return true
	}

	// Evaluate the rule on the user.
	return parser.Evaluate(f.GetRule(), userToMap(user))
}

// string display correctly a flag
func (f FlagData) String() string {
	var strBuilder strings.Builder
	strBuilder.WriteString(fmt.Sprintf("percentage=%d%%, ", int64(math.Round(f.GetPercentage()))))
	if f.GetRule() != "" {
		strBuilder.WriteString(fmt.Sprintf("rule=\"%s\", ", f.GetRule()))
	}
	strBuilder.WriteString(fmt.Sprintf("true=\"%v\", ", f.GetTrue()))
	strBuilder.WriteString(fmt.Sprintf("false=\"%v\", ", f.GetFalse()))
	strBuilder.WriteString(fmt.Sprintf("true=\"%v\", ", f.GetDefault()))
	strBuilder.WriteString(fmt.Sprintf("disable=\"%v\"", f.GetDisable()))

	if f.TrackEvents != nil {
		strBuilder.WriteString(fmt.Sprintf(", trackEvents=\"%v\"", f.GetTrackEvents()))
	}

	return strBuilder.String()
}

// Hash is taking a string and convert.
func Hash(s string) uint32 {
	h := fnv.New32a()
	_, err := h.Write([]byte(s))
	// if we have a problem to get the hash we return 0
	if err != nil {
		return 0
	}
	return h.Sum32()
}

// userToMap convert the user to a MAP to use the query on it.
func userToMap(u ffuser.User) map[string]interface{} {
	// We don't have a json copy of the user.
	userCopy := make(map[string]interface{})

	// Duplicate the map to keep User un-mutable
	for key, value := range u.GetCustom() {
		userCopy[key] = value
	}
	userCopy["anonymous"] = u.IsAnonymous()
	userCopy["key"] = u.GetKey()
	return userCopy
}

// getActualPercentage return the the actual percentage of the flag.
// the result value is the version with the percentageMultiplier.
func (f *FlagData) getActualPercentage() float64 {
	flagPercentage := f.GetPercentage() * percentageMultiplier
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

// GetRule is the getter of the field Rule
func (f *FlagData) GetRule() string {
	if f.Rule == nil {
		return ""
	}
	return *f.Rule
}

// GetPercentage is the getter of the field Percentage
func (f *FlagData) GetPercentage() float64 {
	if f.Percentage == nil {
		return 0
	}
	return *f.Percentage
}

// GetTrue is the getter of the field True
func (f *FlagData) GetTrue() interface{} {
	if f.True == nil {
		return nil
	}
	return *f.True
}

// GetFalse is the getter of the field False
func (f *FlagData) GetFalse() interface{} {
	if f.False == nil {
		return nil
	}
	return *f.False
}

// GetDefault is the getter of the field Default
func (f *FlagData) GetDefault() interface{} {
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

// GetDisable is the getter of the field Disable
func (f *FlagData) GetDisable() bool {
	if f.Disable == nil {
		return false
	}
	return *f.Disable
}

// GetRollout is the getter of the field Rollout
func (f *FlagData) GetRollout() *Rollout {
	return f.Rollout
}

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

// Flag describe the fields of a flag.
type Flag struct {
	// Rule is the query use to select on which user the flag should apply.
	// Rule format is based on the nikunjy/rules module.
	// If no rule set, the flag apply to all users (percentage still apply).
	Rule string `json:"rule,omitempty" yaml:"rule,omitempty" toml:"rule,omitempty" slack_short:"false"`

	// Percentage of the users affect by the flag.
	// Default value is 0
	Percentage float64 `json:"percentage,omitempty" yaml:"percentage,omitempty" toml:"percentage,omitempty"`

	// True is the value return by the flag if apply to the user (rule is evaluated to true)
	// and user is in the active percentage.
	True interface{} `json:"true,omitempty" yaml:"true,omitempty" toml:"true,omitempty"`

	// False is the value return by the flag if apply to the user (rule is evaluated to true)
	// and user is not in the active percentage.
	False interface{} `json:"false,omitempty" yaml:"false,omitempty" toml:"false,omitempty"`

	// Default is the value return by the flag if not apply to the user (rule is evaluated to false).
	Default interface{} `json:"default,omitempty" yaml:"default,omitempty" toml:"default,omitempty"`

	// TrackEvents is false if you don't want to export the data in your data exporter.
	// Default value is true
	TrackEvents *bool `json:"trackEvents,omitempty" yaml:"trackEvents,omitempty" toml:"trackEvents,omitempty"`

	// Disable is true if the flag is disabled.
	Disable bool `json:"disable,omitempty" yaml:"disable,omitempty" toml:"disable,omitempty"`

	// Deprecated: you should use Rollout.Experimentation instead.
	Experimentation *Experimentation `json:"experimentation,omitempty" yaml:"experimentation,omitempty" toml:"experimentation,omitempty" slack_ignore:"true"` // nolint: lll

	// Rollout is the object to configure how the flag is rollout.
	// You have different rollout strategy available but only one is used at a time.
	Rollout *Rollout `json:"rollout,omitempty" yaml:"rollout,omitempty" toml:"rollout,omitempty" slack_short:"false"` // nolint: lll
}

// Value is returning the Value associate to the flag (True / False / Default ) based
// if the toggle apply to the user or not.
func (f *Flag) Value(flagName string, user ffuser.User) (interface{}, VariationType) {
	if f.isExperimentationOver() {
		// if we have an experimentation that has not started or that is finished we use the default value.
		return f.Default, VariationDefault
	}

	if f.evaluateRule(user) {
		if f.isInPercentage(flagName, user) {
			// Rule applied and user in the cohort.
			return f.True, VariationTrue
		}
		// Rule applied and user not in the cohort.
		return f.False, VariationFalse
	}

	// Default value is used if the rule does not applied to the user.
	return f.Default, VariationDefault
}

func (f *Flag) isExperimentationOver() bool {
	now := time.Now()
	// legacy experimentation notation
	// TO BE REMOVED when deprecated time is over
	legacy := f.Experimentation != nil && (
		(f.Experimentation.Start != nil && now.Before(*f.Experimentation.Start)) ||
			(f.Experimentation.StartDate != nil && now.Before(*f.Experimentation.StartDate)) ||
			(f.Experimentation.End != nil && now.After(*f.Experimentation.End)) ||
			(f.Experimentation.EndDate != nil && now.After(*f.Experimentation.EndDate)))

	if legacy {
		return legacy
	}
	// END TO BE REMOVED

	return f.Rollout != nil && f.Rollout.Experimentation != nil && (
		(f.Rollout.Experimentation.Start != nil && now.Before(*f.Rollout.Experimentation.Start)) ||
			(f.Rollout.Experimentation.End != nil && now.After(*f.Rollout.Experimentation.End)) ||
			// Remove bellow when deprecated field are removed
			(f.Rollout.Experimentation.StartDate != nil && now.Before(*f.Rollout.Experimentation.StartDate)) ||
			(f.Rollout.Experimentation.EndDate != nil && now.After(*f.Rollout.Experimentation.EndDate)))
}

// isInPercentage check if the user is in the cohort for the toggle.
func (f *Flag) isInPercentage(flagName string, user ffuser.User) bool {
	percentage := int32(f.getPercentage())
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
func (f *Flag) evaluateRule(user ffuser.User) bool {
	// Flag disable we cannot apply it.
	if f.Disable {
		return false
	}

	// No rule means that all user can be impacted.
	if f.Rule == "" {
		return true
	}

	// Evaluate the rule on the user.
	return parser.Evaluate(f.Rule, userToMap(user))
}

// string display correctly a flag
func (f Flag) String() string {
	var strBuilder strings.Builder
	strBuilder.WriteString(fmt.Sprintf("percentage=%d%%, ", int64(math.Round(f.Percentage))))
	if f.Rule != "" {
		strBuilder.WriteString(fmt.Sprintf("rule=\"%s\", ", f.Rule))
	}
	strBuilder.WriteString(fmt.Sprintf("true=\"%v\", ", f.True))
	strBuilder.WriteString(fmt.Sprintf("false=\"%v\", ", f.False))
	strBuilder.WriteString(fmt.Sprintf("true=\"%v\", ", f.Default))
	strBuilder.WriteString(fmt.Sprintf("disable=\"%v\"", f.Disable))

	if f.TrackEvents != nil {
		strBuilder.WriteString(fmt.Sprintf(", trackEvents=\"%v\"", *f.TrackEvents))
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

// getPercentage return the the actual percentage of the flag.
// the result value is the version with the percentageMultiplier.
func (f *Flag) getPercentage() float64 {
	flagPercentage := f.Percentage * percentageMultiplier
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

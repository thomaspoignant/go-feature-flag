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

type Experimentation struct {
	StartDate *time.Time `json:"startDate,omitempty" yaml:"startDate,omitempty" toml:"startDate,omitempty"`
	EndDate   *time.Time `json:"endDate,omitempty" yaml:"endDate,omitempty" toml:"endDate,omitempty"`
}

func (e Experimentation) String() string {
	buf := make([]string, 0)
	lo, _ := time.LoadLocation("UTC")
	if e.StartDate != nil {
		buf = append(buf, fmt.Sprintf("start:[%v]", e.StartDate.In(lo).Format(time.RFC3339)))
	}
	if e.EndDate != nil {
		buf = append(buf, fmt.Sprintf("end:[%v]", e.EndDate.In(lo).Format(time.RFC3339)))
	}
	return strings.Join(buf, " ")
}

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

	// Experimentation is your object to configure an experimentation, it will allow you to configure a start date and
	// an end date for your flag.
	// When the experimentation is not running, the flag will serve the default value.
	Experimentation *Experimentation `json:"experimentation,omitempty" yaml:"experimentation,omitempty" toml:"experimentation,omitempty" slack_short:"false"` // nolint: lll
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
	return f.Experimentation != nil && (
		(f.Experimentation.StartDate != nil && now.Before(*f.Experimentation.StartDate)) ||
			(f.Experimentation.EndDate != nil && now.After(*f.Experimentation.EndDate)))
}

// isInPercentage check if the user is in the cohort for the toggle.
func (f *Flag) isInPercentage(flagName string, user ffuser.User) bool {
	// 100%
	if f.Percentage == 100 {
		return true
	}

	// 0%
	if f.Percentage == 0 {
		return false
	}

	hashID := Hash(flagName+user.GetKey()) % 100
	return hashID < uint32(f.Percentage)
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

package flag

import (
	"encoding/json"
	"strconv"
)

// DtoFlag is struct able to manage all types of Flag that go-feature-flag supports.
type DtoFlag struct {
	// --- FLAGv1 FIELDS ---
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

	// --- COMMON FIELDS ---
	// Rollout is how we rollout the flag
	Rollout *DtoRollout `json:"rollout,omitempty" yaml:"rollout,omitempty" toml:"rollout,omitempty"`

	// TrackEvents is false if you don't want to export the data in your data exporter.
	// Default value is true
	TrackEvents *bool `json:"trackEvents,omitempty" yaml:"trackEvents,omitempty" toml:"trackEvents,omitempty"`

	// Disable is true if the flag is disabled.
	Disable *bool `json:"disable,omitempty" yaml:"disable,omitempty" toml:"disable,omitempty"`

	// Version (optional) This field contains the version of the flag.
	// The version is manually managed when you configure your flags and it is used to display the information
	// in the notifications and data collection.
	Version *string `json:"version,omitempty" yaml:"version,omitempty" toml:"version,omitempty"`

	// --- FLAGv2 FIELDS ---
	// Variations are all the variations available for this flag.
	// The minimum is 2 variations and we don't have any max limit except
	// if the variationValue is a bool, the max is 2.
	// Variations *map[string]*interface{}
	// `json:"variations,omitempty" yaml:"variations,omitempty" toml:"variations,omitempty"`
	//
	// Rules is the list of Rule for this flag.
	// This an optional field.
	// Rules *[]Rule `json:"targeting,omitempty" yaml:"targeting,omitempty" toml:"targeting,omitempty"`
	//
	// DefaultRule is the rule applied after checking that any other rules
	// matched the user.
	// DefaultRule *Rule `json:"defaultRule,omitempty" yaml:"defaultRule,omitempty" toml:"defaultRule,omitempty"`
}

// ConvertToFlagData detect the type of Flag and use the right convertor to create a FlagData instance.
func (d *DtoFlag) ConvertToFlagData() (FlagData, error) {
	// TODO: rewrite this function

	fd := FlagData{
		Rule:        d.Rule,
		Percentage:  d.Percentage,
		True:        d.True,
		False:       d.False,
		Default:     d.Default,
		TrackEvents: d.TrackEvents,
		Disable:     d.Disable,
	}

	if d.Version != nil {
		version, _ := strconv.ParseFloat(*d.Version, 64)
		fd.Version = &version
	}

	if d.Rollout != nil {
		r := Rollout{
			Experimentation: d.Rollout.Experimentation,
			Progressive:     d.Rollout.Progressive,
		}
		if d.Rollout.Scheduled != nil {
			scheduled := ScheduledRollout{}
			jsonString, _ := json.Marshal(d.Rollout.Scheduled)
			err := json.Unmarshal(jsonString, &scheduled)
			if err != nil {
				// TODO: log impossible to read scheduled
				return FlagData{}, err
			}
			*r.Scheduled = scheduled
		}
		fd.Rollout = &r
	}

	return fd, nil
}

package dto

import "github.com/thomaspoignant/go-feature-flag/internal/flagv1"

// DTO is representing all the fields we can have in a flag.
// This DTO supports all flag formats and convert them into an InternalFlag using a converter.
type DTO struct {
	DTOv0 `json:",inline" yaml:",inline" toml:",inline"`
	// Converter (optional) is the name of converter to use, if no converter specified we try to determine
	// which converter to use based on the fields we receive for the flag
	Converter *string `json:"converter,omitempty" yaml:"converter,omitempty" toml:"converter,omitempty"`
}

// DTOv0 describe the fields of a flag.
type DTOv0 struct {
	// Rule is the query use to select on which user the flag should apply.
	// Rule format is based on the nikunjy/rules module.
	// If no rule set, the flag apply to all users (percentage still apply).
	Rule *string `json:"rule,omitempty" yaml:"rule,omitempty" toml:"rule,omitempty"`

	// Percentage of the users affected by the flag.
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

	// Rollout is the object to configure how the flag is rolled out.
	// You have different rollout strategy available but only one is used at a time.
	Rollout *flagv1.Rollout `json:"rollout,omitempty" yaml:"rollout,omitempty" toml:"rollout,omitempty"`

	// Version (optional) This field contains the version of the flag.
	// The version is manually managed when you configure your flags and it is used to display the information
	// in the notifications and data collection.
	Version *float64 `json:"version,omitempty" yaml:"version,omitempty" toml:"version,omitempty"`
}

func (d *DTO) Convert() flagv1.FlagData {
	return flagv1.FlagData{
		Rule:        d.Rule,
		Percentage:  d.Percentage,
		True:        d.True,
		False:       d.False,
		Default:     d.Default,
		TrackEvents: d.TrackEvents,
		Disable:     d.Disable,
		Rollout:     d.Rollout,
		Version:     d.Version,
	}
}

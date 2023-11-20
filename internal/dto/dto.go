package dto

import (
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
)

// DTO is representing all the fields we can have in a flag.
// This DTO supports all flag formats and convert them into an InternalFlag using a converter.
type DTO struct {
	DTOv1 `json:",inline" yaml:",inline" toml:",inline"`
	DTOv0 `json:",inline" yaml:",inline" toml:",inline"`

	// TrackEvents is false if you don't want to export the data in your data exporter.
	// Default value is true
	TrackEvents *bool `json:"trackEvents,omitempty" yaml:"trackEvents,omitempty" toml:"trackEvents,omitempty"`

	// Disable is true if the flag is disabled.
	Disable *bool `json:"disable,omitempty" yaml:"disable,omitempty" toml:"disable,omitempty"`

	// Version (optional) This field contains the version of the flag.
	// The version is manually managed when you configure your flags and, it is used to display the information
	// in the notifications and data collection.
	Version *string `json:"version,omitempty" yaml:"version,omitempty" toml:"version,omitempty"`

	// Converter (optional) is the name of converter to use, if no converter specified we try to determine
	// which converter to use based on the fields we receive for the flag
	Converter *string `json:"converter,omitempty" yaml:"converter,omitempty" toml:"converter,omitempty"`
}

// DTOv1 is the new format of the flags since version 1.X.X
type DTOv1 struct {
	// Variations are all the variations available for this flag. The minimum is 2 variations and, we don't have any max
	// limit except if the variationValue is a bool, the max is 2.
	Variations *map[string]*interface{} `json:"variations,omitempty" yaml:"variations,omitempty" toml:"variations,omitempty"  jsonschema:"required,title=variations,description=All the variations available for this flag. You need at least 2 variations and it is a key value pair. All the variations should have the same type."` // nolint:lll

	// Rules is the list of Rule for this flag.
	// This an optional field.
	Rules *[]flag.Rule `json:"targeting,omitempty" yaml:"targeting,omitempty" toml:"targeting,omitempty" jsonschema:"title=targeting,description=List of rule to target a subset of the users based on the evaluation context."` // nolint: lll

	// DefaultRule is the rule applied after checking that any other rules
	// matched the user.
	DefaultRule *flag.Rule `json:"defaultRule,omitempty" yaml:"defaultRule,omitempty" toml:"defaultRule,omitempty" jsonschema:"required,title=defaultRule,description=How do we evaluate the flag if the user is not part of any of the targeting rule."` // nolint: lll

	// Scheduled is your struct to configure an update on some fields of your flag over time.
	// You can add several steps that updates the flag, this is typically used if you want to gradually add more user
	// in your flag.
	Scheduled *[]flag.ScheduledStep `json:"scheduledRollout,omitempty" yaml:"scheduledRollout,omitempty" toml:"scheduledRollout,omitempty" jsonschema:"title=scheduledRollout,description=Configure an update on some fields of your flag over time."` // nolint: lll

	// Experimentation is your struct to configure an experimentation.
	// It will allow you to configure a start date and an end date for your flag.
	// When the experimentation is not running, the flag will serve the default value.
	Experimentation *ExperimentationDto `json:"experimentation,omitempty" yaml:"experimentation,omitempty" toml:"experimentation,omitempty" jsonschema:"title=experimentation,description=Configure an experimentation. It will allow you to configure a start date and an end date for your flag."` // nolint: lll

	// Metadata is a field containing information about your flag such as an issue tracker link, a description, etc ...
	Metadata *map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty" toml:"metadata,omitempty" jsonschema:"title=metadata,description=A field containing information about your flag such as an issue tracker link a description etc..."` // nolint: lll
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

	// Rollout is the object to configure how the flag is rolled out.
	// You have different rollout strategy available but only one is used at a time.
	Rollout *Rollout `json:"rollout,omitempty" yaml:"rollout,omitempty" toml:"rollout,omitempty"`
}

func (d *DTO) Convert() flag.InternalFlag {
	if d == nil || (DTO{}) == *d {
		return flag.InternalFlag{}
	}
	if (d.Converter != nil && *d.Converter == "v0") || d.True != nil || d.False != nil {
		return ConvertV0DtoToInternalFlag(*d, false)
	}
	return ConvertV1DtoToInternalFlag(*d)
}

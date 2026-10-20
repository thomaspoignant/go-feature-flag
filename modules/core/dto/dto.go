package dto

import (
	"time"

	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
)

// DTO is representing all the fields we can have in a flag.
// This DTO supports all flag formats and convert them into an InternalFlag using a converter.
type DTO struct {
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

	// Variations are all the variations available for this flag. The minimum is 2 variations and, we don't have any max
	// limit except if the variationValue is a bool, the max is 2.
	Variations *map[string]*any `json:"variations,omitempty" yaml:"variations,omitempty" toml:"variations,omitempty" jsonschema:"required,title=variations,description=All the variations available for this flag. You need at least 2 variations and it is a key value pair. All the variations should have the same type."` // nolint:lll

	// Rules is the list of Rule for this flag.
	// This an optional field.
	Rules *[]flag.Rule `json:"targeting,omitempty" yaml:"targeting,omitempty" toml:"targeting,omitempty" jsonschema:"title=targeting,description=List of rule to target a subset of the users based on the evaluation context."` // nolint: lll

	// BucketingKey defines a source for a dynamic targeting key
	BucketingKey *string `json:"bucketingKey,omitempty" yaml:"bucketingKey,omitempty" toml:"bucketingKey,omitempty"`

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
	Metadata *map[string]any `json:"metadata,omitempty" yaml:"metadata,omitempty" toml:"metadata,omitempty" jsonschema:"title=metadata,description=A field containing information about your flag such as an issue tracker link a description etc..."` // nolint: lll

	// CreatedAt is the date used as the starting point of the TTL of the flag, in UTC.
	CreatedAt *time.Time `json:"createdAt,omitempty" yaml:"createdAt,omitempty" toml:"createdAt,omitempty" jsonschema:"title=createdAt,description=Creation date of the flag in UTC (for example 2026-01-15T00:00:00Z), starting point of its TTL. A flag without createdAt never expires."` // nolint: lll

	// TTL is the time to live of the flag, such as 90d or 12h.
	TTL *string `json:"ttl,omitempty" yaml:"ttl,omitempty" toml:"ttl,omitempty" jsonschema:"title=ttl,description=Time to live of the flag (for example 90d or 12h). Overrides the default TTL."` // nolint: lll

	// ExpiresAt is the date at which the flag expires, in UTC.
	ExpiresAt *time.Time `json:"expiresAt,omitempty" yaml:"expiresAt,omitempty" toml:"expiresAt,omitempty" jsonschema:"title=expiresAt,description=Date at which the flag expires in UTC (for example 2026-04-15T00:00:00Z). Overrides createdAt and ttl."` // nolint: lll

	// DependsOn binds the variations of the flag to the state of other flags.
	DependsOn *[]flag.FlagDependency `json:"dependsOn,omitempty" yaml:"dependsOn,omitempty" toml:"dependsOn,omitempty" jsonschema:"title=dependsOn,description=Binds the variations of the flag to the state of other flags."` // nolint: lll
}

// Convert is converting the DTO into a flag.InternalFlag.
func (d *DTO) Convert() flag.InternalFlag {
	if d == nil || (DTO{}) == *d {
		return flag.InternalFlag{}
	}
	return ConvertDtoToInternalFlag(*d)
}

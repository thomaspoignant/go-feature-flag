package flag

import (
	"github.com/thomaspoignant/go-feature-flag/ffuser"
)

type Flag interface {
	// Value is returning the Value associate to the flag (True / False / Default ) based
	// if the flag apply to the user or not.
	Value(flagName string, user ffuser.User) (interface{}, string)

	// String display correctly a flag with the right formatting
	String() string

	// GetVersion is the getter for the field Version
	// Default: 0.0
	GetVersion() float64

	// GetTrackEvents is the getter of the field TrackEvents
	// Default: true
	GetTrackEvents() bool

	// GetDisable is the getter for the field Disable
	// Default: false
	GetDisable() bool

	// GetDefaultVariation return the name of the default variation (if something goes wrong)
	GetDefaultVariation() string

	// GetVariationValue return the value of variation from his name
	GetVariationValue(variationName string) interface{}

	// GetRawValues is returning a raw value of the Flag used by the notifiers
	// We should not have any logic based on these values, this is only to
	// display  the information.
	GetRawValues() map[string]string
}

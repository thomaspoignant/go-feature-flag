package exporter

import (
	"time"

	"github.com/thomaspoignant/go-feature-flag/ffuser"
)

func NewFeatureEvent(
	user ffuser.User,
	flagKey string,
	value interface{},
	variation string,
	failed bool,
	version float64) FeatureEvent {
	contextKind := "user"
	if user.IsAnonymous() {
		contextKind = "anonymousUser"
	}

	return FeatureEvent{
		Kind:         "feature",
		ContextKind:  contextKind,
		UserKey:      user.GetKey(),
		CreationDate: time.Now().Unix(),
		Key:          flagKey,
		Variation:    variation,
		Value:        value,
		Default:      failed,
		Version:      version,
	}
}

type FeatureEvent struct {
	// Kind for a feature event is feature.
	// A feature event will only be generated if the trackEvents attribute of the flag is set to true.
	Kind string `json:"kind"`

	// ContextKind is the kind of context which generated an event. This will only be "anonymousUser" for events generated
	// on behalf of an anonymous user or the reserved word "user" for events generated on behalf of a non-anonymous user
	ContextKind string `json:"contextKind,omitempty"`

	// UserKey The key of the user object used in a feature flag evaluation. Details for the user object used in a feature
	// flag evaluation as reported by the "feature" event are transmitted periodically with a separate index event.
	UserKey string `json:"userKey"`

	// CreationDate When the feature flag was requested at Unix epoch time in milliseconds.
	CreationDate int64 `json:"creationDate"`

	// Key of the feature flag requested.
	Key string `json:"key"`

	// Variation  of the flag requested. Flag variation values can be "True", "False", "Default" or "SdkDefault"
	// depending on which value was taken during flag evaluation. "SdkDefault" is used when an error is detected and the
	// default value passed during the call to your variation is used.
	Variation string `json:"variation"`

	// Value of the feature flag returned by feature flag evaluation.
	Value interface{} `json:"value"`

	// Default value is set to true if feature flag evaluation failed, in which case the value returned was the default
	// value passed to variation. If the default field is omitted, it is assumed to be false.
	Default bool `json:"default"`

	// Version contains the version of the flag. If the field is omitted for the flag in the configuration file
	// the default version will be 0.
	Version float64 `json:"version"`
}

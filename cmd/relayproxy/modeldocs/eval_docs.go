package modeldocs

// EvalFlagDoc is the documentation struct for the Swagger doc.
type EvalFlagDoc struct {
	// `true` if the event was tracked by the relay proxy.
	TrackEvents bool `json:"trackEvents"   example:"true"`
	// The variation used to give you this value.
	VariationType string `json:"variationType" example:"variation-A"`
	// `true` if something went wrong in the relay proxy (flag does not exists, ...) and we serve the defaultValue.
	Failed bool `json:"failed"        example:"false"`
	// The version of the flag used.
	Version string `json:"version"       example:"1.0"`
	// reason why we have returned this value.
	Reason string `json:"reason"        example:"TARGETING_MATCH"`
	// Code of the error returned by the server.
	ErrorCode string `json:"errorCode"     example:""`
	// The flag value for this user.
	Value any `json:"value"`
	// Metadata is a field containing information about your flag such as an issue tracker link, a description, etc ...
	Metadata *map[string]any `json:"metadata"                                yaml:"metadata,omitempty" toml:"metadata,omitempty"`
}

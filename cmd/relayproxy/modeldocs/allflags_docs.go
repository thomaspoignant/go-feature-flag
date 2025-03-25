package modeldocs

// AllFlags model info
// @Description AllFlags contains the full list of all the flags available for the user
type AllFlags struct {
	// flags is the list of flag for the user.
	Flags map[string]FlagState `json:"flags"`

	// `true` if something went wrong in the relay proxy (flag does not exists, ...) and we serve the defaultValue.
	Valid bool `json:"valid" example:"false"`
}

// FlagState represents the state of an individual feature flag, with regard to a specific user, when it was called.
type FlagState struct {
	// Value is the flag value, it can be any JSON types.
	Value any `json:"value"`

	// Timestamp is the time when the flag was evaluated.
	Timestamp int64 `json:"timestamp" example:"1652113076"`

	// VariationType is the name of the variation used to have the flag value.
	VariationType string `json:"variationType" example:"variation-A"`

	// TrackEvents this flag is trackable.
	TrackEvents bool `json:"trackEvents" example:"false"`
	Failed      bool `json:"-"`
}

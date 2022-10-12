package modeldocs

// AllFlags model info
// @Description AllFlags contains the full list of all the flags available for the user
type AllFlags struct {
	// flags is the list of flag for the user.
	Flags map[string]FlagState `json:"flags"`

	// Valid if false it means there was an error (such as the data store not being available),
	// in which case no flag data is in this object.
	Valid bool `json:"valid" example:"false"`
}

// FlagState represents the state of an individual feature flag, with regard to a specific user, when it was called.
type FlagState struct {
	// Value is the flag value, it can be any JSON types.
	Value interface{} `json:"value"`

	// Timestamp is the time when the flag was evaluated.
	Timestamp int64 `json:"timestamp" example:"1652113076"`

	// VariationType is the name of the variation used to have the flag value.
	VariationType string `json:"variationType" example:"variation-A"`

	// TrackEvents this flag is trackable.
	TrackEvents bool `json:"trackEvents" example:"false"`
	Failed      bool `json:"-"`
}

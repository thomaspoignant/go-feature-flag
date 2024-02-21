package flagstate

import (
	"encoding/json"
)

// NewAllFlags will create a new AllFlags for a specific user.
// It sets valid to true, because until we have an error everything is valid.
func NewAllFlags() AllFlags {
	return AllFlags{
		valid: true,
		flags: map[string]FlagState{},
	}
}

// AllFlags is a snapshot of the state of multiple feature flags with regard to a specific user.
// This is the return type of ffclient.AllFlagsState().
// Serializing this object to JSON MarshalJSON() will produce a JSON you can sent to your front-end.
type AllFlags struct {
	// flags is the list of flag for the user.
	flags map[string]FlagState

	// Valid if false it means there was an error (such as the data store not being available),
	// in which case no flag data is in this object.
	valid bool
}

// AddFlag is adding a flag in the list for the specific user.
func (a *AllFlags) AddFlag(flagKey string, state FlagState) {
	a.flags[flagKey] = state
	if a.valid && state.Failed {
		a.valid = false
	}
}

// MarshalJSON is serializing the object to JSON, it will produce a JSON you can sent to your front-end.
func (a AllFlags) MarshalJSON() ([]byte, error) {
	res := struct {
		Flags map[string]FlagState `json:"flags,omitempty"`
		Valid bool                 `json:"valid"`
	}{
		Flags: a.flags,
		Valid: a.IsValid(),
	}
	return json.Marshal(res)
}

// IsValid is a getter to know if the AllFlags object is valid.
func (a *AllFlags) IsValid() bool {
	return a.valid
}

func (a *AllFlags) GetFlags() map[string]FlagState {
	return a.flags
}

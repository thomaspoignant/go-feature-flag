package model

type AllFlagRequest struct {
	// User The representation of a user for your feature flag system.
	User *UserRequest `json:"user" xml:"user" form:"user" query:"user"`
}

type EvalFlagRequest struct {
	AllFlagRequest `json:",inline" yaml:",inline" toml:",inline"`
	// The value will we use if we are not able to get the variation of the flag.
	DefaultValue interface{} `json:"defaultValue" xml:"defaultValue" form:"defaultValue" query:"defaultValue"`
}

// UserRequest The representation of a user for your feature flag system.
type UserRequest struct {
	// Key is the identifier of the UserRequest.
	Key string `json:"key" xml:"key" form:"key" query:"key" example:"08b5ffb7-7109-42f4-a6f2-b85560fbd20f"`

	// Anonymous set if this is a logged-in user or not.
	Anonymous bool `json:"anonymous" xml:"anonymous" form:"anonymous" query:"anonymous" example:"false"`

	// Custom is a map containing all extra information for this user.
	Custom map[string]interface{} `json:"custom" xml:"custom" form:"custom" query:"custom"  swaggertype:"object,string" example:"email:contact@gofeatureflag.org,firstname:John,lastname:Doe,company:GO Feature Flag"`
}

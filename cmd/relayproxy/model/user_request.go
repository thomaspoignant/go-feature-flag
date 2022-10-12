package model

type RelayProxyRequest struct {
	User         *UserRequest `json:"user" xml:"user" form:"user" query:"user"`
	DefaultValue interface{}  `json:"defaultValue" xml:"defaultValue" form:"defaultValue" query:"defaultValue"`
}

type UserRequest struct {
	// Key is the identifier of the UserRequest.
	Key string `json:"key" xml:"key" form:"key" query:"key"`

	// Anonymous set if this is a logged-in user or not.
	Anonymous bool `json:"anonymous" xml:"anonymous" form:"anonymous" query:"anonymous"`

	// Custom is a map containing all extra information for this user.
	Custom map[string]interface{} `json:"custom" xml:"custom" form:"custom" query:"custom"`
}

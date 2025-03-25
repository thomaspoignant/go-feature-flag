package model

type FlagDefinition struct {
	FlagType     FlagType `json:"flagType"`
	DefaultValue any      `json:"defaultValue"`
	Description  string   `json:"description"`
}

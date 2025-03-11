package model

type FlagDefinition struct {
	FlagType     FlagType    `json:"flagType"`
	DefaultValue interface{} `json:"defaultValue"`
	Description  string      `json:"description"`
}

package model

type FlagType string

const (
	FlagTypeBoolean FlagType = "boolean"
	FlagTypeString  FlagType = "string"
	FlagTypeInteger FlagType = "integer"
	FlagTypeFloat   FlagType = "float"
	FlagTypeObject  FlagType = "object"
)

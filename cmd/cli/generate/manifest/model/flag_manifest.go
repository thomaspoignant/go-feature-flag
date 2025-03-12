package model

type FlagManifest struct {
	Flags map[string]FlagDefinition `json:"flags"`
}

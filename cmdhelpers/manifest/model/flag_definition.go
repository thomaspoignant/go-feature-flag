package model

import "github.com/thomaspoignant/go-feature-flag/model/dto"

type FlagDefinition struct {
	dto.DTO      `json:",inline"`
	FlagType     FlagType `json:"flagType"`
	DefaultValue any      `json:"defaultValue"`
	Description  string   `json:"description"`
}

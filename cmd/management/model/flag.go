package model

import (
	"encoding/json"
	"fmt"
	"time"

	coredto "github.com/thomaspoignant/go-feature-flag/modules/core/dto"
)

type Flag struct {
	ID               string          `db:"id" json:"id"`
	FlagsetID        string          `db:"flagset_id" json:"flagsetId"`
	Name             string          `db:"name" json:"name"`
	CurrentVersionID *string         `db:"current_version_id" json:"currentVersionId,omitempty"`
	CurrentVersion   *int            `json:"currentVersion,omitempty"`
	Disabled         bool            `db:"disabled" json:"disabled"`
	Payload          json.RawMessage `json:"payload,omitempty" swaggertype:"object"`
	CreatedAt        time.Time       `db:"created_at" json:"createdAt"`
	UpdatedAt        time.Time       `db:"updated_at" json:"updatedAt"`
}

type FlagListItem struct {
	ID             string    `db:"id" json:"id"`
	FlagsetID      string    `db:"flagset_id" json:"flagsetId"`
	Name           string    `db:"name" json:"name"`
	CurrentVersion *int      `db:"current_version" json:"currentVersion,omitempty"`
	Disabled       bool      `db:"disabled" json:"disabled"`
	UpdatedAt      time.Time `db:"updated_at" json:"updatedAt"`
}

type CreateFlagRequest struct {
	Name    string      `json:"name"`
	Comment string      `json:"comment,omitempty"`
	Payload coredto.DTO `json:"payload"`
}

type UpdateFlagRequest struct {
	Comment string      `json:"comment,omitempty"`
	Payload coredto.DTO `json:"payload"`
}

type DisableFlagRequest struct {
	Disabled bool `json:"disabled"`
}

type FlagFilters struct {
	Name     *string
	Disabled *bool
	Page     int
	PageSize int
}

func ValidateFlagName(name string) ValidationErrors {
	var errs ValidationErrors
	if name == "" {
		errs = append(errs, ValidationError{Field: "name", Message: "flag name is required"})
	}
	if len(name) > 200 {
		errs = append(errs, ValidationError{Field: "name", Message: "flag name max 200 chars"})
	}
	return errs
}

func ValidateFlagPayload(p coredto.DTO) ValidationErrors {
	var errs ValidationErrors
	if p.Variations == nil || len(*p.Variations) < 2 {
		errs = append(errs, ValidationError{Field: "payload.variations", Message: "at least 2 variations are required"})
	}
	if p.DefaultRule == nil {
		errs = append(errs, ValidationError{Field: "payload.defaultRule", Message: "defaultRule is required"})
	} else {
		if p.DefaultRule.Percentages != nil {
			var sum float64
			for _, v := range *p.DefaultRule.Percentages {
				sum += v
			}
			if sum > 0 && (sum < 99.999 || sum > 100.001) {
				errs = append(errs, ValidationError{Field: "payload.defaultRule.percentage", Message: fmt.Sprintf("percentages must sum to 100, got %.2f", sum)})
			}
		}
	}
	if p.Rules != nil {
		for i, r := range *p.Rules {
			if r.Query == nil || *r.Query == "" {
				errs = append(errs, ValidationError{Field: fmt.Sprintf("payload.targeting[%d].query", i), Message: "query is required for targeting rules"})
			}
		}
	}
	return errs
}

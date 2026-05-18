package model

import (
	"encoding/json"
	"time"
)

type FlagVersion struct {
	ID            string          `db:"id" json:"id"`
	FlagID        string          `db:"flag_id" json:"flagId"`
	VersionNumber int             `db:"version_number" json:"versionNumber"`
	Payload       json.RawMessage `db:"payload" json:"payload" swaggertype:"object"`
	Comment       string          `db:"comment" json:"comment"`
	CreatedBy     *string         `db:"created_by" json:"createdBy,omitempty"`
	CreatedByName *string         `db:"created_by_name" json:"createdByName,omitempty"`
	CreatedAt     time.Time       `db:"created_at" json:"createdAt"`
}

type RollbackRequest struct {
	Comment string `json:"comment,omitempty"`
}

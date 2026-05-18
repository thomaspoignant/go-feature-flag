package model

import (
	"encoding/json"
	"time"
)

type Flagset struct {
	ID                string          `db:"id" json:"id"`
	TeamID            string          `db:"team_id" json:"teamId"`
	Name              string          `db:"name" json:"name"`
	Description       string          `db:"description" json:"description"`
	APIKeyHashes      []string        `db:"api_key_hashes" json:"-"`
	APIKeyCount       int             `json:"apiKeyCount"`
	PollingIntervalMs *int            `db:"polling_interval_ms" json:"pollingIntervalMs,omitempty"`
	FileFormat        *string         `db:"file_format" json:"fileFormat,omitempty"`
	Retrievers        json.RawMessage `db:"retrievers" json:"retrievers,omitempty" swaggertype:"object"`
	Exporters         json.RawMessage `db:"exporters" json:"exporters,omitempty" swaggertype:"object"`
	Notifiers         json.RawMessage `db:"notifiers" json:"notifiers,omitempty" swaggertype:"object"`
	CreatedAt         time.Time       `db:"created_at" json:"createdAt"`
	UpdatedAt         time.Time       `db:"updated_at" json:"updatedAt"`
}

type CreateFlagsetRequest struct {
	Name              string          `json:"name"`
	Description       string          `json:"description"`
	PollingIntervalMs *int            `json:"pollingIntervalMs,omitempty"`
	FileFormat        *string         `json:"fileFormat,omitempty"`
	Retrievers        json.RawMessage `json:"retrievers,omitempty" swaggertype:"object"`
	Exporters         json.RawMessage `json:"exporters,omitempty" swaggertype:"object"`
	Notifiers         json.RawMessage `json:"notifiers,omitempty" swaggertype:"object"`
}

type UpdateFlagsetRequest struct {
	Name              *string         `json:"name,omitempty"`
	Description       *string         `json:"description,omitempty"`
	PollingIntervalMs *int            `json:"pollingIntervalMs,omitempty"`
	FileFormat        *string         `json:"fileFormat,omitempty"`
	Retrievers        json.RawMessage `json:"retrievers,omitempty" swaggertype:"object"`
	Exporters         json.RawMessage `json:"exporters,omitempty" swaggertype:"object"`
	Notifiers         json.RawMessage `json:"notifiers,omitempty" swaggertype:"object"`
}

type CreateAPIKeyResponse struct {
	APIKey  string `json:"apiKey"`
	KeyHash string `json:"keyHash"`
}

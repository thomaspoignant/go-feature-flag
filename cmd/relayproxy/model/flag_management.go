package model

import (
	manifestmodel "github.com/thomaspoignant/go-feature-flag/cmdhelpers/manifest/model"
	"github.com/thomaspoignant/go-feature-flag/modules/core/dto"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
)

// FlagType is the OpenFeature value type of a flag, inferred from its variation values.
type FlagType = manifestmodel.FlagType

const (
	FlagTypeBoolean = manifestmodel.FlagTypeBoolean
	FlagTypeString  = manifestmodel.FlagTypeString
	FlagTypeInteger = manifestmodel.FlagTypeInteger
	FlagTypeFloat   = manifestmodel.FlagTypeFloat
	FlagTypeObject  = manifestmodel.FlagTypeObject
)

// FlagDefinition is the native GO Feature Flag flag definition (variations/targeting/
// defaultRule/metadata/...), reused verbatim from dto.DTO to guarantee wire-format parity
// with the rest of the relay proxy and avoid a parallel type.
type FlagDefinition = dto.DTO

// CreateFlagCommand is the body of POST /v1/flags.
type CreateFlagCommand struct {
	Key        string  `json:"key"`
	Definition dto.DTO `json:"definition"`
}

// FlagResponse is a flag plus server-derived read-only fields, returned by every
// flag-management endpoint that returns a single flag.
type FlagResponse struct {
	Key        string   `json:"key"`
	Type       FlagType `json:"type"`
	Definition dto.DTO  `json:"definition"`
	// Source is the retriever owning this flag, e.g. "postgresql:go_feature_flag".
	Source string `json:"source,omitempty"`
	// Editable is false when the owning retriever/flagset doesn't support writes.
	Editable bool `json:"editable"`
	// ETag is the current revision; echo it back as If-Match on the next write.
	ETag string `json:"etag,omitempty"`
}

// FlagListPagination is the pagination metadata of FlagListResponse.
type FlagListPagination struct {
	Total  int `json:"total"`
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

// FlagListResponse is the body returned by GET /v1/flags.
type FlagListResponse struct {
	Flags []FlagResponse     `json:"flags"`
	Meta  FlagListPagination `json:"pagination"`
}

// ErrorResponse is the error envelope shared by every flag-management endpoint, matching
// the relay proxy's existing OFREP/configuration error shape.
type ErrorResponse struct {
	ErrorCode    flag.ErrorCode `json:"errorCode"`
	ErrorDetails string         `json:"errorDetails,omitempty"`
}

// SetFlagStateCommand is the body of PATCH /v1/flags/{flag_key}/state.
type SetFlagStateCommand struct {
	Disable bool `json:"disable"`
}

package retriever

import (
	"context"
	"errors"
)

// ErrFlagNotFound is returned by a WritableRetriever when the requested flag key does not exist.
var ErrFlagNotFound = errors.New("flag not found")

// ErrFlagAlreadyExists is returned by CreateFlag when the flag key already exists.
var ErrFlagAlreadyExists = errors.New("flag already exists")

// ErrETagMismatch is returned when the caller-supplied If-Match ETag does not match the
// current revision of the flag (optimistic-concurrency conflict).
var ErrETagMismatch = errors.New("etag mismatch")

// ErrFlagsetNotConfigured is returned by every write method when the retriever has no
// explicit, non-empty flagset configured: there is no unambiguous target to write to.
var ErrFlagsetNotConfigured = errors.New("no explicit flagset configured for write operations")

// WritableRetriever is an optional capability a Retriever can implement to support the
// GO Feature Flag relay-proxy flag-management API (create/replace/delete a single flag).
// It is intentionally not part of the base Retriever interface: most retrievers (file, S3,
// GitHub, ...) have no compare-and-swap primitive and simply do not implement it, which lets
// callers detect write support with a type assertion instead of a feature flag.
type WritableRetriever interface {
	// GetFlag returns the JSON definition bytes of a single flag and its current ETag.
	// Returns ErrFlagNotFound if the flag does not exist.
	GetFlag(threadContext context.Context, flagKey string) (definition []byte, etag string, getFlagErr error)

	// CreateFlag creates a new flag. Returns ErrFlagAlreadyExists if the key is already used.
	CreateFlag(threadContext context.Context, flagKey string, definition []byte) (etag string, createFlagErr error)

	// UpsertFlag replaces the flag definition, creating it if it does not exist yet.
	// created is true when the flag was created (no prior row) and false when it was updated.
	// When ifMatch is non-nil and the flag already exists, its current ETag must equal
	// *ifMatch or ErrETagMismatch is returned; ifMatch is ignored when the flag doesn't exist yet.
	UpsertFlag(
		threadContext context.Context, flagKey string, definition []byte, ifMatch *string,
	) (created bool, etag string, upsertFlagErr error)

	// DeleteFlag removes a flag. Returns ErrFlagNotFound if it doesn't exist. When ifMatch is
	// non-nil, its value must equal the flag's current ETag or ErrETagMismatch is returned.
	DeleteFlag(threadContext context.Context, flagKey string, ifMatch *string) error

	// Source identifies the retriever instance owning the flag, e.g. "postgresql:<table>".
	// Surfaced in the management API response so callers know where a flag is stored.
	Source() string
}

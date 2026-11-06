package configfile

import (
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/thomaspoignant/go-feature-flag/modules/core/dto"
)

// errInternalValidation is the generic error used when the validation of a flag panics,
// it avoids leaking Go runtime details to the callers.
var errInternalValidation = errors.New(
	"internal validation error: flag definition could not be validated",
)

// FlagError is the validation error of a single flag.
type FlagError struct {
	Flag string
	Err  error
}

// Error implements the error interface.
func (e FlagError) Error() string {
	return fmt.Sprintf("invalid flag %s: %v", e.Flag, e.Err)
}

// Unwrap returns the underlying validation error.
func (e FlagError) Unwrap() error { return e.Err }

// ValidateFlags converts and validates every flag.
// It returns the validation errors sorted by flag key.
func ValidateFlags(flags map[string]dto.DTO) []FlagError {
	var errs []FlagError
	for _, key := range slices.Sorted(maps.Keys(flags)) {
		if err := validateFlag(flags[key]); err != nil {
			errs = append(errs, FlagError{Flag: key, Err: err})
		}
	}
	return errs
}

// validateFlag converts and validates a single flag.
// The recover guards against panics in the IsValid method of modules/core (for example on partially
// filled rollout definitions): such a flag is reported as invalid instead of crashing the caller.
func validateFlag(flagDto dto.DTO) (err error) {
	defer func() {
		if recover() != nil {
			err = errInternalValidation
		}
	}()
	flag := flagDto.Convert()
	return flag.IsValid()
}

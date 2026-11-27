// Package validation provides opt-in validation of flag variation values.
package validation

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"slices"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/thomaspoignant/go-feature-flag/modules/core/dto"
)

// VariationError identifies a variation value that does not match its JSON Schema.
type VariationError struct {
	FlagKey          string
	Variation        string
	InstanceLocation string // JSON Pointer within the variation value (empty for the root).
	ScheduledStep    *int   // Zero-based index in scheduledRollout, or nil for a base variation.
	Message          string
}

// ValidateVariations checks the base and scheduled variation values of a parsed
// flag definition against a caller-provided JSON Schema. The schema is compiled
// once per call. References within the schema are resolved without loading
// external files or URLs. Variation values are checked in their JSON form.
//
// Schema violations are returned in the slice. An invalid schema or a
// validation failure other than a schema violation is returned as an error.
// Callers can check the flag definition separately using dto.DTO.Convert and
// the resulting flag's IsValid method.
func ValidateVariations(flagKey string, definition dto.DTO, schemaJSON []byte) ([]VariationError, error) {
	const schemaURL = "https://go-feature-flag.invalid/variation-schema.json"
	compiler := jsonschema.NewCompiler()
	compiler.LoadURL = func(url string) (io.ReadCloser, error) {
		return nil, fmt.Errorf("external JSON Schema resource %q is not supported", url)
	}
	if err := compiler.AddResource(schemaURL, bytes.NewReader(schemaJSON)); err != nil {
		return nil, fmt.Errorf("parse JSON Schema for flag %q: %w", flagKey, err)
	}
	schema, err := compiler.Compile(schemaURL)
	if err != nil {
		return nil, fmt.Errorf("compile JSON Schema for flag %q: %w", flagKey, err)
	}

	parsed := definition.Convert()
	violations, err := validateValues(nil, schema, flagKey, parsed.GetVariations(), nil)
	if err != nil {
		return nil, err
	}
	if parsed.Scheduled != nil {
		for step, scheduled := range *parsed.Scheduled {
			violations, err = validateValues(violations, schema, flagKey, scheduled.GetVariations(), &step)
			if err != nil {
				return nil, fmt.Errorf("scheduledRollout[%d]: %w", step, err)
			}
		}
	}
	return violations, nil
}

func validateValues(
	violations []VariationError, schema *jsonschema.Schema, flagKey string, values map[string]*any, step *int,
) ([]VariationError, error) {
	for _, name := range slices.Sorted(maps.Keys(values)) {
		err := validateJSONValue(schema, values[name])
		if err == nil {
			continue
		}
		var violation *jsonschema.ValidationError
		if !errors.As(err, &violation) {
			return nil, fmt.Errorf("validate flag %q variation %q: %w", flagKey, name, err)
		}
		location := VariationError{FlagKey: flagKey, Variation: name, ScheduledStep: step}
		violations = appendViolations(violations, location, violation)
	}
	return violations, nil
}

func validateJSONValue(schema *jsonschema.Schema, value *any) error {
	// YAML and TOML can produce Go values such as time.Time, which encode as JSON strings.
	encoded, err := json.Marshal(value)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.UseNumber()
	var instance any
	if err := decoder.Decode(&instance); err != nil {
		return err
	}
	return schema.Validate(instance)
}

func appendViolations(
	violations []VariationError, location VariationError, violation *jsonschema.ValidationError,
) []VariationError {
	if len(violation.Causes) == 0 {
		location.InstanceLocation = violation.InstanceLocation
		location.Message = violation.Message
		return append(violations, location)
	}
	for _, cause := range violation.Causes {
		violations = appendViolations(violations, location, cause)
	}
	return violations
}

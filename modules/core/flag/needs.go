package flag

import (
	"reflect"

	"github.com/thomaspoignant/go-feature-flag/modules/core/utils"
)

// NeedsDependency represents a single dependency condition declared by a flag through its
// `needs` field. A flag is only evaluated normally when all the dependencies it declares are
// satisfied; if any of them is unmet the flag is treated as disabled (it returns the SDK
// default value without evaluating its targeting rules or default rule).
//
// The dependency resolution is limited to one level: the dependency flag is resolved but its
// own `needs` field is ignored. This keeps the evaluation predictable and makes dependency
// cycles safe (A needs B, B needs A).
type NeedsDependency struct {
	// Flag is the name of the flag this flag depends on.
	Flag *string `json:"flag,omitempty" yaml:"flag,omitempty" toml:"flag,omitempty" jsonschema:"required,title=flag,description=Name of the flag this flag depends on."` // nolint: lll

	// Value is the expected resolved value of the dependency flag.
	// This field is optional: when it is omitted the dependency is considered satisfied if the
	// dependency flag resolves to true (convenient for boolean flags).
	Value *any `json:"value,omitempty" yaml:"value,omitempty" toml:"value,omitempty" jsonschema:"title=value,description=Expected resolved value of the dependency flag. Optional; when omitted it defaults to true."` // nolint: lll
}

// GetFlag returns the name of the dependency flag.
func (n *NeedsDependency) GetFlag() string {
	if n.Flag == nil {
		return ""
	}
	return *n.Flag
}

// GetExpectedValue returns the value the dependency flag is expected to resolve to.
// When no value is configured, it defaults to true.
func (n *NeedsDependency) GetExpectedValue() any {
	if n.Value == nil {
		return true
	}
	return *n.Value
}

// needsValueEqual compares the resolved value of a dependency flag with the expected value.
// Numeric values are coerced to float64 before comparison so that the way JSON/YAML/TOML decode
// numbers (int vs float64) does not create false mismatches (e.g. an expected 1 matches a
// resolved 1.0).
func needsValueEqual(resolved, expected any) bool {
	if resolvedNumber, ok := utils.ToFloat(resolved); ok {
		if expectedNumber, ok := utils.ToFloat(expected); ok {
			return resolvedNumber == expectedNumber
		}
		return false
	}
	return reflect.DeepEqual(resolved, expected)
}

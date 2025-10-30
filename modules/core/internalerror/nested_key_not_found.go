package internalerror

import "fmt"

// NestedKeyNotFoundError is the error returned when a nested key is not found.
type NestedKeyNotFoundError struct {
	Key string
}

// Implement the Error() method for the custom error type
func (e *NestedKeyNotFoundError) Error() string {
	return fmt.Sprintf("Error: nested key not found: %s", e.Key)
}

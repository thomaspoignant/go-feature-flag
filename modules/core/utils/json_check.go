package utils

import (
	"encoding/json"
	"strings"
)

// IsJSONObject checks if a string is a valid JSON
func IsJSONObject(s string) bool {
	// Decoding into a map only succeeds for an object or for null, so skip the copy
	// and the decode when s can be neither. JSON allows only these four whitespace
	// characters before a value.
	trimmed := strings.TrimLeft(s, " \t\n\r")
	if !strings.HasPrefix(trimmed, "{") && !strings.HasPrefix(trimmed, "null") {
		return false
	}
	var js map[string]any
	return json.Unmarshal([]byte(s), &js) == nil
}

package utils

import (
	"encoding/json"
	"strings"
)

// IsJSONObject checks if a string is a valid JSON
func IsJSONObject(s string) bool {
	// Anything that does not start (after JSON whitespace) with "{" or "null" can never unmarshal
	// into a map, so we skip the decoder for the common non-JSON (nikunjy) queries.
	t := strings.TrimLeft(s, " \t\r\n")
	// "null" unmarshals into a nil map without error, only when followed by nothing but whitespace.
	if rest, ok := strings.CutPrefix(t, "null"); ok {
		return strings.TrimLeft(rest, " \t\r\n") == ""
	}
	if !strings.HasPrefix(t, "{") {
		return false
	}
	var js map[string]any
	return json.Unmarshal([]byte(s), &js) == nil
}

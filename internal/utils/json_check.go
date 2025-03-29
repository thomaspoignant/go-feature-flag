package utils

import "encoding/json"

// IsJSONObject checks if a string is a valid JSON
func IsJSONObject(s string) bool {
	var js map[string]any
	return json.Unmarshal([]byte(s), &js) == nil
}

package utils

import "strings"

// StringToArray is a helper function to convert a slice of strings
func StringToArray(item []string) []string {
	if len(item) > 0 {
		return strings.Split(item[0], ",")
	}
	return item
}

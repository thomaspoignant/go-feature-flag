package utils

import "strings"

func StrTrim(s string) string {
	if !strings.Contains(s, "\n") {
		return strings.TrimLeft(s, " ")
	}
	trimmed := strings.Split(s, "\n")
	for index, item := range trimmed {
		trimmed[index] = strings.TrimLeft(item, " ")
	}
	return strings.Join(trimmed, "")
}

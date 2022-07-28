package utils

import (
	"fmt"
)

func AppendIfHasValue(toString []string, key string, value string) []string {
	if value != "" {
		toString = append(toString, fmt.Sprintf("%s:[%v]", key, value))
	}
	return toString
}

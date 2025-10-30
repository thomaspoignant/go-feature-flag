package utils

import (
	"fmt"
	"strings"
)

// GetNestedFieldValue returns the value from a nested path in the given map.
// If the path does not exist or an error occurs, it returns an error.
func GetNestedFieldValue(ctx map[string]interface{}, bucketingKey string) (interface{}, error) {
	if ctx == nil || bucketingKey == "" {
		return nil, fmt.Errorf("nested key not found: %s", bucketingKey)
	}

	parts := strings.Split(bucketingKey, ".")
	var current interface{} = ctx

	for _, part := range parts {
		m, ok := current.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("nested key not found: %s", bucketingKey)
		}

		current, ok = m[part]
		if !ok {
			return nil, fmt.Errorf("nested key not found: %s", bucketingKey)
		}
	}
	return current, nil
}

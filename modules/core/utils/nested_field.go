package utils

import (
	"strings"

	"github.com/thomaspoignant/go-feature-flag/modules/core/internalerror"
)

// GetNestedFieldValue returns the value from a nested path in the given map.
// If the path does not exist or an error occurs, it returns an error.
func GetNestedFieldValue(ctx map[string]any, bucketingKey string) (any, error) {
	if ctx == nil || bucketingKey == "" {
		return nil, &internalerror.NestedKeyNotFoundError{Key: bucketingKey}
	}

	parts := strings.Split(bucketingKey, ".")
	var current any = ctx

	for _, part := range parts {
		m, ok := current.(map[string]any)
		if !ok {
			return nil, &internalerror.NestedKeyNotFoundError{Key: bucketingKey}
		}

		current, ok = m[part]
		if !ok {
			return nil, &internalerror.NestedKeyNotFoundError{Key: bucketingKey}
		}
	}
	return current, nil
}

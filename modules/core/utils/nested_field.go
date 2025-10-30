package utils

import (
	"encoding/json"
	"fmt"

	"github.com/tidwall/gjson"
)

// GetNestedFieldValue returns the value from a nested path in the given map using
// tidwall/gjson. The path is a dot-separated key (e.g., "a.b.c").
// If the path does not exist or an error occurs, it returns an error.
func GetNestedFieldValue(ctx map[string]interface{}, bucketingKey string) (interface{}, error) {
	if ctx == nil || bucketingKey == "" {
		return nil, fmt.Errorf("nested key not found: %s", bucketingKey)
	}

	data, err := json.Marshal(ctx)
	if err != nil {
		return nil, err
	}

	res := gjson.GetBytes(data, bucketingKey)
	if !res.Exists() {
		return nil, fmt.Errorf("nested key not found: %s", bucketingKey)
	}
	return res.Value(), nil
}

package utils

func JSONTypeExtractor(variation any) (string, error) {
	switch variation.(type) {
	case string:
		return "(string)", nil
	case float64, int:
		return "(number)", nil
	case bool:
		return "(bool)", nil
	case []any:
		return "([]interface{})", nil
	case map[string]any:
		return "(map[string]interface{})", nil
	}
	return "", nil
}

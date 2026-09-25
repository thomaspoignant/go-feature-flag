package utils

import "encoding/json"

// IsIntegral returns true if the float is an integer.
func IsIntegral(val float64) bool {
	return val == float64(int64(val))
}

// ToFloat converts the numeric types produced when decoding a flag configuration (JSON decodes
// numbers as float64, YAML as int, TOML as int64, or json.Number when a decoder uses UseNumber)
// into a float64. It returns false for non-numeric values.
func ToFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case float32:
		return float64(n), true
	case float64:
		return n, true
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}

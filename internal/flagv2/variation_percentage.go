package flagv2

import "fmt"

// VariationPercentage represents the percentage to affect for a specific
// variation in your rule.
type VariationPercentage map[string]float64

func (v *VariationPercentage) GetVariationName() (string, error) {
	for key := range *v {
		return key, nil
	}
	return "", fmt.Errorf("invalid percentage format")
}

func (v *VariationPercentage) GetPercentage() (float64, error) {
	for _, value := range *v {
		return value, nil
	}
	return 0, fmt.Errorf("invalid percentage format")
}

func (v VariationPercentage) String() string {
	for key, value := range v {
		return fmt.Sprintf("%s:%6.4f", key, value)
	}
	return ""
}

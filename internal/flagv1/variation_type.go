package flagv1

// VariationType enum which describe the decision taken
type VariationType = string

const (
	// VariationTrue is a constant to explain that we are using the "True" variation
	VariationTrue VariationType = "True"

	// VariationFalse is a constant to explain that we are using the "False" variation
	VariationFalse VariationType = "False"

	// VariationDefault is a constant to explain that we are using the "Default" variation
	VariationDefault VariationType = "Default"

	// VariationSDKDefault is a constant to explain that we are using the default from the SDK variation
	VariationSDKDefault VariationType = "SdkDefault"
)

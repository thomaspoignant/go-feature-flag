package model

type VariationResult struct {
	TrackEvents   bool
	VariationType VariationType
	Failed        bool
	Version       float64
}

// BoolVarResult is the internal result format of a bool variation.
// This is used by the ffclient.BoolVariation functions and by the ffclient.AllFlagsState function
type BoolVarResult struct {
	VariationResult
	Value bool
}

// IntVarResult is the internal result format of a int variation.
// This is used by the ffclient.IntVariation functions and by the ffclient.AllFlagsState function
type IntVarResult struct {
	VariationResult
	Value int
}

// Float64VarResult is the internal result format of a float64 variation.
// This is used by the ffclient.Float64Variation functions and by the ffclient.AllFlagsState function
type Float64VarResult struct {
	VariationResult
	Value float64
}

// StringVarResult is the internal result format of a string variation.
// This is used by the ffclient.StringVariation functions and by the ffclient.AllFlagsState function
type StringVarResult struct {
	VariationResult
	Value string
}

// JSONVarResult is the internal result format of a json variation.
// This is used by the ffclient.JSONVariation functions and by the ffclient.AllFlagsState function
type JSONVarResult struct {
	VariationResult
	Value map[string]interface{}
}

// JSONArrayVarResult is the internal result format of a json array variation.
// This is used by the ffclient.JSONArrayVariation functions and by the ffclient.AllFlagsState function
type JSONArrayVarResult struct {
	VariationResult
	Value []interface{}
}

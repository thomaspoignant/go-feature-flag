package flag

type EvaluationContext struct {
	// Environment is the name of your current env
	// this value will be added to the custom information of your user and,
	// it will allow to create rules based on this environment,
	Environment string

	// DefaultSdkValue is the default value of the SDK when calling the variation.
	DefaultSdkValue interface{}
}

package flag

type Context struct {
	// EvaluationContextEnrichment will be merged with the evaluation context sent during the evaluation.
	// It is useful to add common attributes to all the evaluation, such as a server version, environment, ...
	//
	// All those fields will be included in the custom attributes of the evaluation context,
	// if in the evaluation context you have a field with the same name, it will override the common one.
	// Default: nil
	EvaluationContextEnrichment map[string]any `json:"evaluationContextEnrichment,omitempty"`

	// DefaultSdkValue is the default value of the SDK when calling the variation.
	DefaultSdkValue any `json:"defaultSdkValue,omitempty"`

	// DependencyFlagResolver resolves a sibling flag by name so that a flag declaring a `needs`
	// dependency can be evaluated against it. The consumer running the evaluation is responsible
	// for setting it (typically backed by the flag cache), which is why it is not serialized.
	//
	// When it is nil, a flag that declares a `needs` dependency is treated as disabled
	// (fail-closed). Flags without a `needs` field are never impacted.
	DependencyFlagResolver func(flagName string) (Flag, bool) `json:"-"`
}

// AddIntoEvaluationContextEnrichment adds a key and value to the evaluation context enrichment.
func (s *Context) AddIntoEvaluationContextEnrichment(key string, value any) {
	if s.EvaluationContextEnrichment == nil {
		s.EvaluationContextEnrichment = make(map[string]any)
	}
	s.EvaluationContextEnrichment[key] = value
}

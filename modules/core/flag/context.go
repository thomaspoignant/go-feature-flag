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
}

// AddIntoEvaluationContextEnrichment adds a key and value to the evaluation context enrichment.
func (s *Context) AddIntoEvaluationContextEnrichment(key string, value any) {
	if s.EvaluationContextEnrichment == nil {
		s.EvaluationContextEnrichment = make(map[string]any)
	}
	s.EvaluationContextEnrichment[key] = value
}

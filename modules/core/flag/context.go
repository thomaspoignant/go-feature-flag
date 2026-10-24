package flag

import "time"

// Context is copied on purpose by the dependency helpers (value receivers), the pointer receiver is only a setter.
//
//nolint:recvcheck
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

	// DefaultTTL is the TTL applied to the flags without their own TTL.
	// A zero or negative value means that the flags without their own TTL never expire.
	DefaultTTL time.Duration `json:"defaultTTL,omitempty"`

	// GetFlag gives access to the other flags, it is required to evaluate the dependencies between flags.
	// Without it a flag with a dependency never has it met.
	GetFlag func(flagName string) (Flag, bool) `json:"-"`

	// dependencyChain is the list of flags being resolved, used to stop on circular dependencies.
	dependencyChain []string

	// memo keeps what has been computed about the parent flags, so a parent shared by several
	// paths is resolved once.
	memo *dependencyMemo
}

// WithMemo returns a copy of the context that shares what is computed about the parent flags.
// It is valid for one evaluation, or for one pass over a fixed set of flags: do not keep it longer.
func (c Context) WithMemo() Context {
	if c.memo == nil {
		c.memo = &dependencyMemo{}
	}
	return c
}

// AddIntoEvaluationContextEnrichment adds a key and value to the evaluation context enrichment.
func (s *Context) AddIntoEvaluationContextEnrichment(key string, value any) {
	if s.EvaluationContextEnrichment == nil {
		s.EvaluationContextEnrichment = make(map[string]any)
	}
	s.EvaluationContextEnrichment[key] = value
}

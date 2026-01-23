package ffcontext

import "encoding/json"

var _ Context = (*EvaluationContext)(nil)

type Context interface {
	// GetKey return the unique targetingKey for the context.
	GetKey() string
	// IsAnonymous return if the context is about an anonymous user or not.
	IsAnonymous() bool
	// GetCustom return all the attributes properties added to the context.
	GetCustom() map[string]any
	// AddCustomAttribute allows to add a attributes attribute into the context.
	AddCustomAttribute(name string, value any)
	// ExtractGOFFProtectedFields extract the goff specific attributes from the evaluation context.
	ExtractGOFFProtectedFields() GoffContextSpecifics
}

// value is a type to define attributes.
type value map[string]any

// NewEvaluationContext creates a new evaluation context identified by the given targetingKey.
func NewEvaluationContext(key string) EvaluationContext {
	return EvaluationContext{targetingKey: key, attributes: map[string]any{}}
}

// Deprecated: NewAnonymousEvaluationContext is here for compatibility reason.
// Please use NewEvaluationContext instead and add a attributes attribute to know that it is an anonymous user.
//
// ctx := NewEvaluationContext("my-targetingKey")
// ctx.AddCustomAttribute("anonymous", true)
func NewAnonymousEvaluationContext(key string) EvaluationContext {
	return EvaluationContext{targetingKey: key, attributes: map[string]any{
		"anonymous": true,
	}}
}

// EvaluationContext contains specific attributes for your evaluation.
// Most of the time it is identifying a user browsing your site.
// The only mandatory property is the Key, which must a unique identifier.
// For authenticated users, this may be a username or e-mail address.
// For anonymous users, this could be an IP address or session ID.
//
// EvaluationContext fields are immutable and can be accessed only via getter methods.
// To construct an EvaluationContext, use either a simple constructor (NewEvaluationContext) or the builder pattern
// with NewEvaluationContextBuilder.
type EvaluationContext struct {
	// uniquely identifying the subject (end-user, or client service) of a flag evaluation
	targetingKey string
	attributes   value
}

// MarshalJSON is a custom JSON marshaller for EvaluationContext.
// It will only marshal the targetingKey and the attributes of the context and avoid to expose the internal structure.
func (u EvaluationContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		TargetingKey string `json:"targetingKey"`
		Attributes   value  `json:"attributes"`
	}{
		TargetingKey: u.targetingKey,
		Attributes:   u.attributes,
	})
}

// GetKey return the unique targetingKey for the user.
func (u EvaluationContext) GetKey() string {
	return u.targetingKey
}

// IsAnonymous return if the user is anonymous or not.
func (u EvaluationContext) IsAnonymous() bool {
	anonymous := u.attributes["anonymous"]
	switch v := anonymous.(type) {
	case bool:
		return v
	default:
		return false
	}
}

// GetCustom return all the attributes properties of a user.
func (u EvaluationContext) GetCustom() map[string]any {
	return u.attributes
}

// AddCustomAttribute allows to add a attributes attribute into the user.
func (u EvaluationContext) AddCustomAttribute(name string, value any) {
	if name != "" {
		u.attributes[name] = value
	}
}

func (u EvaluationContext) ToMap() map[string]any {
	resMap := u.attributes
	resMap["targetingKey"] = u.targetingKey
	return resMap
}

// ExtractGOFFProtectedFields extract the goff specific attributes from the evaluation context.
func (u EvaluationContext) ExtractGOFFProtectedFields() GoffContextSpecifics {
	goff := GoffContextSpecifics{}
	switch v := u.attributes["gofeatureflag"].(type) {
	case map[string]string:
		goff.addCurrentDateTime(v["currentDateTime"])
		goff.addListFlags(v["flagList"])
		goff.addExporterMetadata(v["exporterMetadata"])
	case map[string]any:
		goff.addCurrentDateTime(v["currentDateTime"])
		goff.addListFlags(v["flagList"])
		goff.addExporterMetadata(v["exporterMetadata"])
	case GoffContextSpecifics:
		return v
	}
	return goff
}

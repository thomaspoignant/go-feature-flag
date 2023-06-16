package ffcontext

type Context interface {
	// GetKey return the unique key for the context.
	GetKey() string
	// IsAnonymous return if the context is about an anonymous user or not.
	IsAnonymous() bool
	// GetCustom return all the custom properties added to the context.
	GetCustom() map[string]interface{}
	// AddCustomAttribute allows to add a custom attribute into the context.
	AddCustomAttribute(name string, value interface{})
}

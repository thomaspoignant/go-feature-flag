package ffuser

// value is a type to define custom attribute.
type value map[string]any

// Deprecated: NewUser, please use ffcontext.NewEvaluationContext instead
//
// NewUser creates a new user identified by the given key.
func NewUser(key string) User {
	return User{key: key, custom: map[string]any{}}
}

// Deprecated: NewAnonymousUser, please use ffcontext.NewAnonymousEvaluationContext instead.
//
// NewAnonymousUser creates a new anonymous user identified by the given key.
func NewAnonymousUser(key string) User {
	return User{key: key, anonymous: true, custom: map[string]any{}}
}

// Deprecated: User, please use ffcontext.EvaluationContext instead.
//
// A User contains specific attributes of a user browsing your site. The only mandatory property is the Key,
// which must uniquely identify each user. For authenticated users, this may be a username or e-mail address.
// For anonymous users, this could be an IP address or session ID.
//
// User fields are immutable and can be accessed only via getter methods. To construct a User, use either
// a simple constructor (NewUser, NewAnonymousUser) or the builder pattern with NewUserBuilder.
type User struct {
	key       string // only mandatory attribute
	anonymous bool
	custom    value
}

// GetKey return the unique key for the user.
func (u User) GetKey() string {
	return u.key
}

// IsAnonymous return if the user is anonymous or not.
func (u User) IsAnonymous() bool {
	return u.anonymous
}

// GetCustom return all the custom properties of a user.
func (u User) GetCustom() map[string]any {
	return u.custom
}

// AddCustomAttribute allows to add a custom attribute into the user.
func (u User) AddCustomAttribute(name string, value any) {
	if name != "" {
		u.custom[name] = value
	}
}

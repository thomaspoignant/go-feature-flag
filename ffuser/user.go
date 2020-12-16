package ffuser

type Value map[string]interface{}

// NewUser creates a new user identified by the given key.
func NewUser(key string) User {
	return User{key: key, custom: map[string]interface{}{}}
}

// NewAnonymousUser creates a new anonymous user identified by the given key.
func NewAnonymousUser(key string) User {
	return User{key: key, anonymous: true, custom: map[string]interface{}{}}
}

// A User contains specific attributes of a user browsing your site. The only mandatory property is the Key,
// which must uniquely identify each user. For authenticated users, this may be a username or e-mail address.
// For anonymous users, this could be an IP address or session ID.
//
// User fields are immutable and can be accessed only via getter methods. To construct a User, use either
// a simple constructor (NewUser, NewAnonymousUser) or the builder pattern with NewUserBuilder.
type User struct {
	key       string // only mandatory attribute
	anonymous bool
	custom    Value
}

// GetKey return the unique key for the user.
func (u *User) GetKey() string {
	return u.key
}

// IsAnonymous return if the user is anonymous or not.
func (u *User) IsAnonymous() bool {
	return u.anonymous
}

// GetCustom return all the custom properties of a user.
func (u *User) GetCustom() map[string]interface{} {
	return u.custom
}

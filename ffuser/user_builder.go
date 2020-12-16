package ffuser

// NewUserBuilder constructs a new UserBuilder, specifying the user key.
//
// For authenticated users, the key may be a username or e-mail address. For anonymous users,
// this could be an IP address or session ID.
func NewUserBuilder(key string) UserBuilder {
	return &userBuilderImpl{
		key:    key,
		custom: map[string]interface{}{},
	}
}

type UserBuilder interface {
	Anonymous(bool) UserBuilder
	AddCustom(string, interface{}) UserBuilder
	Build() User
}

type userBuilderImpl struct {
	key       string // only mandatory attribute
	anonymous bool
	custom    Value
}

func (u *userBuilderImpl) Anonymous(anonymous bool) UserBuilder {
	u.anonymous = anonymous
	return u
}

func (u *userBuilderImpl) AddCustom(key string, value interface{}) UserBuilder {
	u.custom[key] = value
	return u
}

func (u *userBuilderImpl) Build() User {
	return User{
		key:       u.key,
		anonymous: u.anonymous,
		custom:    u.custom,
	}
}

package ffuser

// Deprecated: NewUserBuilder, please use ffcontext.NewEvaluationContextBuilder instead.
//
// NewUserBuilder constructs a new UserBuilder, specifying the user key.
//
// For authenticated users, the key may be a username or e-mail address. For anonymous users,
// this could be an IP address or session ID.
func NewUserBuilder(key string) UserBuilder {
	return &userBuilderImpl{
		key:    key,
		custom: map[string]any{},
	}
}

// Deprecated: UserBuilder, please use ffcontext.EvaluationContextBuilder instead.
//
// UserBuilder is a builder to create a User.
type UserBuilder interface {
	Anonymous(bool) UserBuilder
	AddCustom(string, any) UserBuilder
	Build() User
}

type userBuilderImpl struct {
	key       string // only mandatory attribute
	anonymous bool
	custom    value
}

// Anonymous is to set the user as anonymous.
func (u *userBuilderImpl) Anonymous(anonymous bool) UserBuilder {
	u.anonymous = anonymous
	return u
}

// AddCustom allows you to add a custom attribute to the user.
func (u *userBuilderImpl) AddCustom(key string, value any) UserBuilder {
	u.custom[key] = value
	return u
}

// Build is creating the user.
func (u *userBuilderImpl) Build() User {
	return User{
		key:       u.key,
		anonymous: u.anonymous,
		custom:    u.custom,
	}
}

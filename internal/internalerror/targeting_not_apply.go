package internalerror

import (
	"fmt"

	"github.com/thomaspoignant/go-feature-flag/ffuser"
)

type RuleNotApply struct {
	User ffuser.User
}

func (m *RuleNotApply) Error() string {
	return fmt.Sprintf("Rule does not apply for this user %s", m.User.GetKey())
}

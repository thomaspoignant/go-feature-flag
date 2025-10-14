package internalerror

import (
	"fmt"

	"github.com/thomaspoignant/go-feature-flag/modules/core/ffcontext"
)

type RuleNotApplyError struct {
	Context ffcontext.Context
}

func (m *RuleNotApplyError) Error() string {
	return fmt.Sprintf("Rule does not apply for this user %s", m.Context.GetKey())
}

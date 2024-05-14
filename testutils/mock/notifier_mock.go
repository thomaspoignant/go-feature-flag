package mock

import (
	"github.com/thomaspoignant/go-feature-flag/notifier"
)

type Notifier struct {
	NumberCalls int
}

func (n *Notifier) Notify(cache notifier.DiffCache) error {
	n.NumberCalls++
	return nil
}

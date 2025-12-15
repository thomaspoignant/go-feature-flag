package mock

import (
	"sync"

	"github.com/thomaspoignant/go-feature-flag/notifier"
)

type Notifier struct {
	NotifyCalls int
	mu          sync.Mutex
}

func (n *Notifier) Notify(_ notifier.DiffCache) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.NotifyCalls++
	return nil
}

func (n *Notifier) GetNotifyCalls() int {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.NotifyCalls
}

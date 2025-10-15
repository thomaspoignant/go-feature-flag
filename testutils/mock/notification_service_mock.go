package mock

import (
	"sync"

	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

type NotificationService struct {
	NotifyCalls int
	CloseCalled bool
	mu          sync.Mutex
}

func (n *NotificationService) Notify(
	oldCache map[string]flag.Flag,
	newCache map[string]flag.Flag,
	log *fflog.FFLogger,
) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.NotifyCalls++
}

func (n *NotificationService) Close() {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.CloseCalled = true
}

func (n *NotificationService) GetNotifyCalls() int {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.NotifyCalls
}

func (n *NotificationService) WasCloseCalled() bool {
	return n.CloseCalled
}

package service

import (
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"sync"
)

type notifierRelayProxy struct {
	websocketService WebsocketService
}

func NewNotifierRelayProxy(websocketService WebsocketService) notifier.Notifier {
	return &notifierRelayProxy{
		websocketService: websocketService,
	}
}

func (n *notifierRelayProxy) Notify(diff notifier.DiffCache, waitGroup *sync.WaitGroup) error {
	defer waitGroup.Done()
	n.websocketService.BroadcastFlagChanges(diff)
	return nil
}

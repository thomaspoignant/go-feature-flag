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

func (n *notifierRelayProxy) Notify(_ notifier.DiffCache, waitGroup *sync.WaitGroup) error {
	defer waitGroup.Done()
	n.websocketService.BroadcastText("test broadcast")
	return nil
}

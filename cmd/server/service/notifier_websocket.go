package service

import (
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"sync"
)

type websocketNotifier struct {
	websocketService WebsocketService
}

func NewWebsocketNotifier(websocketService WebsocketService) notifier.Notifier {
	return &websocketNotifier{
		websocketService: websocketService,
	}
}

func (n *websocketNotifier) Notify(diff notifier.DiffCache, waitGroup *sync.WaitGroup) error {
	defer waitGroup.Done()
	n.websocketService.BroadcastFlagChanges(diff)
	return nil
}

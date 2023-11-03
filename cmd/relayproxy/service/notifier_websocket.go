package service

import (
	"github.com/thomaspoignant/go-feature-flag/notifier"
)

type notifierWebsocket struct {
	websocketService WebsocketService
}

func NewNotifierWebsocket(websocketService WebsocketService) notifier.Notifier {
	return &notifierWebsocket{
		websocketService: websocketService,
	}
}

func (n *notifierWebsocket) Notify(diff notifier.DiffCache) error {
	n.websocketService.BroadcastFlagChanges(diff)
	return nil
}

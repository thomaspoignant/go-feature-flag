package proxynotifier

import (
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"github.com/thomaspoignant/go-feature-flag/notifier"
)

type notifierWebsocket struct {
	websocketService service.WebsocketService
}

func NewNotifierWebsocket(websocketService service.WebsocketService) notifier.Notifier {
	return &notifierWebsocket{
		websocketService: websocketService,
	}
}

func (n *notifierWebsocket) Notify(diff notifier.DiffCache) error {
	n.websocketService.BroadcastFlagChanges(diff)
	return nil
}

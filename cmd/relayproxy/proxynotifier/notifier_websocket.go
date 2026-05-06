package proxynotifier

import (
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service/stream"
	"github.com/thomaspoignant/go-feature-flag/notifier"
)

type notifierWebsocket struct {
	websocketService stream.WebsocketService
}

func NewNotifierWebsocket(websocketService stream.WebsocketService) notifier.Notifier {
	return &notifierWebsocket{
		websocketService: websocketService,
	}
}

func (n *notifierWebsocket) Notify(diff notifier.DiffCache) error {
	n.websocketService.BroadcastFlagChanges(diff)
	return nil
}

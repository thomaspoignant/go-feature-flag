package proxynotifier

import (
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service/stream"
	"github.com/thomaspoignant/go-feature-flag/notifier"
)

type notifierWebsocket struct {
	websocketService stream.WebsocketService
	flagsetName      string
}

func NewNotifierWebsocket(websocketService stream.WebsocketService, flagsetName string) notifier.Notifier {
	return &notifierWebsocket{
		websocketService: websocketService,
		flagsetName:      flagsetName,
	}
}

func (n *notifierWebsocket) Notify(diff notifier.DiffCache) error {
	n.websocketService.BroadcastFlagChanges(n.flagsetName, diff)
	return nil
}

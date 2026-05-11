package proxynotifier

import (
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service/stream"
	"github.com/thomaspoignant/go-feature-flag/notifier"
)

type notifierSSE struct {
	sseService  stream.SSEService
	flagsetName string
}

// NewNotifierSSE creates a notifier that forwards flag change events to the
// SSE service scoped to the given flagset name.
func NewNotifierSSE(sseService stream.SSEService, flagsetName string) notifier.Notifier {
	return &notifierSSE{
		sseService:  sseService,
		flagsetName: flagsetName,
	}
}

func (n *notifierSSE) Notify(diff notifier.DiffCache) error {
	return n.sseService.BroadcastFlagChanges(n.flagsetName, diff)
}

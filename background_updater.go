package ffclient

import (
	"time"
)

// backgroundUpdater contains what is needed to manage the
// background update of the flags.
type backgroundUpdater struct {
	ticker      *time.Ticker
	updaterChan chan struct{}
}

// newBackgroundUpdater init default value for the ticker and the channel.
func newBackgroundUpdater(pollInterval int) backgroundUpdater {
	return backgroundUpdater{
		ticker:      time.NewTicker(time.Duration(pollInterval) * time.Second),
		updaterChan: make(chan struct{}),
	}
}

// close stop the ticker and close the channel.
func (bgu *backgroundUpdater) close() {
	bgu.ticker.Stop()
	close(bgu.updaterChan)
}

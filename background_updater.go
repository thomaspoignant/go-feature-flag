package ffclient

import (
	"math/rand"
	"time"
)

// backgroundUpdater contains what is needed to manage the
// background update of the flags.
type backgroundUpdater struct {
	ticker      *time.Ticker
	updaterChan chan struct{}
}

// newBackgroundUpdater init default value for the ticker and the channel.
func newBackgroundUpdater(pollingInterval time.Duration, useJitter bool) backgroundUpdater {
	tickerDuration := pollingInterval
	if useJitter {
		// we accept a deviation of maximum 10% of the polling interval
		maxJitter := float64(pollingInterval) * 0.1
		jitter := time.Duration(0)
		if int64(maxJitter) > 0 {
			jitter = time.Duration(rand.Int63n(int64(maxJitter))) // nolint: gosec
		}
		if jitter%2 == 0 {
			tickerDuration = pollingInterval + jitter
		} else {
			tickerDuration = pollingInterval - jitter
		}
	}

	return backgroundUpdater{
		ticker:      time.NewTicker(tickerDuration),
		updaterChan: make(chan struct{}),
	}
}

// close stops the ticker and closes the channel.
func (bgu *backgroundUpdater) close() {
	if bgu.ticker != nil {
		bgu.ticker.Stop()
	}
	if !bgu.isClosed() {
		close(bgu.updaterChan)
	}
}

// isClosed checks if the channel is closed.
func (bgu *backgroundUpdater) isClosed() bool {
	select {
	case <-bgu.updaterChan:
		return true
	default:
	}
	return false
}

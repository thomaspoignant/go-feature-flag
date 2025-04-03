package logsnotifier

import (
	"log/slog"

	"github.com/thomaspoignant/go-feature-flag/notifier"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

// Notifier is the component in charge of sending flag changes to the logs.
type Notifier struct {
	Logger *fflog.FFLogger
}

// Notify is the notifying all the changes to the notifier.
func (c *Notifier) Notify(diff notifier.DiffCache) error {
	for key := range diff.Deleted {
		c.Logger.Info("flag removed", slog.String("key", key))
	}

	for key := range diff.Added {
		c.Logger.Info("flag added", slog.String("key", key))
	}

	for key, flagDiff := range diff.Updated {
		if flagDiff.After.IsDisable() != flagDiff.Before.IsDisable() {
			if flagDiff.After.IsDisable() {
				// Flag is disabled
				c.Logger.Info("flag is turned OFF", slog.String("key", key))
				continue
			}
			c.Logger.Info("flag is turned ON", slog.String("key", key))
			continue
		}
		// key has changed in cache
		c.Logger.Info("flag updated", slog.String("key", key))
	}

	return nil
}

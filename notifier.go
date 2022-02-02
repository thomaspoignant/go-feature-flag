package ffclient

import (
	"github.com/thomaspoignant/go-feature-flag/internal/fflog"
	"github.com/thomaspoignant/go-feature-flag/internal/notifier"
)

// getNotifiers is creating Notifier from the config
func getNotifiers(config Config, logger fflog.Logger) ([]notifier.Notifier, error) {
	notifiers := make([]notifier.Notifier, 0)

	if config.Logger != nil {
		notifiers = append(notifiers, &notifier.LogNotifier{Logger: logger})
	}

	// add all the notifiers
	for _, whConf := range config.Notifiers {
		n, err := whConf.GetNotifier(logger)
		if err != nil {
			return nil, err
		}
		notifiers = append(notifiers, n)
	}

	return notifiers, nil
}

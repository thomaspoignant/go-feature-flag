package ffclient

import (
	"github.com/thomaspoignant/go-feature-flag/ffnotifier"
	"github.com/thomaspoignant/go-feature-flag/internal/notifier"
)

// getNotifiers is creating Notifier from the config
func getNotifiers(config Config) ([]ffnotifier.Notifier, error) {
	notifiers := make([]ffnotifier.Notifier, 0)
	if config.Logger != nil {
		notifiers = append(notifiers, &notifier.LogNotifier{Logger: config.Logger})
	}

	// add all the notifiers
	for _, whConf := range config.Notifiers {
		notifier, err := whConf.GetNotifier(config)
		if err != nil {
			return nil, err
		}
		notifiers = append(notifiers, notifier)
	}

	return notifiers, nil
}

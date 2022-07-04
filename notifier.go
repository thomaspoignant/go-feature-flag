package ffclient

import (
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"github.com/thomaspoignant/go-feature-flag/notifier/logs"
)

// getNotifiers is creating Notifier from the config
func getNotifiers(config Config) ([]notifier.Notifier, error) {
	notifiers := make([]notifier.Notifier, 0)
	if config.Logger != nil {
		notifiers = append(notifiers, &logs.Notifier{Logger: config.Logger})
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

package ffclient

import (
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/thomaspoignant/go-feature-flag/internal/notifier"
)

// getNotifiers is creating Notifier from the config
func getNotifiers(config Config) ([]notifier.Notifier, error) {
	notifiers := make([]notifier.Notifier, 0)
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

	// Deprecated: we will have to remove that block when webhook will be un-supported
	wh, err := getWebhooks(config)
	if err != nil {
		return nil, err
	}
	notifiers = append(notifiers, wh...)
	// end deprecated

	return notifiers, nil
}

// Deprecated: use getNotifiers instead
func getWebhooks(config Config) ([]notifier.Notifier, error) {
	res := make([]notifier.Notifier, len(config.Webhooks))
	for index, whConf := range config.Webhooks {
		// httpClient used to call the webhook
		httpClient := http.Client{
			Timeout: 10 * time.Second,
		}

		// Deal with meta informations
		if whConf.Meta == nil {
			whConf.Meta = make(map[string]string)
		}
		hostname, _ := os.Hostname()
		whConf.Meta["hostname"] = hostname

		endpointURL, err := url.Parse(whConf.EndpointURL)
		if err != nil {
			return nil, err
		}

		w := notifier.WebhookNotifier{
			Logger:      config.Logger,
			EndpointURL: *endpointURL,
			Secret:      whConf.Secret,
			Meta:        whConf.Meta,
			HTTPClient:  &httpClient,
		}
		res[index] = &w
	}
	return res, nil
}

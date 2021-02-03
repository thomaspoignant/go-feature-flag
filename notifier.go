package ffclient

import (
	"net/http"
	"os"
	"time"

	"github.com/thomaspoignant/go-feature-flag/internal/notifier"
)

// getNotifiers is creating Notifier from the config
func (g *GoFeatureFlag) getNotifiers() []notifier.Notifier {
	var notifiers []notifier.Notifier
	if g.config.Logger != nil {
		notifiers = append(notifiers, &notifier.LogNotifier{Logger: g.config.Logger})
	}
	notifiers = append(notifiers, g.getWebhooks()...)
	return notifiers
}

func (g *GoFeatureFlag) getWebhooks() []notifier.Notifier {
	res := make([]notifier.Notifier, len(g.config.Webhooks))
	for index, whConf := range g.config.Webhooks {
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

		w := notifier.WebhookNotifier{
			Logger:     g.config.Logger,
			PayloadURL: whConf.PayloadURL,
			Secret:     whConf.Secret,
			Meta:       whConf.Meta,
			HTTPClient: &httpClient,
		}
		res[index] = &w
	}
	return res
}

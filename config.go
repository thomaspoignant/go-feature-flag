package ffclient

import (
	"context"
	"errors"
	"log"

	"github.com/thomaspoignant/go-feature-flag/internal"
	"github.com/thomaspoignant/go-feature-flag/internal/notifier"
	"github.com/thomaspoignant/go-feature-flag/internal/retriever"
)

// Config is the configuration of go-feature-flag.
// PollInterval is the interval in seconds where we gonna read the file to update the cache.
// You should also have a retriever to specify where to read the flags file.
type Config struct {
	// PollInterval (optional) Poll every X seconds
	// Default: 60 seconds
	PollInterval int

	// Logger (optional) logger use by the library
	// Default: No log
	Logger *log.Logger

	// Context (optional) used to call other services (HTTP, S3 ...)
	// Default: context.Background()
	Context context.Context

	// Retriever is the component in charge to retrieve your flag file
	Retriever Retriever

	// Notifiers (optional) is the list of notifiers called when a flag change
	Notifiers []NotifierConfig

	// FileFormat (optional) is the format of the file to retrieve (available YAML, TOML and JSON)
	// Default: YAML
	FileFormat string

	// Deprecated: Use Notifiers instead, webhooks will be delete in a future version
	Webhooks []WebhookConfig // Webhooks we should call when a flag create/update/delete

	// DataExporter (optional) is the configuration where we store how we should output the flags variations results
	DataExporter DataExporter

	// StartWithRetrieverError (optional) If true, the SDK will start even if we did not get any flags from the retriever.
	// It will serve only default values until the retriever returns the flags.
	// The init method will not return any error if the flag file is unreachable.
	// Default: false
	StartWithRetrieverError bool
}

// GetRetriever returns a retriever.FlagRetriever configure with the retriever available in the config.
func (c *Config) GetRetriever() (retriever.FlagRetriever, error) {
	if c.Retriever == nil {
		return nil, errors.New("no retriever in the configuration, impossible to get the flags")
	}
	return c.Retriever.getFlagRetriever()
}

// NotifierConfig is the interface for your notifiers.
// You can use as notifier a WebhookConfig
//
// Notifiers: []ffclient.NotifierConfig{
//        &ffclient.WebhookConfig{
//            PayloadURL: " https://example.com/hook",
//            Secret:     "Secret",
//            Meta: map[string]string{
//                "app.name": "my app",
//            },
//        },
//        // ...
//    }
type NotifierConfig interface {
	GetNotifier(config Config) (notifier.Notifier, error)
}

// WebhookConfig is the configuration of your webhook.
// we will call this URL with a POST request with the following format
//
//   {
//    "meta":{
//        "hostname": "server01"
//    },
//    "flags":{
//        "deleted": {
//            "test-flag": {
//                "rule": "key eq \"random-key\"",
//                "percentage": 100,
//                "true": true,
//                "false": false,
//                "default": false
//            }
//        },
//        "added": {
//            "test-flag3": {
//                "percentage": 5,
//                "true": "test",
//                "false": "false",
//                "default": "default"
//            }
//        },
//        "updated": {
//            "test-flag2": {
//                "old_value": {
//                    "rule": "key eq \"not-a-key\"",
//                    "percentage": 100,
//                    "true": true,
//                    "false": false,
//                    "default": false
//                },
//                "new_value": {
//                    "disable": true,
//                    "rule": "key eq \"not-a-key\"",
//                    "percentage": 100,
//                    "true": true,
//                    "false": false,
//                    "default": false
//                }
//            }
//        }
//    }
//  }
type WebhookConfig struct {
	PayloadURL string            // PayloadURL of your webhook
	Secret     string            // Secret used to sign your request body.
	Meta       map[string]string // Meta information that you want to send to your webhook (not mandatory)
}

// GetNotifier convert the configuration in a Notifier struct
func (w *WebhookConfig) GetNotifier(config Config) (notifier.Notifier, error) {
	notifier, err := notifier.NewWebhookNotifier(
		config.Logger,
		internal.DefaultHTTPClient(),
		w.PayloadURL, w.Secret, w.Meta)
	return &notifier, err
}

type SlackNotifier struct {
	SlackWebhookURL string
}

// GetNotifier convert the configuration in a Notifier struct
func (w *SlackNotifier) GetNotifier(config Config) (notifier.Notifier, error) {
	notifier := notifier.NewSlackNotifier(config.Logger, internal.DefaultHTTPClient(), w.SlackWebhookURL)
	return &notifier, nil
}

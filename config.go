package ffclient

import (
	"context"
	"errors"
	"log"

	"github.com/thomaspoignant/go-feature-flag/internal/retriever"
)

// Config is the configuration of go-feature-flag.
// PollInterval is the interval in seconds where we gonna read the file to update the cache.
// You should also have a retriever to specify where to read the flags file.
type Config struct {
	PollInterval int // Poll every X seconds
	Logger       *log.Logger
	Context      context.Context // default is context.Background()
	Retriever    Retriever
	Webhooks     []WebhookConfig // webhooks we should call when a flag create/update/delete
	FileFormat   string
}

// GetRetriever returns a retriever.FlagRetriever configure with the retriever available in the config.
func (c *Config) GetRetriever() (retriever.FlagRetriever, error) {
	if c.Retriever == nil {
		return nil, errors.New("no retriever in the configuration, impossible to get the flags")
	}
	return c.Retriever.getFlagRetriever()
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
//   }
type WebhookConfig struct {
	PayloadURL string            // PayloadURL of your webhook
	Secret     string            // Secret used to sign your request body.
	Meta       map[string]string // Meta information that you want to send to your webhook (not mandatory)
}

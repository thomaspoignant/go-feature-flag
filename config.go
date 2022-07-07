package ffclient

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/thomaspoignant/go-feature-flag/retriever"

	"github.com/thomaspoignant/go-feature-flag/notifier"
)

// Config is the configuration of go-feature-flag.
// You should also have a retriever to specify where to read the flags file.
type Config struct {
	// PollingInterval (optional) Poll every X time
	// The minimum possible is 1 second
	// Default: 60 seconds
	PollingInterval time.Duration

	// Logger (optional) logger use by the library
	// Default: No log
	Logger *log.Logger

	// Context (optional) used to call other services (HTTP, S3 ...)
	// Default: context.Background()
	Context context.Context

	// Environment (optional), can be checked in feature flag rules
	// Default: ""
	Environment string

	// Retriever is the component in charge to retrieve your flag file
	Retriever retriever.Retriever

	// Notifiers (optional) is the list of notifiers called when a flag change
	Notifiers []notifier.Notifier

	// FileFormat (optional) is the format of the file to retrieve (available YAML, TOML and JSON)
	// Default: YAML
	FileFormat string

	// DataExporter (optional) is the configuration where we store how we should output the flags variations results
	DataExporter DataExporter

	// StartWithRetrieverError (optional) If true, the SDK will start even if we did not get any flags from the retriever.
	// It will serve only default values until the retriever returns the flags.
	// The init method will not return any error if the flag file is unreachable.
	// Default: false
	StartWithRetrieverError bool

	// Offline (optional) If true, the SDK will not try to retrieve the flag file and will not export any data.
	// No notification will be send neither.
	// Default: false
	Offline bool
}

// GetRetriever returns a retriever.FlagRetriever configure with the retriever available in the config.
func (c *Config) GetRetriever() (retriever.Retriever, error) {
	if c.Retriever == nil {
		return nil, errors.New("no retriever in the configuration, impossible to get the flags")
	}
	return c.Retriever, nil
}

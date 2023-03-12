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

	// Retrievers is the list of components in charge to retrieving your flag files.
	// We are dealing with config files in order, if you have the same flag name in multiple files it will be override
	// based of the order of the retrievers in the slice.
	//
	// Note: If both Retriever and Retrievers are set, we will start by calling the Retriever and,
	// after we will use the order of Retrievers.
	Retrievers []retriever.Retriever

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
	// No notification will be sent neither.
	// Default: false
	Offline bool
}

// GetRetrievers returns a retriever.Retriever configure with the retriever available in the config.
func (c *Config) GetRetrievers() ([]retriever.Retriever, error) {
	if c.Retriever == nil && (c.Retrievers == nil || len(c.Retrievers) == 0) {
		return nil, errors.New("no retriever in the configuration, impossible to get the flags")
	}

	retrievers := make([]retriever.Retriever, 0)
	// If we have both Retriever and Retrievers fields configured we are 1st looking at what is available
	// in Retriever before looking at what is in Retrievers.
	if c.Retriever != nil {
		retrievers = append(retrievers, c.Retriever)
	}
	if c.Retrievers != nil && len(c.Retrievers) > 0 {
		retrievers = append(retrievers, c.Retrievers...)
	}
	return retrievers, nil
}

package ffclient

import (
	"context"
	"errors"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
	"log"
	"log/slog"
	"sync"
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

	// EnablePollingJitter (optional) set to true if you want to avoid having true periodicity when
	// retrieving your flags. It is useful to avoid having spike on your flag configuration storage
	// in case your application is starting multiple instance at the same time.
	// We ensure a deviation that is maximum + or - 10% of your polling interval.
	// Default: false
	EnablePollingJitter bool

	// Deprecated: Use LeveledLogger instead
	// Logger (optional) logger use by the library
	// Default: No log
	Logger *log.Logger

	// LeveledLogger (optional) logger use by the library
	// Default: No log
	LeveledLogger *slog.Logger

	// Context (optional) used to call other services (HTTP, S3 ...)
	// Default: context.Background()
	Context context.Context

	// Environment (optional) can be checked in feature flag rules
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
	// It will serve only default values until all the retrievers returns the flags.
	// The init method will not return any error if the flag file is unreachable.
	// Default: false
	StartWithRetrieverError bool

	// Offline (optional) If true, the SDK will not try to retrieve the flag file and will not export any data.
	// No notification will be sent neither.
	// Default: false
	Offline bool

	// EvaluationContextEnrichment (optional) will be merged with the evaluation context sent during the evaluation.
	// It is useful to add common attributes to all the evaluation, such as a server version, environment, ...
	//
	// All those fields will be included in the custom attributes of the evaluation context,
	// if in the evaluation context you have a field with the same name, it will override the common one.
	// Default: nil
	EvaluationContextEnrichment map[string]interface{}

	// PersistentFlagConfigurationFile (optional) if set GO Feature Flag will store flags configuration in this file
	//  to be able to serve the flags even if none of the retrievers is available during starting time.
	//
	// By default, the flag configuration is not persisted and stays on the retriever system. By setting a file here,
	// you ensure that GO Feature Flag will always start with a configuration but which can be out-dated.
	PersistentFlagConfigurationFile string

	// offlineMutex is a mutex to protect the Offline field.
	offlineMutex *sync.RWMutex

	// internalLogger is the logger used by the library everywhere
	// this logger is a superset of the logging system to be able to migrate easily to slog.
	internalLogger *fflog.FFLogger
}

// GetRetrievers returns a retriever.Retriever configure with the retriever available in the config.
func (c *Config) GetRetrievers() ([]retriever.Retriever, error) {
	if c.Retriever == nil && len(c.Retrievers) == 0 {
		return nil, errors.New("no retriever in the configuration, impossible to get the flags")
	}

	retrievers := make([]retriever.Retriever, 0)
	// If we have both Retriever and Retrievers fields configured, we are 1st looking at what is available
	// in Retriever before looking at what is in Retrievers.
	if c.Retriever != nil {
		retrievers = append(retrievers, c.Retriever)
	}
	if len(c.Retrievers) > 0 {
		retrievers = append(retrievers, c.Retrievers...)
	}
	return retrievers, nil
}

// SetOffline set GO Feature Flag in offline mode.
func (c *Config) SetOffline(control bool) {
	if c.offlineMutex == nil {
		c.offlineMutex = &sync.RWMutex{}
	}
	c.offlineMutex.Lock()
	defer c.offlineMutex.Unlock()
	c.Offline = control
}

// IsOffline return if the GO Feature Flag is in offline mode.
func (c *Config) IsOffline() bool {
	if c.offlineMutex == nil {
		c.offlineMutex = &sync.RWMutex{}
	}
	c.offlineMutex.RLock()
	defer c.offlineMutex.RUnlock()
	return c.Offline
}

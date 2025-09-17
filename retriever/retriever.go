package retriever

import (
	"context"
	"log"

	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

// Retriever is the interface to create a Retriever to load you flags.
type Retriever interface {
	// Retrieve function is supposed to load the file and to return a []byte of your flag configuration file.
	Retrieve(ctx context.Context) ([]byte, error)
}

// InitializableRetrieverLegacy is an extended version of the retriever that can be initialized and shutdown.
type InitializableRetrieverLegacy interface {
	CommonInitializableRetriever
	Init(ctx context.Context, logger *log.Logger) error
}

// InitializableRetriever is an extended version of the retriever that can be initialized and shutdown.
type InitializableRetriever interface {
	CommonInitializableRetriever
	Init(ctx context.Context, logger *fflog.FFLogger) error
}

// InitializableRetrieverWithFlagset is an extended version of the retriever that can be initialized and shutdown.
// It is used to initialize the retriever with a specific flagset.
type InitializableRetrieverWithFlagset interface {
	CommonInitializableRetriever
	Init(ctx context.Context, logger *fflog.FFLogger, flagset *string) error
}

// CommonInitializableRetriever is the common interface for all versions of
// retrievers that can be initialized and shutdown.
type CommonInitializableRetriever interface {
	Retriever
	Shutdown(ctx context.Context) error
	Status() Status
}

// Status is the status of the retriever.
// It can be used to check if the retriever is ready to be used.
// If not ready, we wi will not use it.
type Status = string

const (
	// RetrieverReady is the status when the retriever is ready to be used.
	RetrieverReady Status = "READY"
	// RetrieverNotReady is the status when the retriever is not ready yet to be used.
	RetrieverNotReady Status = "NOT_READY"
	// RetrieverError is the status when the retriever is in error.
	RetrieverError Status = "ERROR"
)

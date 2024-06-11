package retriever

import (
	"context"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
	"log"
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
	RetrieverReady    Status = "READY"
	RetrieverNotReady Status = "NOT_READY"
	RetrieverError    Status = "ERROR"
)

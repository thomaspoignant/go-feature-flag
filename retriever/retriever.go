package retriever

import (
	"context"
	"log"
)

// Retriever is the interface to create a Retriever to load you flags.
type Retriever interface {
	// Retrieve function is supposed to load the file and to return a []byte of your flag configuration file.
	Retrieve(ctx context.Context) ([]byte, error)
}

// InitializableRetriever is an extended version of the retriever that can be initialized and shutdown.
type InitializableRetriever interface {
	Retrieve(ctx context.Context) ([]byte, error)
	Init(ctx context.Context, logger *log.Logger) error
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

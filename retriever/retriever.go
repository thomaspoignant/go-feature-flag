package retriever

import (
	"context"
)

// Retriever is the interface to create a Retriever to load you flags.
type Retriever interface {
	// Retrieve function is supposed to load the file and to return a []byte of your flag configuration file.
	Retrieve(ctx context.Context) ([]byte, error)
}

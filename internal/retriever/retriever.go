package retriever

import "context"

// FlagRetriever is an interface that force to have a Retrieve() function for
// different way of getting the config file.
type FlagRetriever interface {
	Retrieve(ctx context.Context) ([]byte, error)
}

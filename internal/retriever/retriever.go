package retriever

// FlagRetriever is an interface that force to have a Retrieve() function for
// different way of getting the config file.
type FlagRetriever interface {
	Retrieve() ([]byte, error)
}

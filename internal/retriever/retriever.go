package retriever

type FlagRetriever interface {
	Retrieve() ([]byte, error)
}

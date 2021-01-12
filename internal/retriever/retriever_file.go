package retriever

import (
	"context"
	"io/ioutil"
)

// NewLocalRetriever is the retriever for local file.
func NewLocalRetriever(path string) FlagRetriever {
	return &localRetriever{path}
}

type localRetriever struct {
	path string
}

func (l *localRetriever) Retrieve(ctx context.Context) ([]byte, error) {
	content, err := ioutil.ReadFile(l.path)
	if err != nil {
		return nil, err
	}
	return content, nil
}

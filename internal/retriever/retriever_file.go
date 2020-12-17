package retriever

import (
	"io/ioutil"
)

// NewLocalRetriever is the retriever for local file.
func NewLocalRetriever(path string) FlagRetriever {
	return &localRetriever{path}
}

type localRetriever struct {
	path string
}

func (l *localRetriever) Retrieve() ([]byte, error) {
	content, err := ioutil.ReadFile(l.path)
	if err != nil {
		return nil, err
	}
	return content, nil
}

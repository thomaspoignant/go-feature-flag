package file

import (
	"context"
	"io/ioutil"
)

// Retriever is a configuration struct for a local flat file.
type Retriever struct {
	Path string
}

// Retrieve is reading the file and return the content
func (r *Retriever) Retrieve(ctx context.Context) ([]byte, error) {
	content, err := ioutil.ReadFile(r.Path)
	if err != nil {
		return nil, err
	}
	return content, nil
}

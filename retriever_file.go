package ffclient

import (
	"context"
	"io/ioutil"
)

// FileRetriever is a configuration struct for a local flat file.
type FileRetriever struct {
	Path string
}

// Retrieve is reading the file and return the content
func (r *FileRetriever) Retrieve(ctx context.Context) ([]byte, error) {
	content, err := ioutil.ReadFile(r.Path)
	if err != nil {
		return nil, err
	}
	return content, nil
}

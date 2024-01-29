package initializableretriever

import (
	"context"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"log"
	"os"
)

func NewMockInitializableRetriever(path string, status retriever.Status) Retriever {
	return Retriever{
		context: context.Background(),
		Path:    path,
		status:  status,
	}
}

// Retriever is a mock provider, that create a file as init step and delete it at shutdown.
type Retriever struct {
	context context.Context
	Path    string
	status  retriever.Status
}

// Retrieve is reading the file and return the content
func (r *Retriever) Retrieve(_ context.Context) ([]byte, error) {
	content, err := os.ReadFile(r.Path)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func (r *Retriever) Init(_ context.Context, _ *log.Logger) error {
	yamlString := `flag-xxxx-123:
  variations:
    A: true
    B: false
  defaultRule:
    variation: A`

	yamlBytes := []byte(yamlString)
	return os.WriteFile(r.Path, yamlBytes, 0600)
}

func (r *Retriever) Shutdown(_ context.Context) error {
	return os.Remove(r.Path)
}

func (r *Retriever) Status() retriever.Status {
	return r.status
}

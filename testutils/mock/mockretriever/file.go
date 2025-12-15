package mockretriever

import (
	"context"
	"os"

	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

// FileInitializableRetriever is the existing file-based mock that creates/deletes files
// This maintains compatibility with existing tests
type FileInitializableRetriever struct {
	Path          string
	CurrentStatus retriever.Status
}

func NewFileInitializableRetriever(path string, status retriever.Status) *FileInitializableRetriever {
	return &FileInitializableRetriever{
		Path:          path,
		CurrentStatus: status,
	}
}

// Retrieve reads the file and returns the content
func (r *FileInitializableRetriever) Retrieve(_ context.Context) ([]byte, error) {
	content, err := os.ReadFile(r.Path)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func (r *FileInitializableRetriever) Init(_ context.Context, _ *fflog.FFLogger) error {
	yamlString := `flag-xxxx-123:
  variations:
    A: true
    B: false
  defaultRule:
    variation: A`

	yamlBytes := []byte(yamlString)
	return os.WriteFile(r.Path, yamlBytes, 0600)
}

func (r *FileInitializableRetriever) Shutdown(_ context.Context) error {
	return os.Remove(r.Path)
}

func (r *FileInitializableRetriever) Status() retriever.Status {
	return r.CurrentStatus
}

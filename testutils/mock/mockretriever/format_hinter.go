package mockretriever

import "context"

// FormatHintingRetriever is a mock that implements retriever.FormatHinter so
// tests can verify that the manager honors a retriever-declared output format
// regardless of the global FileFormat configuration.
type FormatHintingRetriever struct {
	Name    string
	Format  string
	Content []byte
}

func NewFormatHintingRetriever(name, format string, content []byte) *FormatHintingRetriever {
	return &FormatHintingRetriever{
		Name:    name,
		Format:  format,
		Content: content,
	}
}

func (m *FormatHintingRetriever) Retrieve(_ context.Context) ([]byte, error) {
	return m.Content, nil
}

func (m *FormatHintingRetriever) OutputFormat() string {
	return m.Format
}

func (m *FormatHintingRetriever) GetName() string {
	return m.Name
}

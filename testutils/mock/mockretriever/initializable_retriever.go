// nolint: dupl
package mockretriever

import (
	"context"
	"errors"

	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

// InitializableRetriever implements the standard InitializableRetriever interface
type InitializableRetriever struct {
	Name           string
	InitShouldFail bool
	RetrieveCalled bool
	InitCalled     bool
	ShutdownCalled bool
	StatusCalled   bool
	ReceivedLogger *fflog.FFLogger
}

func NewInitializableRetriever(name string) *InitializableRetriever {
	return &InitializableRetriever{Name: name}
}

func (m *InitializableRetriever) Retrieve(_ context.Context) ([]byte, error) {
	m.RetrieveCalled = true
	return []byte(defaultFlagConfig), nil
}

func (m *InitializableRetriever) Init(_ context.Context, logger *fflog.FFLogger) error {
	m.InitCalled = true
	m.ReceivedLogger = logger
	if m.InitShouldFail {
		return errors.New("initialization failed")
	}
	return nil
}

func (m *InitializableRetriever) Shutdown(_ context.Context) error {
	m.ShutdownCalled = true
	return nil
}

func (m *InitializableRetriever) Status() retriever.Status {
	m.StatusCalled = true
	if m.InitShouldFail {
		return retriever.RetrieverError
	}
	return retriever.RetrieverReady
}

func (m *InitializableRetriever) GetName() string {
	return m.Name
}

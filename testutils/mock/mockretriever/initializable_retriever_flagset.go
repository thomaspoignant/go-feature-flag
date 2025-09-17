package mockretriever

import (
	"context"
	"errors"

	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

// InitializableRetrieverWithFlagset implements the InitializableRetrieverWithFlagset interface
type InitializableRetrieverWithFlagset struct {
	Name            string
	InitShouldFail  bool
	RetrieveCalled  bool
	InitCalled      bool
	ShutdownCalled  bool
	StatusCalled    bool
	ReceivedLogger  *fflog.FFLogger
	ReceivedFlagset *string
}

func NewInitializableRetrieverWithFlagset(name string) *InitializableRetrieverWithFlagset {
	return &InitializableRetrieverWithFlagset{Name: name}
}

func (m *InitializableRetrieverWithFlagset) Retrieve(_ context.Context) ([]byte, error) {
	m.RetrieveCalled = true
	return []byte(defaultFlagConfig), nil
}

func (m *InitializableRetrieverWithFlagset) Init(_ context.Context, logger *fflog.FFLogger, flagset *string) error {
	m.InitCalled = true
	m.ReceivedLogger = logger
	m.ReceivedFlagset = flagset
	if m.InitShouldFail {
		return errors.New("initialization failed")
	}
	return nil
}

func (m *InitializableRetrieverWithFlagset) Shutdown(_ context.Context) error {
	m.ShutdownCalled = true
	return nil
}

func (m *InitializableRetrieverWithFlagset) Status() retriever.Status {
	m.StatusCalled = true
	if m.InitShouldFail {
		return retriever.RetrieverError
	}
	return retriever.RetrieverReady
}

func (m *InitializableRetrieverWithFlagset) GetName() string {
	return m.Name
}

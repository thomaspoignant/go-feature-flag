package mockretriever

import (
	"context"
	"errors"

	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

// ContextAwareRetriever is a mock that respects context cancellation
type ContextAwareRetriever struct{}

func NewContextAwareRetriever() *ContextAwareRetriever {
	return &ContextAwareRetriever{}
}

func (m *ContextAwareRetriever) Retrieve(ctx context.Context) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return []byte(defaultFlagConfig), nil
	}
}

func (m *ContextAwareRetriever) Init(ctx context.Context, _ *fflog.FFLogger) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func (m *ContextAwareRetriever) Shutdown(_ context.Context) error {
	return nil
}

func (m *ContextAwareRetriever) Status() retriever.Status {
	return retriever.RetrieverReady
}

// StatusChangingRetriever allows testing status changes during operations
type StatusChangingRetriever struct {
	StatusCalled  bool
	CurrentStatus retriever.Status
}

func NewStatusChangingRetriever() *StatusChangingRetriever {
	return &StatusChangingRetriever{}
}

func (m *StatusChangingRetriever) Retrieve(_ context.Context) ([]byte, error) {
	return []byte(defaultFlagConfig), nil
}

func (m *StatusChangingRetriever) Init(_ context.Context, _ *fflog.FFLogger) error {
	m.CurrentStatus = retriever.RetrieverReady
	return nil
}

func (m *StatusChangingRetriever) Shutdown(_ context.Context) error {
	m.CurrentStatus = retriever.RetrieverNotReady
	return nil
}

func (m *StatusChangingRetriever) Status() retriever.Status {
	m.StatusCalled = true
	if m.CurrentStatus == "" {
		return retriever.RetrieverNotReady
	}
	return m.CurrentStatus
}

// RecoverableRetriever can fail initially but succeed on retry
type RecoverableRetriever struct {
	Name          string
	FailFirstInit bool
	InitCalled    bool
	InitCallCount int
}

func NewRecoverableRetriever(name string) *RecoverableRetriever {
	return &RecoverableRetriever{Name: name}
}

func (m *RecoverableRetriever) Retrieve(_ context.Context) ([]byte, error) {
	return []byte(defaultFlagConfig), nil
}

func (m *RecoverableRetriever) Init(_ context.Context, _ *fflog.FFLogger) error {
	m.InitCalled = true
	m.InitCallCount++

	if m.FailFirstInit && m.InitCallCount == 1 {
		return errors.New("first init failed")
	}
	return nil
}

func (m *RecoverableRetriever) Shutdown(_ context.Context) error {
	return nil
}

func (m *RecoverableRetriever) Status() retriever.Status {
	if m.FailFirstInit && m.InitCallCount == 1 {
		return retriever.RetrieverError
	}
	return retriever.RetrieverReady
}

func (m *RecoverableRetriever) GetName() string {
	return m.Name
}

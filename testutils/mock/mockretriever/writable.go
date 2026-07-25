package mockretriever

import (
	"context"

	"github.com/thomaspoignant/go-feature-flag/retriever"
)

// WritableRetriever is a configurable mock implementing both retriever.Retriever and
// retriever.WritableRetriever, used to test the flag-management write path without a
// real database.
type WritableRetriever struct {
	Name string

	GetFlagFunc    func(ctx context.Context, flagKey string) ([]byte, string, error)
	CreateFlagFunc func(ctx context.Context, flagKey string, definition []byte) (string, error)
	UpsertFlagFunc func(
		ctx context.Context, flagKey string, definition []byte, ifMatch *string,
	) (bool, string, error)
	DeleteFlagFunc func(ctx context.Context, flagKey string, ifMatch *string) error
	SourceValue    string
	RetrieveFunc   func(ctx context.Context) ([]byte, error)
}

func NewWritableRetriever(name string) *WritableRetriever {
	return &WritableRetriever{Name: name, SourceValue: "mock:" + name}
}

func (m *WritableRetriever) Retrieve(ctx context.Context) ([]byte, error) {
	if m.RetrieveFunc != nil {
		return m.RetrieveFunc(ctx)
	}
	return []byte(defaultFlagConfig), nil
}

func (m *WritableRetriever) GetFlag(ctx context.Context, flagKey string) ([]byte, string, error) {
	if m.GetFlagFunc != nil {
		return m.GetFlagFunc(ctx, flagKey)
	}
	return nil, "", retriever.ErrFlagNotFound
}

func (m *WritableRetriever) CreateFlag(ctx context.Context, flagKey string, definition []byte) (string, error) {
	if m.CreateFlagFunc != nil {
		return m.CreateFlagFunc(ctx, flagKey, definition)
	}
	return "", nil
}

func (m *WritableRetriever) UpsertFlag(
	ctx context.Context, flagKey string, definition []byte, ifMatch *string,
) (bool, string, error) {
	if m.UpsertFlagFunc != nil {
		return m.UpsertFlagFunc(ctx, flagKey, definition, ifMatch)
	}
	return false, "", nil
}

func (m *WritableRetriever) DeleteFlag(ctx context.Context, flagKey string, ifMatch *string) error {
	if m.DeleteFlagFunc != nil {
		return m.DeleteFlagFunc(ctx, flagKey, ifMatch)
	}
	return nil
}

func (m *WritableRetriever) Source() string {
	return m.SourceValue
}

package retriever

import (
	"context"
	"fmt"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
	"log/slog"
)

// Manager is a struct that managed the retrievers.
type Manager struct {
	ctx              context.Context
	retrievers       []Retriever
	onErrorRetriever []Retriever
	logger           *fflog.FFLogger
}

// NewManager create a new Manager.
func NewManager(ctx context.Context, retrievers []Retriever, logger *fflog.FFLogger) *Manager {
	return &Manager{
		ctx:              ctx,
		retrievers:       retrievers,
		onErrorRetriever: make([]Retriever, 0),
		logger:           logger,
	}
}

// Init the retrievers.
// This function will call the Init function of the retrievers that implements the InitializableRetriever interface.
func (m *Manager) Init(ctx context.Context) error {
	return m.initRetrievers(ctx, m.retrievers)
}

// initRetrievers is a helper function to initialize the retrievers.
func (m *Manager) initRetrievers(ctx context.Context, retrieversToInit []Retriever) error {
	m.onErrorRetriever = make([]Retriever, 0)
	for _, retriever := range retrieversToInit {
		if r, ok := retriever.(InitializableRetrieverLegacy); ok {
			err := r.Init(ctx, m.logger.GetLogLogger(slog.LevelError))
			if err != nil {
				m.onErrorRetriever = append(m.onErrorRetriever, retriever)
			}
		}

		if r, ok := retriever.(InitializableRetriever); ok {
			err := r.Init(ctx, m.logger)
			if err != nil {
				m.onErrorRetriever = append(m.onErrorRetriever, retriever)
			}
		}
	}
	if len(m.onErrorRetriever) > 0 {
		return fmt.Errorf("error while initializing the retrievers: %v", m.onErrorRetriever)
	}
	return nil
}

// Shutdown the retrievers.
// This function will call the Shutdown function of the retrievers that implements the InitializableRetriever interface.
func (m *Manager) Shutdown(ctx context.Context) error {
	onErrorRetriever := make([]Retriever, 0)
	for _, retriever := range m.retrievers {
		if r, ok := retriever.(CommonInitializableRetriever); ok {
			err := r.Shutdown(ctx)
			if err != nil {
				onErrorRetriever = append(onErrorRetriever, retriever)
			}
		}
	}
	if len(onErrorRetriever) > 0 {
		return fmt.Errorf("error while shutting down the retrievers: %v", onErrorRetriever)
	}
	return nil
}

// GetRetrievers return the retrievers.
// If an error occurred during the initialization of the retrievers, we will return the retrievers that are ready.
func (m *Manager) GetRetrievers() []Retriever {
	if len(m.onErrorRetriever) > 0 {
		_ = m.initRetrievers(m.ctx, m.onErrorRetriever)
	}
	return m.retrievers
}

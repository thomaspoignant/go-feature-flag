package exporter

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"sync"
	"time"

	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

const (
	defaultFlushInterval    = 60 * time.Second
	defaultMaxEventInMemory = int64(100000)
)

// ExporterConfig holds the configuration for an individual exporter
type Config struct {
	Exporter         CommonExporter
	FlushInterval    time.Duration
	MaxEventInMemory int64
}

// ExporterState maintains the state for a single exporter
type State struct {
	config    Config
	ticker    *time.Ticker
	lastIndex int // Index of the last processed event
}

// Scheduler handles data collection for one or more exporters
type Scheduler struct {
	sharedCache     []FeatureEvent
	bulkExporters   map[CommonExporter]*State // Only bulk exporters that need periodic flushing
	directExporters []CommonExporter          // Non-bulk exporters that flush immediately
	mutex           sync.Mutex
	daemonChan      chan struct{}
	logger          *fflog.FFLogger
	ctx             context.Context
}

// NewScheduler creates a new scheduler that handles one exporter
func NewScheduler(ctx context.Context, flushInterval time.Duration, maxEventInMemory int64,
	exp CommonExporter, logger *fflog.FFLogger,
) *Scheduler {
	// Convert single exporter parameters to ExporterConfig
	config := Config{
		Exporter:         exp,
		FlushInterval:    flushInterval,
		MaxEventInMemory: maxEventInMemory,
	}
	return NewMultiScheduler(ctx, []Config{config}, logger)
}

// NewMultiScheduler creates a scheduler that handles multiple exporters
func NewMultiScheduler(ctx context.Context, exporterConfigs []Config, logger *fflog.FFLogger,
) *Scheduler {
	if ctx == nil {
		ctx = context.Background()
	}

	bulkExporters := make(map[CommonExporter]*State)
	directExporters := make([]CommonExporter, 0)

	for _, config := range exporterConfigs {
		if config.FlushInterval == 0 {
			config.FlushInterval = defaultFlushInterval
		}
		if config.MaxEventInMemory == 0 {
			config.MaxEventInMemory = defaultMaxEventInMemory
		}

		if config.Exporter.IsBulk() {
			state := &State{
				config:    config,
				lastIndex: -1,
				ticker:    time.NewTicker(config.FlushInterval),
			}
			bulkExporters[config.Exporter] = state
		} else {
			directExporters = append(directExporters, config.Exporter)
		}
	}

	return &Scheduler{
		sharedCache:     make([]FeatureEvent, 0),
		bulkExporters:   bulkExporters,
		directExporters: directExporters,
		mutex:           sync.Mutex{},
		daemonChan:      make(chan struct{}),
		logger:          logger,
		ctx:             ctx,
	}
}

// AddEvent adds an event to the shared cache and handles immediate export for non-bulk exporters
func (s *Scheduler) AddEvent(event FeatureEvent) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Handle non-bulk exporters immediately
	for _, exporter := range s.directExporters {
		err := sendEvents(s.ctx, exporter, s.logger, []FeatureEvent{event})
		if err != nil {
			s.logger.Error(err.Error())
		}
	}

	// If we have no bulk exporters, we're done
	if len(s.bulkExporters) == 0 {
		return
	}

	// Add event to shared cache for bulk exporters
	s.sharedCache = append(s.sharedCache, event)
	currentIndex := len(s.sharedCache) - 1

	// Check if any bulk exporters need to flush due to max events
	for _, state := range s.bulkExporters {
		pendingCount := currentIndex - state.lastIndex
		if state.config.MaxEventInMemory > 0 && int64(pendingCount) >= state.config.MaxEventInMemory {
			s.flushExporter(state)
		}
	}

	// Clean up events that have been processed by all exporters
	s.cleanupProcessedEvents()
}

// getPendingEvents returns events that haven't been processed by this exporter
func (s *Scheduler) getPendingEvents(state *State) []FeatureEvent {
	if state.lastIndex+1 >= len(s.sharedCache) {
		return nil
	}
	return s.sharedCache[state.lastIndex+1:]
}

// flushExporter sends pending events to the specified exporter
func (s *Scheduler) flushExporter(state *State) {
	pendingEvents := s.getPendingEvents(state)
	if len(pendingEvents) == 0 {
		return
	}

	err := sendEvents(s.ctx, state.config.Exporter, s.logger, pendingEvents)
	if err != nil {
		s.logger.Error(err.Error())
		return
	}

	// Update last processed index
	state.lastIndex = len(s.sharedCache) - 1
}

// cleanupProcessedEvents removes events that have been processed by all bulk exporters
func (s *Scheduler) cleanupProcessedEvents() {
	// If no bulk exporters, we can clear the cache
	if len(s.bulkExporters) == 0 {
		s.sharedCache = make([]FeatureEvent, 0)
		return
	}

	// Find minimum lastIndex among bulk exporters
	minIndex := len(s.sharedCache)
	for _, state := range s.bulkExporters {
		if state.lastIndex < minIndex {
			minIndex = state.lastIndex
		}
	}

	// If all exporters have processed some events, we can remove them
	if minIndex > 0 {
		// Keep events from minIndex+1 onwards
		s.sharedCache = s.sharedCache[minIndex+1:]
		// Update lastIndex for all exporters
		for _, state := range s.bulkExporters {
			state.lastIndex -= (minIndex + 1)
		}
	}
}

// StartDaemon starts the periodic flush for bulk exporters
func (s *Scheduler) StartDaemon() {
	// If no bulk exporters, no need for daemon
	if len(s.bulkExporters) == 0 {
		return
	}

	for {
		select {
		case <-s.daemonChan:
			return
		default:
			s.mutex.Lock()
			for _, state := range s.bulkExporters {
				select {
				case <-state.ticker.C:
					s.flushExporter(state)
				default:
					// Continue if this ticker hasn't triggered
				}
			}
			s.cleanupProcessedEvents()
			s.mutex.Unlock()
			// Small sleep to prevent busy waiting
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// Close stops all tickers and flushes remaining events
func (s *Scheduler) Close() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Stop all tickers and flush bulk exporters
	for _, state := range s.bulkExporters {
		state.ticker.Stop()
		s.flushExporter(state)
	}

	close(s.daemonChan)
	s.sharedCache = nil
}

// GetLogger returns the logger used by the scheduler
func (s *Scheduler) GetLogger(level slog.Level) *log.Logger {
	if s.logger == nil {
		return nil
	}
	return s.logger.GetLogLogger(level)
}

func sendEvents(ctx context.Context, exporter CommonExporter, logger *fflog.FFLogger, events []FeatureEvent) error {
	if len(events) == 0 {
		return nil
	}

	switch exp := exporter.(type) {
	case DeprecatedExporter:
		var legacyLogger *log.Logger
		if logger != nil {
			legacyLogger = logger.GetLogLogger(slog.LevelError)
		}
		// use dc exporter as a DeprecatedExporter
		err := exp.Export(ctx, legacyLogger, events)
		slog.Warn("You are using an exporter with the old logger."+
			"Please update your custom exporter to comply to the new Exporter interface.",
			slog.Any("err", err))
		if err != nil {
			return fmt.Errorf("error while exporting data: %w", err)
		}
		break
	case Exporter:
		err := exp.Export(ctx, logger, events)
		if err != nil {
			return fmt.Errorf("error while exporting data: %w", err)
		}
		break
	default:
		return fmt.Errorf("this is not a valid exporter")
	}
	return nil
}

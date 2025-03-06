package exporter

//
//import (
//	"context"
//	"fmt"
//	"github.com/google/uuid"
//	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
//	"log"
//	"log/slog"
//	"time"
//)
//
//func NewDataExporter[T any](ctx context.Context, flushInterval time.Duration, maxEventInMemory int64,
//	exporters []DataExporter, logger *fflog.FFLogger) Manager[T] {
//	if ctx == nil {
//		ctx = context.Background()
//	}
//
//	if flushInterval == 0 {
//		flushInterval = defaultFlushInterval
//	}
//
//	if maxEventInMemory == 0 {
//		maxEventInMemory = defaultMaxEventInMemory
//	}
//
//	bulkExporters := make([]singleExporter[T],0)
//	consumers := make([]string, 0)
//	liveExporters := make([]DataExporter, 0)
//	eventStore := NewEventStore[T](consumers, 30*time.Second)z
//
//
//
//	for _, exporter := range exporters {
//		if exporter.Exporter.IsBulk() {
//			exp := singleExporter[T]{
//				ctx:        ctx,
//				consumerId: uuid.New().String(),
//				eventStore: eventStore,
//				logger:    logger,
//				exporter:   exporter.Exporter,
//			}
//			consumers = append(consumers, consumerId)
//		} else {
//			liveExporters = append(liveExporters, exporter)
//		}
//	}
//
//	return &manager[T]{
//		bulkExporters:    bulkExporters,
//		liveExporters:    liveExporters,
//		localCache:
//		ctx:              ctx,
//		logger:           logger,
//		maxEventInMemory: maxEventInMemory,
//	}
//}
//
//type manager[T any] struct {
//	localCache       EventStore[T]
//	bulkExporters    singleExporter[T]
//	liveExporters    []DataExporter
//	ctx              context.Context
//	logger           *fflog.FFLogger
//	maxEventInMemory int64
//}
//
//func (m *manager[T]) StartDaemon() {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (m *manager[T]) Close() {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (m *manager[T]) AddEvent(event T) {
//	for _, exporter := range m.liveExporters {
//		switch events := any(event).(type) {
//		case FeatureEvent:
//			go func() {
//				err := sendEvents(m.ctx, exporter.Exporter, m.logger, []FeatureEvent{events})
//				if err != nil {
//					m.logger.Error(err.Error())
//				}
//			}()
//			break
//		default:
//			m.logger.Error("this is not a valid exporter")
//		}
//	}
//
//	m.localCache.Add(event)
//	//for consumerId, exporter := range m.bulkExporters {
//	//	eventCount, err := m.localCache.GetPendingEventCount(consumerId)
//	//	if err != nil {
//	//		// log something here
//	//		m.logger.LeveledLogger.Error("error while getting pending event count", err)
//	//		continue
//	//	}
//	//	if eventCount >= m.maxEventInMemory {
//	//		m.flush(consumerId)
//	//	}
//	//}
//
//}
//
//type singleExporter[T any] struct {
//	ctx         context.Context
//	consumerId string
//	eventStore EventStore[T]
//	logger     *fflog.FFLogger
//	exporter    CommonExporter
//
//	daemonChan chan struct{}
//	ticker     *time.Ticker
//}
//
//func (s *singleExporter[T]) flush() {
//	eventList, err := s.eventStore.FetchPendingEvents(s.consumerId)
//	if err != nil {
//		// log something here
//		s.logger.Error("error while fetching pending events", err)
//		return
//	}
//
//	err = s.sendEvents(s.ctx, eventList.Events)
//	if err != nil {
//		s.logger.Error(err.Error())
//		return
//	}
//	err = s.eventStore.UpdateConsumerOffset(s.consumerId, eventList.NewOffset)
//	if err != nil {
//		s.logger.Error(err.Error())
//	}
//}
//
//func (s *singleExporter[T]) StartDaemon() {
//	for {
//		select {
//		case <-s.ticker.C:
//			// send data and clear local cache
//			s.flush()
//		case <-s.daemonChan:
//			// stop the daemon
//			return
//		}
//	}
//}
//
//func (s *singleExporter[T]) sendEvents(ctx context.Context, events []T) error {
//	if len(events) == 0 {
//		return nil
//	}
//	switch exp := s.exporter.(type) {
//	case DeprecatedExporter:
//		var legacyLogger *log.Logger
//		if s.logger != nil {
//			legacyLogger = s.logger.GetLogLogger(slog.LevelError)
//		}
//		switch events := any(events).(type) {
//		case []FeatureEvent:
//			// use dc exporter as a DeprecatedExporter
//			err := exp.Export(ctx, legacyLogger, events)
//			slog.Warn("You are using an exporter with the old logger."+
//				"Please update your custom exporter to comply to the new Exporter interface.",
//				slog.Any("err", err))
//			if err != nil {
//				return fmt.Errorf("error while exporting data: %w", err)
//			}
//		default:
//			return fmt.Errorf("this is not a valid exporter")
//		}
//		break
//	case Exporter:
//		switch events := any(events).(type) {
//		case []FeatureEvent:
//			err := exp.Export(ctx, s.logger, events)
//			if err != nil {
//				return fmt.Errorf("error while exporting data: %w", err)
//			}
//		default:
//			return fmt.Errorf("this is not a valid exporter")
//		}
//		break
//	default:
//		return fmt.Errorf("this is not a valid exporter")
//	}
//	return nil
//}
//
//func (s *singleExporter[T]) Stop() {
//	s.ticker.Stop()
//	close(s.daemonChan)
//	s.flush()
//}

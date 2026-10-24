package flaghistory

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"go.uber.org/zap"
)

const (
	defaultQueueSize        = 1000
	defaultBackoffMin       = time.Second
	defaultBackoffMax       = time.Minute
	defaultFlushTimeout     = 10 * time.Second
	defaultPurgeInterval    = 24 * time.Hour
	defaultOperationTimeout = 15 * time.Second
)

// ErrNotReady is returned when the history database is not reachable yet.
var ErrNotReady = errors.New("flag history storage is not ready")

// Config is the configuration of the History.
type Config struct {
	// Connect opens the storage, it is retried until it succeeds.
	Connect func(ctx context.Context) (Store, error)
	// Retention is the age after which the entries are deleted, 0 disables the purge.
	Retention time.Duration
	// Source identifies this relay proxy instance in the history.
	Source string
	// QueueSize is the number of changes kept in memory while the storage is unavailable.
	QueueSize int
	Logger    *zap.Logger

	backoffMin   time.Duration
	backoffMax   time.Duration
	flushTimeout time.Duration
	purgeEvery   time.Duration
}

// History records the modifications of the flags in a durable storage.
// It never blocks the flags polling nor the evaluations: a change is queued in memory, and
// written by a background worker that retries until the storage is back.
type History struct {
	cfg    Config
	queue  chan Entry
	logger *zap.Logger

	mu    sync.RWMutex
	store Store

	dropped atomic.Int64
	ready   chan struct{}
	stop    chan struct{}
	stopCtx context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	once    sync.Once
}

// New creates the History, Start must be called to begin the work.
func New(cfg Config) *History {
	if cfg.QueueSize <= 0 {
		cfg.QueueSize = defaultQueueSize
	}
	if cfg.backoffMin <= 0 {
		cfg.backoffMin = defaultBackoffMin
	}
	if cfg.backoffMax <= 0 {
		cfg.backoffMax = defaultBackoffMax
	}
	if cfg.flushTimeout <= 0 {
		cfg.flushTimeout = defaultFlushTimeout
	}
	if cfg.purgeEvery <= 0 {
		cfg.purgeEvery = defaultPurgeInterval
	}
	logger := cfg.Logger
	if logger == nil {
		logger = zap.NewNop()
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &History{
		cfg:     cfg,
		queue:   make(chan Entry, cfg.QueueSize),
		logger:  logger,
		ready:   make(chan struct{}),
		stop:    make(chan struct{}),
		stopCtx: ctx,
		cancel:  cancel,
	}
}

// Start connects to the storage (in the background, the relay proxy starts even if the
// database is down) and starts writing the queued changes.
func (h *History) Start() {
	h.wg.Add(1)
	go h.run()
	if h.cfg.Retention > 0 {
		h.wg.Add(1)
		go h.purgeLoop()
	}
}

// Store returns the storage, or ErrNotReady while the connection is not established.
func (h *History) Store() (Store, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.store == nil {
		return nil, ErrNotReady
	}
	return h.store, nil
}

// Dropped is the number of changes lost because the queue was full.
func (h *History) Dropped() int64 { return h.dropped.Load() }

// Ping checks the storage, used by the health endpoint.
func (h *History) Ping(ctx context.Context) error {
	s, err := h.Store()
	if err != nil {
		return err
	}
	return s.Ping(ctx)
}

// NotifierFor returns the notifier to plug on the flagset.
func (h *History) NotifierFor(flagset string) notifier.Notifier {
	return &historyNotifier{history: h, flagset: flagset}
}

type historyNotifier struct {
	history *History
	flagset string
}

// Notify queues the changes and returns immediately.
func (n *historyNotifier) Notify(diff notifier.DiffCache) error {
	entries, err := EntriesFromDiff(n.flagset, n.history.cfg.Source, diff)
	if err != nil {
		return err
	}
	for _, e := range entries {
		n.history.enqueue(e)
	}
	return nil
}

func (h *History) enqueue(e Entry) {
	e.DetectedAt = time.Now().UTC()
	select {
	case h.queue <- e:
	default:
		h.dropped.Add(1)
		h.logger.Error("flag history queue is full, change lost",
			zap.String("flagset", e.Flagset), zap.String("flag", e.FlagKey))
	}
}

// Reconcile records the changes made while the relay proxy was stopped, by comparing the
// flags currently loaded with the latest recorded state.
func (h *History) Reconcile(flagset string, flags map[string]flag.Flag) {
	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		if !h.waitReady() {
			return
		}
		entries, err := h.reconcileEntries(flagset, flags)
		if err != nil {
			h.logger.Error("flag history reconciliation failed", zap.String("flagset", flagset), zap.Error(err))
			return
		}
		for _, e := range entries {
			h.write(h.stopCtx, e)
		}
	}()
}

func (h *History) reconcileEntries(flagset string, flags map[string]flag.Flag) ([]Entry, error) {
	ctx, cancel := context.WithTimeout(h.stopCtx, defaultOperationTimeout)
	defer cancel()
	s, err := h.Store()
	if err != nil {
		return nil, err
	}
	latest, err := s.LatestPerFlag(ctx, flagset)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	base := baseEntry(flagset, h.cfg.Source, nil)
	base.DetectedAt = now
	entries := make([]Entry, 0)
	for key, f := range flags {
		m, err := toMap(f)
		if err != nil {
			return nil, err
		}
		hash, err := hashFlag(m)
		if err != nil {
			return nil, err
		}
		last, known := latest[key]
		if known && last.Action != ActionDeleted && last.AfterHash == hash {
			continue
		}
		e := base
		e.FlagKey, e.AfterHash = key, hash
		e.Diff = Diff{Reconciled: true, Summary: summarize(m)}
		e.Action = ActionAdded
		if known && last.Action != ActionDeleted {
			e.Action = ActionUpdated
			e.Diff.Summary = nil
		}
		entries = append(entries, e)
	}
	for key, last := range latest {
		if _, stillThere := flags[key]; stillThere || last.Action == ActionDeleted {
			continue
		}
		e := base
		e.FlagKey, e.Action = key, ActionDeleted
		e.Diff = Diff{Reconciled: true}
		entries = append(entries, e)
	}
	return entries, nil
}

// run connects to the storage then writes the queue until Close.
func (h *History) run() {
	defer h.wg.Done()
	if !h.connect() {
		return
	}
	for {
		select {
		case e := <-h.queue:
			if !h.write(h.stopCtx, e) {
				// interrupted by Close: the entry is the first one to flush.
				h.flush(&e)
				return
			}
		case <-h.stop:
			h.flush(nil)
			return
		}
	}
}

func (h *History) connect() bool {
	backoff := h.cfg.backoffMin
	for {
		ctx, cancel := context.WithTimeout(h.stopCtx, defaultOperationTimeout)
		s, err := h.cfg.Connect(ctx)
		if err == nil {
			err = s.Migrate(ctx)
			if err != nil {
				s.Close(ctx)
			}
		}
		cancel()
		if err == nil {
			h.mu.Lock()
			h.store = s
			h.mu.Unlock()
			close(h.ready)
			return true
		}
		h.logger.Error("flag history storage unavailable, retrying",
			zap.Error(err), zap.Duration("retryIn", backoff))
		if !h.sleep(backoff) {
			return false
		}
		backoff = min(backoff*2, h.cfg.backoffMax)
	}
}

// waitReady blocks until the storage is connected, false if the History is closed before.
func (h *History) waitReady() bool {
	select {
	case <-h.ready:
		return true
	case <-h.stopCtx.Done():
		return false
	case <-h.stop:
		return false
	}
}

// write records the entry, retrying with an exponential backoff until it succeeds
// or the context is canceled. It returns false when it has been interrupted.
func (h *History) write(ctx context.Context, e Entry) bool {
	backoff := h.cfg.backoffMin
	for {
		s, err := h.Store()
		if err == nil {
			opCtx, cancel := context.WithTimeout(ctx, defaultOperationTimeout)
			_, err = s.Record(opCtx, e)
			cancel()
		}
		if err == nil {
			return true
		}
		if ctx.Err() != nil {
			return false
		}
		h.logger.Error("impossible to record the flag history, retrying",
			zap.String("flagset", e.Flagset), zap.String("flag", e.FlagKey),
			zap.Error(err), zap.Duration("retryIn", backoff))
		if !h.sleepCtx(ctx, backoff) {
			return false
		}
		backoff = min(backoff*2, h.cfg.backoffMax)
	}
}

// flush writes what is still queued when the relay proxy stops, within a bounded time.
func (h *History) flush(interrupted *Entry) {
	ctx, cancel := context.WithTimeout(context.Background(), h.cfg.flushTimeout)
	defer cancel()
	if interrupted != nil && !h.write(ctx, *interrupted) {
		h.logger.Error("flag history entry lost while stopping",
			zap.String("flagset", interrupted.Flagset), zap.String("flag", interrupted.FlagKey))
		return
	}
	for {
		select {
		case e := <-h.queue:
			if !h.write(ctx, e) {
				h.logger.Error("flag history entries lost while stopping",
					zap.String("flagset", e.Flagset), zap.String("flag", e.FlagKey))
				return
			}
		default:
			return
		}
	}
}

func (h *History) purgeLoop() {
	defer h.wg.Done()
	if !h.waitReady() {
		return
	}
	ticker := time.NewTicker(h.cfg.purgeEvery)
	defer ticker.Stop()
	for {
		h.purge()
		select {
		case <-ticker.C:
		case <-h.stop:
			return
		}
	}
}

func (h *History) purge() {
	s, err := h.Store()
	if err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(h.stopCtx, defaultOperationTimeout)
	defer cancel()
	deleted, err := s.Purge(ctx, time.Now().Add(-h.cfg.Retention))
	if err != nil {
		h.logger.Error("flag history purge failed", zap.Error(err))
		return
	}
	if deleted > 0 {
		h.logger.Info("flag history purged", zap.Int64("deleted", deleted))
	}
}

func (h *History) sleep(d time.Duration) bool { return h.sleepCtx(h.stopCtx, d) }

func (h *History) sleepCtx(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return true
	case <-ctx.Done():
		return false
	}
}

// Close stops the background work and writes the queued changes (bounded by a timeout).
func (h *History) Close() {
	h.once.Do(func() {
		close(h.stop)
		// interrupts the retries in progress, the final flush has its own deadline.
		h.cancel()
		h.wg.Wait()
		h.mu.Lock()
		s := h.store
		h.store = nil
		h.mu.Unlock()
		if s != nil {
			s.Close(context.Background())
		}
	})
}

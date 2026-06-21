package postgresqlretriever

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PoolConfig holds the optional connection pool settings for the PostgreSQL
// retriever. Every field is optional: when a field is left at its zero value,
// the corresponding pgxpool default (or the value already encoded in the
// connection URI) is used, so existing configurations are unaffected.
type PoolConfig struct {
	// MaxOpenConns maps to pgxpool's MaxConns: the maximum number of
	// connections in the pool.
	MaxOpenConns int32
	// MaxIdleConns maps to pgxpool's MinConns: the minimum number of
	// connections the pool keeps open. pgxpool has no direct cap on idle
	// connections (idle connections are bounded by MaxConns), so this controls
	// the pool floor rather than an idle ceiling. It must not exceed
	// MaxOpenConns.
	MaxIdleConns int32
	// ConnMaxLifetime maps to pgxpool's MaxConnLifetime: the duration since
	// creation after which a connection is automatically closed.
	ConnMaxLifetime time.Duration
	// ConnMaxIdleTime maps to pgxpool's MaxConnIdleTime: the duration after
	// which an idle connection is automatically closed by the health check.
	ConnMaxIdleTime time.Duration
}

// IsZero reports whether no pool setting has been configured. When true,
// GetPool builds the pool exactly as it did before pool settings existed.
func (p PoolConfig) IsZero() bool {
	return p.MaxOpenConns == 0 &&
		p.MaxIdleConns == 0 &&
		p.ConnMaxLifetime == 0 &&
		p.ConnMaxIdleTime == 0
}

// cacheKey returns a stable cache key for a URI + pool config combination.
// Two retrievers that share a URI but request different pool settings get
// distinct cached pools instead of silently reusing the first one created.
func cacheKey(uri string, p PoolConfig) string {
	if p.IsZero() {
		return uri
	}
	return fmt.Sprintf("%s|%d|%d|%d|%d",
		uri, p.MaxOpenConns, p.MaxIdleConns, p.ConnMaxLifetime, p.ConnMaxIdleTime)
}

type poolEntry struct {
	pool     *pgxpool.Pool
	refCount int
}

var (
	mu      sync.Mutex
	poolMap = make(map[string]*poolEntry)
)

// GetPool returns a pool for a given URI, creating it if needed.
//
// The optional poolCfg controls the connection pool sizing and lifetimes. When
// it is the zero value, the pool is built exactly as before so existing
// behavior is unchanged.
func GetPool(ctx context.Context, uri string, poolCfg PoolConfig) (*pgxpool.Pool, error) {
	mu.Lock()
	defer mu.Unlock()

	key := cacheKey(uri, poolCfg)

	// If already exists, bump refCount
	if entry, ok := poolMap[key]; ok {
		entry.refCount++
		return entry.pool, nil
	}

	// Create a new pool
	pool, err := newPool(ctx, uri, poolCfg)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	poolMap[key] = &poolEntry{
		pool:     pool,
		refCount: 1,
	}

	return pool, nil
}

// newPool builds a pgxpool.Pool for the given URI. When poolCfg is the zero
// value it falls back to pgxpool.New so the resulting pool is identical to the
// previous behavior. Otherwise it parses the URI and overrides only the fields
// that were explicitly configured.
func newPool(ctx context.Context, uri string, poolCfg PoolConfig) (*pgxpool.Pool, error) {
	if poolCfg.IsZero() {
		return pgxpool.New(ctx, uri)
	}

	config, err := pgxpool.ParseConfig(uri)
	if err != nil {
		return nil, err
	}
	if poolCfg.MaxOpenConns > 0 {
		config.MaxConns = poolCfg.MaxOpenConns
	}
	if poolCfg.MaxIdleConns > 0 {
		config.MinConns = poolCfg.MaxIdleConns
	}
	if poolCfg.ConnMaxLifetime > 0 {
		config.MaxConnLifetime = poolCfg.ConnMaxLifetime
	}
	if poolCfg.ConnMaxIdleTime > 0 {
		config.MaxConnIdleTime = poolCfg.ConnMaxIdleTime
	}

	return pgxpool.NewWithConfig(ctx, config)
}

// ReleasePool decreases refCount and closes/removes when it hits zero.
//
// poolCfg must match the value passed to GetPool so the correct cached pool is
// released; the zero value targets the default (URI-only) pool.
func ReleasePool(_ context.Context, uri string, poolCfg PoolConfig) {
	mu.Lock()
	defer mu.Unlock()

	key := cacheKey(uri, poolCfg)

	entry, ok := poolMap[key]
	if !ok {
		return // nothing to do
	}

	entry.refCount--
	if entry.refCount <= 0 {
		entry.pool.Close()
		delete(poolMap, key)
	}
}

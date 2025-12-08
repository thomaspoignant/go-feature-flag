package postgresqlretriever

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

type poolEntry struct {
	pool     *pgxpool.Pool
	refCount int
}

var (
	mu      sync.Mutex
	poolMap = make(map[string]*poolEntry)
)

// GetPool returns a pool for a given URI, creating it if needed.
func GetPool(ctx context.Context, uri string) (*pgxpool.Pool, error) {
	mu.Lock()
	defer mu.Unlock()

	// If already exists, bump refCount
	if entry, ok := poolMap[uri]; ok {
		entry.refCount++
		return entry.pool, nil
	}

	// Create a new pool
	pool, err := pgxpool.New(ctx, uri)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	poolMap[uri] = &poolEntry{
		pool:     pool,
		refCount: 1,
	}

	return pool, nil
}

// ReleasePool decreases refCount and closes/removes when it hits zero.
func ReleasePool(uri string) {
	mu.Lock()
	defer mu.Unlock()

	entry, ok := poolMap[uri]
	if !ok {
		return // nothing to do
	}

	entry.refCount--
	if entry.refCount <= 0 {
		entry.pool.Close()
		delete(poolMap, uri)
	}
}

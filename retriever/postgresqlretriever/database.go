package postgresqlretriever

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
    pool     *pgxpool.Pool
    mu       sync.Mutex
    refCount int
)

func GetPool(ctx context.Context, uri string) (*pgxpool.Pool, error) {
    mu.Lock()
    defer mu.Unlock()

    if pool == nil {
        p, err := pgxpool.New(ctx, uri)
        if err != nil {
            return nil, err
        }
        if err := p.Ping(ctx); err != nil {
            p.Close()
            return nil, err
        }

        pool = p
    }

    refCount++
    return pool, nil
}

func ReleasePool() {
    mu.Lock()
    defer mu.Unlock()

    refCount--
    if refCount <= 0 {
        if pool != nil {
            pool.Close()
            pool = nil
        }
        refCount = 0
    }
}


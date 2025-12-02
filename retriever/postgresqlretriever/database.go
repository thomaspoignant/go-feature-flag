package postgresqlretriever

import (
    "context"
    "fmt"
    "sync"

    "github.com/jackc/pgx/v5/pgxpool"
)

var poolInstance *pgxpool.Pool
var once sync.Once
var errPool error

// refCount tracks how many retrievers are currently using the pool.
var refCount int
var refCountMutex sync.Mutex

func GetPool(ctx context.Context, uri string) (*pgxpool.Pool, error) {
	// The sync.Once ensures that the inner function is executed only once,
	// even if called by multiple goroutines concurrently.
	once.Do(func() {
		poolInstance, errPool = pgxpool.New(ctx, uri)
		if errPool != nil {
			errPool = fmt.Errorf("failed to create connection pool: %w", errPool)
			return
		}

		// Check connection immediately
		if err := poolInstance.Ping(ctx); err != nil {
			errPool = fmt.Errorf("failed to ping database with new pool: %w", err)
			// Don't close here, the pool remains valid for a retry connection
		}
	})

	if errPool != nil {
		return nil, errPool
	}

	refCountMutex.Lock()
	refCount++
	refCountMutex.Unlock()

	return poolInstance, nil
}

func ReleasePool() {
	refCountMutex.Lock()
	defer refCountMutex.Unlock()

	if refCount > 0 {
		refCount--
	}

	// Only close the physical connection when the last reference is released.
	if refCount == 0 && poolInstance != nil {
		poolInstance.Close()
		poolInstance = nil
		// Reset sync.Once to allow re-initialization if needed
		once = sync.Once{}
		errPool = nil
	}
}

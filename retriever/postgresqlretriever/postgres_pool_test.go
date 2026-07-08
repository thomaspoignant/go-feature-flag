package postgresqlretriever

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPoolConfig_IsZero(t *testing.T) {
	tests := []struct {
		name string
		cfg  PoolConfig
		want bool
	}{
		{
			name: "empty config is zero",
			cfg:  PoolConfig{},
			want: true,
		},
		{
			name: "max conns set",
			cfg:  PoolConfig{MaxConns: 10},
			want: false,
		},
		{
			name: "min conns set",
			cfg:  PoolConfig{MinConns: 2},
			want: false,
		},
		{
			name: "max conn lifetime set",
			cfg:  PoolConfig{MaxConnLifetime: time.Hour},
			want: false,
		},
		{
			name: "max conn idle time set",
			cfg:  PoolConfig{MaxConnIdleTime: 5 * time.Minute},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.cfg.IsZero())
		})
	}
}

// TestNewPool_AppliesConfig verifies that an explicit pool block is reflected in
// the resulting pgxpool configuration. Building a pool does not open a
// connection (that happens on Ping/Acquire), so this test does not need a live
// database.
func TestNewPool_AppliesConfig(t *testing.T) {
	ctx := context.Background()
	uri := "postgres://postgres:password@localhost:5432/postgres?sslmode=disable"

	cfg := PoolConfig{
		MaxConns:        25,
		MinConns:        5,
		MaxConnLifetime: 90 * time.Minute,
		MaxConnIdleTime: 10 * time.Minute,
	}

	pool, err := newPool(ctx, uri, cfg)
	assert.NoError(t, err)
	assert.NotNil(t, pool)
	defer pool.Close()

	applied := pool.Config()
	assert.Equal(t, int32(25), applied.MaxConns)
	assert.Equal(t, int32(5), applied.MinConns)
	assert.Equal(t, 90*time.Minute, applied.MaxConnLifetime)
	assert.Equal(t, 10*time.Minute, applied.MaxConnIdleTime)
}

func TestNewPool_InvalidMinConnsGreaterThanMaxConns(t *testing.T) {
	ctx := context.Background()
	uri := "postgres://postgres:password@localhost:5432/postgres?sslmode=disable"

	_, err := newPool(ctx, uri, PoolConfig{MaxConns: 5, MinConns: 10})
	assert.EqualError(t, err, "invalid pool configuration: minConns (10) must not exceed maxConns (5)")
}

// TestNewPool_PartialConfig verifies that only the explicitly set fields are
// overridden and the rest keep the pgxpool/URI defaults.
func TestNewPool_PartialConfig(t *testing.T) {
	ctx := context.Background()
	uri := "postgres://postgres:password@localhost:5432/postgres?sslmode=disable&pool_max_conns=8"

	pool, err := newPool(ctx, uri, PoolConfig{MaxConnLifetime: 2 * time.Hour})
	assert.NoError(t, err)
	assert.NotNil(t, pool)
	defer pool.Close()

	applied := pool.Config()
	// MaxConns keeps the value from the URI since it was not overridden.
	assert.Equal(t, int32(8), applied.MaxConns)
	assert.Equal(t, 2*time.Hour, applied.MaxConnLifetime)
}

// TestNewPool_InvalidURI surfaces parse errors when an explicit pool block is
// configured against an unparseable URI.
func TestNewPool_InvalidURI(t *testing.T) {
	ctx := context.Background()
	_, err := newPool(ctx, "://not-a-valid-uri", PoolConfig{MaxConns: 4})
	assert.Error(t, err)
}

// TestCacheKey ensures retrievers sharing a URI but with different pool settings
// resolve to distinct cache keys, while a zero config keeps the URI-only key.
func TestCacheKey(t *testing.T) {
	uri := "postgres://user:pass@localhost:5432/db"

	assert.Equal(t, uri, cacheKey(uri, PoolConfig{}),
		"zero config should key on the URI alone for backward compatibility")

	a := cacheKey(uri, PoolConfig{MaxConns: 10})
	b := cacheKey(uri, PoolConfig{MaxConns: 20})
	assert.NotEqual(t, a, b, "different settings must produce different keys")
	assert.NotEqual(t, uri, a, "a configured pool must not collide with the default key")

	// Same settings yield the same key (stable).
	assert.Equal(t,
		cacheKey(uri, PoolConfig{MaxConns: 10, MaxConnLifetime: time.Hour}),
		cacheKey(uri, PoolConfig{MaxConns: 10, MaxConnLifetime: time.Hour}))
}

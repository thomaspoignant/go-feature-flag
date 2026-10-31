package postgresqlretriever

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
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

// TestNewPool covers how a pool block is applied on top of the URI. Building a
// pool does not open a connection (that happens on Ping/Acquire), so these cases
// do not need a live database.
func TestNewPool(t *testing.T) {
	const uri = "postgres://postgres:password@localhost:5432/postgres?sslmode=disable"

	tests := []struct {
		name    string
		uri     string
		cfg     PoolConfig
		wantErr assert.ErrorAssertionFunc
		check   func(t *testing.T, applied *pgxpool.Config)
	}{
		{
			name: "explicit pool block is applied",
			uri:  uri,
			cfg: PoolConfig{
				MaxConns:        25,
				MinConns:        5,
				MaxConnLifetime: 90 * time.Minute,
				MaxConnIdleTime: 10 * time.Minute,
			},
			wantErr: assert.NoError,
			check: func(t *testing.T, applied *pgxpool.Config) {
				assert.Equal(t, int32(25), applied.MaxConns)
				assert.Equal(t, int32(5), applied.MinConns)
				assert.Equal(t, 90*time.Minute, applied.MaxConnLifetime)
				assert.Equal(t, 10*time.Minute, applied.MaxConnIdleTime)
			},
		},
		{
			name:    "partial config keeps the URI value for unset fields",
			uri:     uri + "&pool_max_conns=8",
			cfg:     PoolConfig{MaxConnLifetime: 2 * time.Hour},
			wantErr: assert.NoError,
			check: func(t *testing.T, applied *pgxpool.Config) {
				assert.Equal(t, int32(8), applied.MaxConns)
				assert.Equal(t, 2*time.Hour, applied.MaxConnLifetime)
			},
		},
		{
			name: "min conns greater than max conns is rejected",
			uri:  uri,
			cfg:  PoolConfig{MaxConns: 5, MinConns: 10},
			wantErr: func(t assert.TestingT, err error, msgAndArgs ...any) bool {
				return assert.EqualError(t, err,
					"invalid pool configuration: minConns (10) must not exceed maxConns (5)", msgAndArgs...)
			},
		},
		{
			name:    "unparseable URI surfaces the parse error",
			uri:     "://not-a-valid-uri",
			cfg:     PoolConfig{MaxConns: 4},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool, err := newPool(context.Background(), tt.uri, tt.cfg)
			if !tt.wantErr(t, err) || tt.check == nil {
				return
			}
			assert.NotNil(t, pool)
			defer pool.Close()
			tt.check(t, pool.Config())
		})
	}
}

// TestCacheKey ensures retrievers sharing a URI but with different pool settings
// resolve to distinct cache keys, while a zero config keeps the URI-only key.
func TestCacheKey(t *testing.T) {
	const uri = "postgres://user:pass@localhost:5432/db"

	tests := []struct {
		name      string
		a         string
		b         string
		wantEqual bool
	}{
		{
			name:      "zero config keys on the URI alone",
			a:         cacheKey(uri, PoolConfig{}),
			b:         uri,
			wantEqual: true,
		},
		{
			name:      "different settings produce different keys",
			a:         cacheKey(uri, PoolConfig{MaxConns: 10}),
			b:         cacheKey(uri, PoolConfig{MaxConns: 20}),
			wantEqual: false,
		},
		{
			name:      "configured pool does not collide with the default key",
			a:         cacheKey(uri, PoolConfig{MaxConns: 10}),
			b:         uri,
			wantEqual: false,
		},
		{
			name:      "same settings yield the same key",
			a:         cacheKey(uri, PoolConfig{MaxConns: 10, MaxConnLifetime: time.Hour}),
			b:         cacheKey(uri, PoolConfig{MaxConns: 10, MaxConnLifetime: time.Hour}),
			wantEqual: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantEqual {
				assert.Equal(t, tt.a, tt.b)
				return
			}
			assert.NotEqual(t, tt.a, tt.b)
		})
	}
}

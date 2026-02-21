package retrieverconf

import (
	"crypto/tls"
	"time"

	"github.com/redis/go-redis/v9"
)

// SerializableRedisOptions is a JSON-serializable alternative to redis.Options.
// It contains only the fields that can be serialized and are most commonly used.
// This allows the configuration to be passed via command line or JSON files.
type SerializableRedisOptions struct {
	// Network type, either tcp or unix. Default is tcp.
	Network string `mapstructure:"network" koanf:"network" json:"network,omitempty"`

	// Addr is the address formatted as host:port
	Addr string `mapstructure:"addr" koanf:"addr" json:"addr"`

	// ClientName will execute the `CLIENT SETNAME ClientName` command for each conn.
	ClientName string `mapstructure:"clientName" koanf:"clientname" json:"clientName,omitempty"`

	// Protocol 2 or 3. Use the version to negotiate RESP version with redis-server.
	// Default is 3.
	Protocol int `mapstructure:"protocol" koanf:"protocol" json:"protocol,omitempty"`

	// Username is used to authenticate the current connection
	Username string `mapstructure:"username" koanf:"username" json:"username,omitempty"`

	// Password is an optional password.
	Password string `mapstructure:"password" koanf:"password" json:"password,omitempty"` //nolint:gosec // G117

	// DB is the database to be selected after connecting to the server.
	DB int `mapstructure:"db" koanf:"db" json:"db,omitempty"`

	// MaxRetries is the maximum number of retries before giving up.
	// -1 (not 0) disables retries. Default is 3 retries.
	MaxRetries int `mapstructure:"maxRetries" koanf:"maxretries" json:"maxRetries,omitempty"`

	// MinRetryBackoff is the minimum backoff between each retry in milliseconds.
	// -1 disables backoff. Default is 8 milliseconds.
	MinRetryBackoffMs int64 `mapstructure:"minRetryBackoff" koanf:"minretrybackoff" json:"minRetryBackoff,omitempty"`

	// MaxRetryBackoff is the maximum backoff between each retry in milliseconds.
	// -1 disables backoff. Default is 512 milliseconds.
	MaxRetryBackoffMs int64 `mapstructure:"maxRetryBackoff" koanf:"maxretrybackoff" json:"maxRetryBackoff,omitempty"`

	// DialTimeout for establishing new connections in milliseconds.
	// Default is 5 seconds (5000ms).
	DialTimeoutMs int64 `mapstructure:"dialTimeout" koanf:"dialtimeout" json:"dialTimeout,omitempty"`

	// ReadTimeout for socket reads in milliseconds. If reached, commands will fail
	// with a timeout instead of blocking.
	// -1 = no timeout (block indefinitely)
	// -2 = disables SetReadDeadline calls completely
	// Default is 3 seconds (3000ms).
	ReadTimeoutMs int64 `mapstructure:"readTimeout" koanf:"readtimeout" json:"readTimeout,omitempty"`

	// WriteTimeout for socket writes in milliseconds. If reached, commands will fail
	// with a timeout instead of blocking.
	// Default is ReadTimeout.
	WriteTimeoutMs int64 `mapstructure:"writeTimeout" koanf:"writetimeout" json:"writeTimeout,omitempty"`

	// ContextTimeoutEnabled controls whether the client respects context timeouts and deadlines.
	// Default is false.
	ContextTimeoutEnabled bool `mapstructure:"contextTimeoutEnabled" koanf:"contexttimeoutenabled" json:"contextTimeoutEnabled,omitempty"` // nolint:lll

	// PoolFIFO uses FIFO mode for each node connection pool GET/PUT (default LIFO).
	PoolFIFO bool `mapstructure:"poolFIFO" koanf:"poolfifo" json:"poolFIFO,omitempty"`

	// PoolSize is the maximum number of socket connections.
	// Default is 10 connections per every available CPU.
	PoolSize int `mapstructure:"poolSize" koanf:"poolsize" json:"poolSize,omitempty"`

	// PoolTimeout is the amount of time client waits for connection if all
	// connections are busy before returning an error in milliseconds.
	// Default is ReadTimeout + 1 second.
	PoolTimeoutMs int64 `mapstructure:"poolTimeout" koanf:"pooltimeout" json:"poolTimeout,omitempty"`

	// MinIdleConns is the minimum number of idle connections which is useful when
	// establishing new connection is slow.
	// Default is 0 (no idle connections are retained).
	MinIdleConns int `mapstructure:"minIdleConns" koanf:"minidleconns" json:"minIdleConns,omitempty"`

	// MaxIdleConns is the maximum number of idle connections.
	// Default is 0 (unlimited).
	MaxIdleConns int `mapstructure:"maxIdleConns" koanf:"maxidleconns" json:"maxIdleConns,omitempty"`

	// ConnMaxIdleTime is the maximum amount of time a connection may be idle in milliseconds.
	// Should be less than server's timeout.
	// Default is 30 minutes (1800000ms). -1 disables idle timeout check.
	ConnMaxIdleTimeMs int64 `mapstructure:"connMaxIdleTime" koanf:"connmaxidletime" json:"connMaxIdleTime,omitempty"`

	// ConnMaxLifetime is the maximum amount of time a connection may be reused in milliseconds.
	// Default is to not close aged connections.
	ConnMaxLifetimeMs int64 `mapstructure:"connMaxLifetime" koanf:"connmaxlifetime" json:"connMaxLifetime,omitempty"`

	// TLS Config (simplified - only enable/disable for now)
	// For more complex TLS configurations, use the non-serializable RedisOptions
	TLSEnabled bool `mapstructure:"tlsEnabled" koanf:"tlsenabled" json:"tlsEnabled,omitempty"`

	// DisableIndentity disables set-lib on connect.
	DisableIndentity bool `mapstructure:"disableIndentity" koanf:"disableindentity" json:"disableIndentity,omitempty"`

	// IdentitySuffix is an optional suffix to append to the client name.
	IdentitySuffix string `mapstructure:"identitySuffix" koanf:"identitysuffix" json:"identitySuffix,omitempty"`
}

// ToRedisOptions converts SerializableRedisOptions to redis.Options
func (s *SerializableRedisOptions) ToRedisOptions() *redis.Options {
	opts := &redis.Options{
		Network:               s.Network,
		Addr:                  s.Addr,
		ClientName:            s.ClientName,
		Protocol:              s.Protocol,
		Username:              s.Username,
		Password:              s.Password,
		DB:                    s.DB,
		PoolFIFO:              s.PoolFIFO,
		PoolSize:              s.PoolSize,
		MinIdleConns:          s.MinIdleConns,
		MaxIdleConns:          s.MaxIdleConns,
		DisableIndentity:      s.DisableIndentity,
		IdentitySuffix:        s.IdentitySuffix,
		ContextTimeoutEnabled: s.ContextTimeoutEnabled,
	}

	// Convert milliseconds to time.Duration
	if s.MinRetryBackoffMs != 0 {
		opts.MinRetryBackoff = time.Duration(s.MinRetryBackoffMs) * time.Millisecond
	}
	if s.MaxRetryBackoffMs != 0 {
		opts.MaxRetryBackoff = time.Duration(s.MaxRetryBackoffMs) * time.Millisecond
	}
	if s.DialTimeoutMs != 0 {
		opts.DialTimeout = time.Duration(s.DialTimeoutMs) * time.Millisecond
	}
	if s.ReadTimeoutMs != 0 {
		opts.ReadTimeout = time.Duration(s.ReadTimeoutMs) * time.Millisecond
	}
	if s.WriteTimeoutMs != 0 {
		opts.WriteTimeout = time.Duration(s.WriteTimeoutMs) * time.Millisecond
	}
	if s.PoolTimeoutMs != 0 {
		opts.PoolTimeout = time.Duration(s.PoolTimeoutMs) * time.Millisecond
	}
	if s.ConnMaxIdleTimeMs != 0 {
		opts.ConnMaxIdleTime = time.Duration(s.ConnMaxIdleTimeMs) * time.Millisecond
	}
	if s.ConnMaxLifetimeMs != 0 {
		opts.ConnMaxLifetime = time.Duration(s.ConnMaxLifetimeMs) * time.Millisecond
	}

	if s.MaxRetries != 0 {
		opts.MaxRetries = s.MaxRetries
	}

	// TLS configuration (basic)
	if s.TLSEnabled {
		opts.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	return opts
}

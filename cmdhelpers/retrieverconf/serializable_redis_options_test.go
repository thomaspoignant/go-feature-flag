package retrieverconf_test

import (
	"crypto/tls"
	"encoding/json"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/retrieverconf"
)

func TestOptions_ToRedisOptions(t *testing.T) {
	tests := []struct {
		name     string
		input    *retrieverconf.Options
		validate func(t *testing.T, opts *redis.Options)
	}{
		{
			name: "basic configuration with addr and password",
			input: &retrieverconf.Options{
				Addr:     "localhost:6379",
				Password: "secret",
				DB:       1,
			},
			validate: func(t *testing.T, opts *redis.Options) {
				assert.Equal(t, "localhost:6379", opts.Addr)
				assert.Equal(t, "secret", opts.Password)
				assert.Equal(t, 1, opts.DB)
			},
		},
		{
			name: "full configuration",
			input: &retrieverconf.Options{
				Network:               "tcp",
				Addr:                  "redis.example.com:6380",
				ClientName:            "go-feature-flag",
				Protocol:              3,
				Username:              "admin",
				Password:              "password123",
				DB:                    2,
				MaxRetries:            5,
				MinRetryBackoffMs:     100,
				MaxRetryBackoffMs:     1000,
				DialTimeoutMs:         10000,
				ReadTimeoutMs:         5000,
				WriteTimeoutMs:        5000,
				PoolSize:              20,
				PoolTimeoutMs:         6000,
				MinIdleConns:          5,
				MaxIdleConns:          15,
				ConnMaxIdleTimeMs:     300000,
				ConnMaxLifetimeMs:     600000,
				PoolFIFO:              true,
				ContextTimeoutEnabled: true,
				DisableIndentity:      false,
				IdentitySuffix:        "test",
			},
			validate: func(t *testing.T, opts *redis.Options) {
				assert.Equal(t, "tcp", opts.Network)
				assert.Equal(t, "redis.example.com:6380", opts.Addr)
				assert.Equal(t, "go-feature-flag", opts.ClientName)
				assert.Equal(t, 3, opts.Protocol)
				assert.Equal(t, "admin", opts.Username)
				assert.Equal(t, "password123", opts.Password)
				assert.Equal(t, 2, opts.DB)
				assert.Equal(t, 5, opts.MaxRetries)
				assert.Equal(t, 100*time.Millisecond, opts.MinRetryBackoff)
				assert.Equal(t, 1000*time.Millisecond, opts.MaxRetryBackoff)
				assert.Equal(t, 10000*time.Millisecond, opts.DialTimeout)
				assert.Equal(t, 5000*time.Millisecond, opts.ReadTimeout)
				assert.Equal(t, 5000*time.Millisecond, opts.WriteTimeout)
				assert.Equal(t, 20, opts.PoolSize)
				assert.Equal(t, 6000*time.Millisecond, opts.PoolTimeout)
				assert.Equal(t, 5, opts.MinIdleConns)
				assert.Equal(t, 15, opts.MaxIdleConns)
				assert.Equal(t, 300000*time.Millisecond, opts.ConnMaxIdleTime)
				assert.Equal(t, 600000*time.Millisecond, opts.ConnMaxLifetime)
				assert.True(t, opts.PoolFIFO)
				assert.True(t, opts.ContextTimeoutEnabled)
				assert.False(t, opts.DisableIndentity)
				assert.Equal(t, "test", opts.IdentitySuffix)
			},
		},
		{
			name: "TLS enabled",
			input: &retrieverconf.Options{
				Addr:       "secure-redis.example.com:6379",
				TLSEnabled: true,
			},
			validate: func(t *testing.T, opts *redis.Options) {
				assert.Equal(t, "secure-redis.example.com:6379", opts.Addr)
				assert.NotNil(t, opts.TLSConfig)
				assert.Equal(t, uint16(tls.VersionTLS12), opts.TLSConfig.MinVersion)
			},
		},
		{
			name: "TLS disabled",
			input: &retrieverconf.Options{
				Addr:       "redis.example.com:6379",
				TLSEnabled: false,
			},
			validate: func(t *testing.T, opts *redis.Options) {
				assert.Equal(t, "redis.example.com:6379", opts.Addr)
				assert.Nil(t, opts.TLSConfig)
			},
		},
		{
			name: "zero values are not set (use redis defaults)",
			input: &retrieverconf.Options{
				Addr: "localhost:6379",
			},
			validate: func(t *testing.T, opts *redis.Options) {
				assert.Equal(t, "localhost:6379", opts.Addr)
				assert.Equal(t, time.Duration(0), opts.MinRetryBackoff)
				assert.Equal(t, time.Duration(0), opts.MaxRetryBackoff)
				assert.Equal(t, time.Duration(0), opts.DialTimeout)
				assert.Equal(t, 0, opts.MaxRetries)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.ToRedisOptions()
			require.NotNil(t, result)
			tt.validate(t, result)
		})
	}
}

func TestOptions_JSONSerialization(t *testing.T) {
	tests := []struct {
		name string
		opts *retrieverconf.Options
	}{
		{
			name: "basic options",
			opts: &retrieverconf.Options{
				Addr:     "localhost:6379",
				Password: "secret",
				DB:       1,
			},
		},
		{
			name: "full options",
			opts: &retrieverconf.Options{
				Network:               "tcp",
				Addr:                  "redis.example.com:6380",
				ClientName:            "test-client",
				Protocol:              3,
				Username:              "user",
				Password:              "pass",
				DB:                    2,
				MaxRetries:            5,
				MinRetryBackoffMs:     100,
				MaxRetryBackoffMs:     1000,
				DialTimeoutMs:         10000,
				ReadTimeoutMs:         5000,
				WriteTimeoutMs:        5000,
				PoolSize:              20,
				PoolTimeoutMs:         6000,
				MinIdleConns:          5,
				MaxIdleConns:          15,
				ConnMaxIdleTimeMs:     300000,
				ConnMaxLifetimeMs:     600000,
				TLSEnabled:            true,
				PoolFIFO:              true,
				ContextTimeoutEnabled: true,
				DisableIndentity:      false,
				IdentitySuffix:        "test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal to JSON
			jsonData, err := json.Marshal(tt.opts)
			require.NoError(t, err)
			require.NotEmpty(t, jsonData)

			// Unmarshal from JSON
			var decoded retrieverconf.Options
			err = json.Unmarshal(jsonData, &decoded)
			require.NoError(t, err)

			// Verify the data matches
			assert.Equal(t, tt.opts.Addr, decoded.Addr)
			assert.Equal(t, tt.opts.Password, decoded.Password)
			assert.Equal(t, tt.opts.DB, decoded.DB)
			assert.Equal(t, tt.opts.Username, decoded.Username)
			assert.Equal(t, tt.opts.Network, decoded.Network)
			assert.Equal(t, tt.opts.ClientName, decoded.ClientName)
			assert.Equal(t, tt.opts.Protocol, decoded.Protocol)
			assert.Equal(t, tt.opts.MaxRetries, decoded.MaxRetries)
			assert.Equal(t, tt.opts.TLSEnabled, decoded.TLSEnabled)
		})
	}
}

func TestOptions_JSONStructure(t *testing.T) {
	// This test ensures the JSON structure is as expected
	jsonStr := `{
		"addr": "localhost:6379",
		"password": "secret",
		"db": 1,
		"username": "admin",
		"maxRetries": 3,
		"dialTimeout": 5000,
		"readTimeout": 3000,
		"tlsEnabled": true
	}`

	var opts retrieverconf.Options
	err := json.Unmarshal([]byte(jsonStr), &opts)
	require.NoError(t, err)

	assert.Equal(t, "localhost:6379", opts.Addr)
	assert.Equal(t, "secret", opts.Password)
	assert.Equal(t, 1, opts.DB)
	assert.Equal(t, "admin", opts.Username)
	assert.Equal(t, 3, opts.MaxRetries)
	assert.Equal(t, int64(5000), opts.DialTimeoutMs)
	assert.Equal(t, int64(3000), opts.ReadTimeoutMs)
	assert.True(t, opts.TLSEnabled)
}

package config

import (
	"errors"
	"time"
)

// DefaultFlagHistoryRetention is how long the history is kept when nothing is configured.
const DefaultFlagHistoryRetention = 365 * 24 * time.Hour

// FlagHistory is the configuration of the persistent history of the flag modifications.
// It is stored in PostgreSQL and exposed by the GET /admin/v1/flag-history endpoint.
type FlagHistory struct {
	// Enabled (optional) activates the history. It is on by default, but it needs a database:
	// without a URI there is nothing to write to, so the history stays inactive.
	// Setting it explicitly to true without a URI is a configuration error.
	// Default: true
	Enabled *bool `mapstructure:"enabled" koanf:"enabled"`

	// URI is the PostgreSQL connection string (same format as the PostgreSQL retriever).
	// The history is active only when it is set.
	URI string `mapstructure:"uri" koanf:"uri"`

	// Table (optional) is the name of the table containing the history.
	// Default: flag_history
	Table string `mapstructure:"table" koanf:"table"`

	// Retention (optional) is the duration after which the history is deleted.
	// A negative value keeps the history forever.
	// Default: 8760h (1 year)
	Retention time.Duration `mapstructure:"retention" koanf:"retention"`

	// QueueSize (optional) is the number of modifications kept in memory while the database is
	// not reachable, the oldest are kept and the new ones are dropped when it is full.
	// Default: 1000
	QueueSize int `mapstructure:"queueSize" koanf:"queuesize"`
}

// RetentionDuration returns the retention to apply, 0 means the history is never purged.
func (f FlagHistory) RetentionDuration() time.Duration {
	switch {
	case f.Retention < 0:
		return 0
	case f.Retention == 0:
		return DefaultFlagHistoryRetention
	default:
		return f.Retention
	}
}

// IsEnabled returns true when the history must run: it is not disabled and a database is configured.
func (f FlagHistory) IsEnabled() bool {
	return (f.Enabled == nil || *f.Enabled) && f.URI != ""
}

// IsValid checks that the configuration is valid.
func (f FlagHistory) IsValid() error {
	if f.Enabled != nil && *f.Enabled && f.URI == "" {
		return errors.New("flagHistory.uri is required when flagHistory.enabled is true")
	}
	if f.IsEnabled() && f.QueueSize < 0 {
		return errors.New("flagHistory.queueSize must be positive")
	}
	return nil
}

package flag

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	// DefaultFlagTTL is the TTL applied to the flags when no other default is configured.
	DefaultFlagTTL = 90 * 24 * time.Hour

	hoursPerDay = 24
)

// ParseTTL parses a TTL such as "90d", "12h" or "1h30m".
// On top of the Go duration syntax it accepts a "d" suffix for days.
func ParseTTL(ttl string) (time.Duration, error) {
	ttl = strings.TrimSpace(ttl)
	if days, ok := strings.CutSuffix(ttl, "d"); ok {
		n, err := strconv.Atoi(days)
		if err != nil || n <= 0 {
			return 0, fmt.Errorf("invalid ttl %q", ttl)
		}
		return time.Duration(n) * hoursPerDay * time.Hour, nil
	}
	d, err := time.ParseDuration(ttl)
	if err != nil || d <= 0 {
		return 0, fmt.Errorf("invalid ttl %q", ttl)
	}
	return d, nil
}

// ExpirationDate returns the date at which the flag expires, in UTC.
// The priority is: ExpiresAt, then CreatedAt + TTL, then CreatedAt + defaultTTL.
// The second return value is false when the flag can't expire: no CreatedAt (nothing to compute from)
// or a non-positive defaultTTL with no flag specific configuration.
func (f *InternalFlag) ExpirationDate(defaultTTL time.Duration) (time.Time, bool) {
	if f.ExpiresAt != nil {
		return f.ExpiresAt.UTC(), true
	}
	if f.CreatedAt == nil {
		return time.Time{}, false
	}
	ttl := defaultTTL
	if f.TTL != nil {
		parsed, err := ParseTTL(*f.TTL)
		if err != nil {
			return time.Time{}, false
		}
		ttl = parsed
	}
	if ttl <= 0 {
		return time.Time{}, false
	}
	return f.CreatedAt.UTC().Add(ttl), true
}

// IsExpired returns true if the TTL of the flag is over at the given date.
func (f *InternalFlag) IsExpired(now time.Time, defaultTTL time.Duration) bool {
	_, expired := f.expirationIfExpired(now, defaultTTL)
	return expired
}

func (f *InternalFlag) expirationIfExpired(now time.Time, defaultTTL time.Duration) (time.Time, bool) {
	expiration, ok := f.ExpirationDate(defaultTTL)
	return expiration, ok && !now.Before(expiration)
}

func (f *InternalFlag) validateTTL() error {
	if f.TTL == nil {
		return nil
	}
	if _, err := ParseTTL(*f.TTL); err != nil {
		return err
	}
	return nil
}

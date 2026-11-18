package config

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
)

const (
	// DefaultValidateFlagsMaxJSONBodySize is the default maximum size of a JSON body (1 MiB).
	DefaultValidateFlagsMaxJSONBodySize ByteSize = 1 << 20
	// DefaultValidateFlagsMaxYAMLBodySize is the default maximum size of a YAML body (64 KiB).
	DefaultValidateFlagsMaxYAMLBodySize ByteSize = 64 << 10
	// DefaultValidateFlagsMaxTOMLBodySize is the default maximum size of a TOML body (64 KiB).
	DefaultValidateFlagsMaxTOMLBodySize ByteSize = 64 << 10
)

// ByteSize is a size in bytes.
// In the configuration it can be a plain integer (bytes) or a human-readable string like "1MiB" or "64KiB".
type ByteSize int64

// UnmarshalText implements encoding.TextUnmarshaler, it is used by the koanf decode hook to
// convert the strings (configuration file and environment variables) to a number of bytes.
func (b *ByteSize) UnmarshalText(text []byte) error {
	s := strings.TrimSpace(string(text))
	if s == "" {
		// an empty value (for example an empty environment variable) means unset.
		*b = 0
		return nil
	}
	if strings.HasPrefix(s, "-") {
		return fmt.Errorf("invalid size %q: must not be negative", s)
	}
	if n, err := strconv.ParseUint(s, 10, 63); err == nil {
		*b = ByteSize(n)
		return nil
	}
	if _, err := strconv.ParseFloat(s, 64); err == nil {
		// a number without unit that is not an integer (1.5, 1e3), do not truncate it silently.
		return fmt.Errorf("invalid size %q: a number of bytes must be an integer", s)
	}
	n, err := humanize.ParseBytes(s)
	if err != nil {
		return fmt.Errorf("invalid size %q: %w", s, err)
	}
	if n > math.MaxInt64 {
		return fmt.Errorf("invalid size %q: too large", s)
	}
	*b = ByteSize(n)
	return nil
}

// Endpoints is the configuration of the endpoints of the relay proxy.
type Endpoints struct {
	// ValidateFlags configures the POST /admin/v1/flags/validate endpoint.
	ValidateFlags ValidateFlagsEndpoint `mapstructure:"validateFlags" koanf:"validateflags"`
}

// ValidateFlagsEndpoint configures the body size limits of the validate flags endpoint.
// A value of 0 means that the default is used.
type ValidateFlagsEndpoint struct {
	// MaxJSONBodySize is the maximum size of a JSON body.
	// Default: 1 MiB
	MaxJSONBodySize ByteSize `mapstructure:"maxJSONBodySize" koanf:"maxjsonbodysize"`
	// MaxYAMLBodySize is the maximum size of a YAML body.
	// Default: 64 KiB
	MaxYAMLBodySize ByteSize `mapstructure:"maxYAMLBodySize" koanf:"maxyamlbodysize"`
	// MaxTOMLBodySize is the maximum size of a TOML body.
	// Default: 64 KiB
	MaxTOMLBodySize ByteSize `mapstructure:"maxTOMLBodySize" koanf:"maxtomlbodysize"`
}

// WithDefaults returns a copy where the unset (zero) limits are replaced by their default.
func (v ValidateFlagsEndpoint) WithDefaults() ValidateFlagsEndpoint {
	if v.MaxJSONBodySize == 0 {
		v.MaxJSONBodySize = DefaultValidateFlagsMaxJSONBodySize
	}
	if v.MaxYAMLBodySize == 0 {
		v.MaxYAMLBodySize = DefaultValidateFlagsMaxYAMLBodySize
	}
	if v.MaxTOMLBodySize == 0 {
		v.MaxTOMLBodySize = DefaultValidateFlagsMaxTOMLBodySize
	}
	return v
}

// IsValid validates the configuration of the endpoints.
func (e Endpoints) IsValid() error {
	return e.ValidateFlags.IsValid()
}

// IsValid validates the limits, negative values are rejected.
func (v ValidateFlagsEndpoint) IsValid() error {
	for _, limit := range []struct {
		name string
		size ByteSize
	}{
		{"endpoints.validateFlags.maxJSONBodySize", v.MaxJSONBodySize},
		{"endpoints.validateFlags.maxYAMLBodySize", v.MaxYAMLBodySize},
		{"endpoints.validateFlags.maxTOMLBodySize", v.MaxTOMLBodySize},
	} {
		if limit.size < 0 {
			return fmt.Errorf("%s must not be negative", limit.name)
		}
	}
	return nil
}

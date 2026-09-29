package config

import (
	"fmt"
	"net/url"
)

// OfrepEventStream holds the configuration used to advertise the flag-change SSE
// endpoint in the OFREP bulk evaluation response (OpenFeature ADR-0008).
type OfrepEventStream struct {
	// BaseURL (optional) is the public base URL clients should use to reach the relay proxy,
	// for example https://gofeatureflag.example.com. The path of the SSE endpoint is appended
	// automatically. When empty, the eventStreams field is not added to the response.
	BaseURL string `mapstructure:"baseUrl" koanf:"baseurl"`

	// InactivityDelaySec (optional) is advertised to the client as the delay in seconds after
	// which it should consider the stream inactive. When 0 the field is omitted.
	InactivityDelaySec int `mapstructure:"inactivityDelaySec" koanf:"inactivitydelaysec"`
}

// IsEnabled returns true if the eventStreams field should be added to the OFREP bulk
// evaluation response.
func (o OfrepEventStream) IsEnabled() bool {
	return o.BaseURL != "" && o.IsValid() == nil
}

// IsValid checks that the configuration is valid.
func (o OfrepEventStream) IsValid() error {
	if o.InactivityDelaySec < 0 {
		return fmt.Errorf("ofrepEventStream.inactivityDelaySec must be positive, got %d", o.InactivityDelaySec)
	}
	if o.BaseURL == "" {
		return nil
	}
	u, err := url.Parse(o.BaseURL)
	if err != nil {
		return fmt.Errorf("ofrepEventStream.baseUrl is not a valid URL: %w", err)
	}
	if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return fmt.Errorf("ofrepEventStream.baseUrl must be an absolute http(s) URL, got %q", o.BaseURL)
	}
	if u.RawQuery != "" || u.Fragment != "" {
		return fmt.Errorf("ofrepEventStream.baseUrl must not contain a query or a fragment, got %q", o.BaseURL)
	}
	return nil
}

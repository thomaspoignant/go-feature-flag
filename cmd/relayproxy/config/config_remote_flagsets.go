package config

import (
	"strings"
	"time"

	"github.com/knadh/koanf/v2"
	httpprovider "github.com/thomaspoignant/koanf-http-provider"
	"go.uber.org/zap"
)

// RemoteFlagSets holds the configuration to load the flagSets from an HTTP(S) URL.
// When URL is set, the flagSets only come from this URL, the rest of the configuration
// still comes from the configuration file and the environment variables.
type RemoteFlagSets struct {
	// URL of a document with a flagSets key, in the same format as the configuration file.
	URL string `mapstructure:"url" koanf:"url"`

	// Headers (optional) are sent with every request, for example an Authorization header.
	Headers map[string][]string `mapstructure:"headers" koanf:"headers"`

	// PollingInterval (optional) is the time in milliseconds between two checks of the URL.
	// Default: 60000
	PollingInterval int `mapstructure:"pollingInterval" koanf:"pollinginterval"`

	// CacheFile (optional) keeps the last valid document, it is used when the URL is unreachable.
	CacheFile string `mapstructure:"cacheFile" koanf:"cachefile"`
}

// parser returns the parser matching the extension of the URL path.
func (r RemoteFlagSets) parser() koanf.Parser {
	path, _, _ := strings.Cut(r.URL, "?") // pre-signed URLs have a query string after the extension
	return selectParserForFile(path)
}

// provider returns a koanf provider polling the URL.
func (r RemoteFlagSets) provider(parser koanf.Parser, log *zap.Logger) *httpprovider.HTTP {
	return httpprovider.Provider(r.URL,
		httpprovider.WithHeaders(r.Headers),
		httpprovider.WithPollInterval(time.Duration(r.PollingInterval)*time.Millisecond),
		httpprovider.WithCache(r.CacheFile),
		// invalid content is never cached and never triggers a reload
		httpprovider.WithValidate(func(content []byte) error {
			_, err := parser.Unmarshal(content)
			return err
		}),
		httpprovider.WithOnFallback(func(fetchErr error, cachedAt time.Time) {
			log.Warn("remote flagSets unreachable, using the cached version",
				zap.Error(fetchErr), zap.Time("cachedAt", cachedAt))
		}),
	)
}

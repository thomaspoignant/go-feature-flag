package config

import (
	"strings"
	"time"

	"github.com/knadh/koanf/v2"
	httpprovider "github.com/thomaspoignant/koanf-http-provider"
	"go.uber.org/zap"
)

// Studio holds the configuration to load the flagSets from GO Feature Flag Studio.
// When URL is set, the flagSets only come from Studio, the rest of the configuration
// still comes from the configuration file and the environment variables.
// Changes to the Studio configuration require restarting the relay proxy.
type Studio struct {
	// URL of the relay proxy configuration served by Studio, only its flagSets key is used.
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
func (s Studio) parser() koanf.Parser {
	path, _, _ := strings.Cut(s.URL, "?") // pre-signed URLs have a query string after the extension
	return selectParserForFile(path)
}

// provider returns a koanf provider polling the URL.
func (s Studio) provider(parser koanf.Parser, log *zap.Logger) *httpprovider.HTTP {
	return httpprovider.Provider(s.URL,
		httpprovider.WithHeaders(s.Headers),
		httpprovider.WithPollInterval(time.Duration(s.PollingInterval)*time.Millisecond),
		httpprovider.WithCache(s.CacheFile),
		// invalid content is never cached and never triggers a reload
		httpprovider.WithValidate(func(content []byte) error {
			_, err := parser.Unmarshal(content)
			return err
		}),
		httpprovider.WithOnFallback(func(fetchErr error, cachedAt time.Time) {
			log.Warn("studio unreachable, using the cached flagSets",
				zap.Error(fetchErr), zap.Time("cachedAt", cachedAt))
		}),
	)
}

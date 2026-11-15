package config_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/testutils"
	"go.uber.org/zap"
)

const studioYaml = `server:
  port: 9999
flagSets:
  - name: team-a
    apiKeys:
      - key-a
    retrievers:
      - kind: file
        path: /tmp/flags.yaml`

// flagSetNames returns the names of the flagSets of a configuration.
func flagSetNames(c *config.Config) []string {
	names := make([]string, 0, len(c.FlagSets))
	for _, flagSet := range c.FlagSets {
		names = append(names, flagSet.Name)
	}
	return names
}

// newStudioServer serves body when the Authorization header matches, it is down when body is empty.
func newStudioServer(t *testing.T, body, authHeader string) *httptest.Server {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != authHeader {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	if body == "" {
		srv.Close()
	}
	return srv
}

// writeWithETag answers with status when it is not 200, with 304 when the client already has etag, else with body.
func writeWithETag(w http.ResponseWriter, r *http.Request, body, etag string, status int) {
	w.Header().Set("ETag", etag)
	if status != http.StatusOK {
		w.WriteHeader(status)
		return
	}
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	_, _ = w.Write([]byte(body))
}

// receive returns the next value of ch, it fails the test with msg after 5 seconds.
func receive[T any](t *testing.T, ch <-chan T, msg string) T {
	t.Helper()
	select {
	case v := <-ch:
		return v
	case <-time.After(5 * time.Second):
		t.Fatal(msg)
		var zero T
		return zero
	}
}

func TestStudio(t *testing.T) {
	tests := []struct {
		name         string
		fileConfig   string            // {{URL}} is the URL of the server, {{CACHE}} the cache file
		body         string            // content served by the server, the server is down if empty
		authHeader   string            // Authorization header required by the server
		cache        string            // content of the cache file
		env          map[string]string // environment variables
		wantFlagSets []string
		wantInvalid  bool // flagSets and studio are both set
	}{
		{
			name: "flagSets from the URL, the rest from the file",
			fileConfig: `server:
  port: 2031
studio:
  url: {{URL}}/flagsets.yaml`,
			body:         studioYaml,
			wantFlagSets: []string{"team-a"},
		},
		{
			name: "flagSets in the file and studio",
			fileConfig: `server:
  port: 2031
flagSets:
  - name: team-file
    apiKeys:
      - key-file
studio:
  url: {{URL}}/flagsets.yaml`,
			body:         studioYaml,
			wantFlagSets: []string{"team-file"},
			wantInvalid:  true,
		},
		{
			name: "toml URL with a query string",
			fileConfig: `server:
  port: 2031
studio:
  url: {{URL}}/flagsets.toml?sig=abc`,
			body: `[[flagSets]]
name = "team-a"
apiKeys = ["key-a"]
[[flagSets.retrievers]]
kind = "file"
path = "/tmp/flags.yaml"`,
			wantFlagSets: []string{"team-a"},
		},
		{
			name: "headers from the file",
			fileConfig: `server:
  port: 2031
studio:
  url: {{URL}}/flagsets.yaml
  headers:
    Authorization: Bearer secret`,
			body:         studioYaml,
			authHeader:   "Bearer secret",
			wantFlagSets: []string{"team-a"},
		},
		{
			name: "headers from an environment variable, URL from the file",
			fileConfig: `server:
  port: 2031
studio:
  url: {{URL}}/flagsets.yaml`,
			body:         studioYaml,
			authHeader:   "Bearer secret",
			env:          map[string]string{"STUDIO_HEADERS_AUTHORIZATION": "Bearer secret"},
			wantFlagSets: []string{"team-a"},
		},
		{
			name: "authorization header missing",
			fileConfig: `server:
  port: 2031
studio:
  url: {{URL}}/flagsets.yaml`,
			body:         studioYaml,
			authHeader:   "Bearer secret",
			wantFlagSets: []string{},
		},
		{
			name: "server down, flagSets from the cache",
			fileConfig: `server:
  port: 2031
studio:
  url: {{URL}}/flagsets.yaml
  cacheFile: {{CACHE}}`,
			cache:        studioYaml,
			wantFlagSets: []string{"team-a"},
		},
		{
			name: "flagSets in the file and studio down",
			fileConfig: `server:
  port: 2031
flagSets:
  - name: team-file
    apiKeys:
      - key-file
studio:
  url: {{URL}}/flagsets.yaml`,
			wantFlagSets: []string{"team-file"},
			wantInvalid:  true,
		},
		{
			name: "no studio, flagSets from the file",
			fileConfig: `server:
  port: 2031
flagSets:
  - name: team-file
    apiKeys:
      - key-file`,
			wantFlagSets: []string{"team-file"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := newStudioServer(t, tt.body, tt.authHeader)
			cacheFile := filepath.Join(t.TempDir(), "flagsets.cache.yaml")
			if tt.cache != "" {
				require.NoError(t, os.WriteFile(cacheFile, []byte(tt.cache), 0600))
			}
			for name, value := range tt.env {
				t.Setenv(name, value)
			}
			fileConfig := strings.NewReplacer("{{URL}}", srv.URL, "{{CACHE}}", cacheFile).Replace(tt.fileConfig)
			configFile := testutils.CopyContentToNewTempFile(t, fileConfig)
			f := pflag.NewFlagSet("config", pflag.ContinueOnError)
			f.String("config", "", "Location of your config file")
			require.NoError(t, f.Parse([]string{"--config=" + configFile.Name()}))

			got, err := config.NewConfigLoader(f, zap.NewNop(), "1.0.0", false).ToConfig()
			require.NoError(t, err)
			assert.Equal(t, 2031, got.Server.Port)
			assert.Equal(t, tt.wantFlagSets, flagSetNames(got))
			if tt.wantInvalid {
				assert.ErrorContains(t, got.IsValid(), "flagSets and studio cannot be used together")
			}
		})
	}
}

func TestStudioReload(t *testing.T) {
	const fileConfig = `server:
  port: 2031
studio:
  url: {{URL}}/flagsets.yaml
  pollingInterval: 50`
	const twoFlagSets = studioYaml + `
  - name: team-b
    apiKeys:
      - key-b
    retrievers:
      - kind: file
        path: /tmp/flags.yaml`

	tests := []struct {
		name         string
		newBody      string // new content served by the server
		newFile      string // new content of the configuration file
		wantPort     int
		wantFlagSets []string
	}{
		{
			name:         "the URL changes",
			newBody:      twoFlagSets,
			wantPort:     2031,
			wantFlagSets: []string{"team-a", "team-b"},
		},
		{
			name:         "the file changes, flagSets still from the URL",
			newFile:      strings.Replace(fileConfig, "2031", "3031", 1),
			wantPort:     3031,
			wantFlagSets: []string{"team-a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mu sync.Mutex
			body := studioYaml
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				mu.Lock()
				defer mu.Unlock()
				_, _ = w.Write([]byte(body))
			}))
			defer srv.Close()
			configFile := testutils.CopyContentToNewTempFile(t, strings.ReplaceAll(fileConfig, "{{URL}}", srv.URL))
			f := pflag.NewFlagSet("config", pflag.ContinueOnError)
			f.String("config", "", "Location of your config file")
			require.NoError(t, f.Parse([]string{"--config=" + configFile.Name()}))

			loader := config.NewConfigLoader(f, zap.NewNop(), "1.0.0", true)
			defer func() { _ = loader.StopWatchChanges() }()
			initialConfig, err := loader.ToConfig()
			require.NoError(t, err)
			assert.Equal(t, []string{"team-a"}, flagSetNames(initialConfig))

			configs := make(chan *config.Config, 10)
			loader.AddConfigChangeCallback(func(newConfig *config.Config) { configs <- newConfig })

			if tt.newBody != "" {
				mu.Lock()
				body = tt.newBody
				mu.Unlock()
			}
			if tt.newFile != "" {
				_ = testutils.CopyContentToExistingTempFile(t, strings.ReplaceAll(tt.newFile, "{{URL}}", srv.URL), configFile)
			}

			newConfig := receive(t, configs, "Timeout waiting for callback to be triggered")
			assert.Equal(t, tt.wantPort, newConfig.Server.Port)
			assert.Equal(t, tt.wantFlagSets, flagSetNames(newConfig))
		})
	}
}

func TestStudioReloadRetriesFailedFetch(t *testing.T) {
	for _, etag := range []string{"", `"v2"`} {
		t.Run("etag="+etag, func(t *testing.T) {
			var mu sync.Mutex
			requests := 0
			retries := make(chan string, 1)
			updated := strings.ReplaceAll(studioYaml, "team-a", "team-b")
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mu.Lock()
				defer mu.Unlock()
				requests++
				switch requests {
				case 1:
					_, _ = w.Write([]byte(studioYaml))
				case 2:
					w.Header().Set("ETag", etag)
					_, _ = w.Write([]byte(updated))
				case 3:
					w.WriteHeader(http.StatusServiceUnavailable)
				default:
					select {
					case retries <- r.Header.Get("If-None-Match"):
					default:
					}
					w.Header().Set("ETag", etag)
					if etag != "" && r.Header.Get("If-None-Match") == etag {
						w.WriteHeader(http.StatusNotModified)
						return
					}
					_, _ = w.Write([]byte(updated))
				}
			}))
			defer srv.Close()
			loader := newStudioTestLoader(t, srv.URL, "")
			type observedConfig struct {
				config   *config.Config
				requests int
			}
			configs := make(chan observedConfig, 10)
			loader.AddConfigChangeCallback(func(c *config.Config) {
				mu.Lock()
				count := requests
				mu.Unlock()
				configs <- observedConfig{config: c, requests: count}
			})

			c := receive(t, configs, "Timeout waiting for recovery after the failed reload")
			assert.Equal(t, []string{"team-b"}, flagSetNames(c.config))
			assert.NoError(t, c.config.IsValid())
			assert.GreaterOrEqual(t, c.requests, 4, "failed reloads must not notify consumers")
			header := receive(t, retries, "Timeout waiting for an unconditional retry")
			assert.Empty(t, header, "the failed reload must invalidate the accepted ETag")
		})
	}
}

func TestStudioRejectedUpdateKeepsETag(t *testing.T) {
	tests := []struct {
		name   string
		body   string
		status int
	}{
		{name: "HTTP error", status: http.StatusServiceUnavailable},
		{name: "invalid YAML", body: "flagSets: [", status: http.StatusOK},
		{name: "invalid field type", body: "flagSets: wrong-type", status: http.StatusOK},
		{
			name:   "invalid flagset",
			body:   strings.Replace(studioYaml, "apiKeys:\n      - key-a", "apiKeys: []", 1),
			status: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mu sync.Mutex
			body, etag, status := studioYaml, `"v1"`, http.StatusOK
			rejected := false
			attempts := make(chan string, 20)
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mu.Lock()
				defer mu.Unlock()
				if rejected {
					select {
					case attempts <- r.Header.Get("If-None-Match"):
					default:
					}
				}
				writeWithETag(w, r, body, etag, status)
			}))
			defer srv.Close()
			cacheFile := filepath.Join(t.TempDir(), "flagsets.cache.yaml")
			loader := newStudioTestLoader(t, srv.URL, cacheFile)
			configs := make(chan *config.Config, 10)
			loader.AddConfigChangeCallback(func(c *config.Config) { configs <- c })
			mu.Lock()
			body, etag, status, rejected = tt.body, `"v2"`, tt.status, true
			mu.Unlock()

			// Failed updates must keep requesting the last successfully accepted ETag.
			for range 2 {
				require.Equal(t, `"v1"`, receive(t, attempts, "Timeout waiting for a retry of the rejected update"))
			}
			cached, err := os.ReadFile(cacheFile)
			require.NoError(t, err)
			assert.Equal(t, studioYaml, string(cached))
			mu.Lock()
			body = strings.ReplaceAll(studioYaml, "team-a", "team-b")
			status, rejected = http.StatusOK, false
			mu.Unlock()

			// Recovery must work even when the server reuses the rejected ETag.
			c := receive(t, configs, "Timeout waiting for recovery with the same ETag")
			assert.Equal(t, []string{"team-b"}, flagSetNames(c))
			assert.NoError(t, c.IsValid())
		})
	}
}

func newStudioTestLoader(t *testing.T, url, cacheFile string) *config.ConfigLoader {
	t.Helper()
	fileConfig := "studio:\n  url: " + url + "/flagsets.yaml\n  pollingInterval: 150\n"
	if cacheFile != "" {
		fileConfig += "  cacheFile: " + cacheFile + "\n"
	}
	configFile := testutils.CopyContentToNewTempFile(t, fileConfig)
	f := pflag.NewFlagSet("config", pflag.ContinueOnError)
	f.String("config", "", "Location of your config file")
	require.NoError(t, f.Parse([]string{"--config=" + configFile.Name()}))
	loader := config.NewConfigLoader(f, zap.NewNop(), "1.0.0", true)
	t.Cleanup(func() { assert.NoError(t, loader.StopWatchChanges()) })
	return loader
}

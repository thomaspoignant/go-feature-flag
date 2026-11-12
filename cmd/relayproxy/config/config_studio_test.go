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
      - key-a`

// flagSetNames returns the names of the flagSets of a configuration.
func flagSetNames(c *config.Config) []string {
	names := make([]string, 0, len(c.FlagSets))
	for _, flagSet := range c.FlagSets {
		names = append(names, flagSet.Name)
	}
	return names
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
apiKeys = ["key-a"]`,
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
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Authorization") != tt.authHeader {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()
			if tt.body == "" {
				srv.Close()
			}
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
      - key-b`

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

			select {
			case newConfig := <-configs:
				assert.Equal(t, tt.wantPort, newConfig.Server.Port)
				assert.Equal(t, tt.wantFlagSets, flagSetNames(newConfig))
			case <-time.After(5 * time.Second):
				t.Fatal("Timeout waiting for callback to be triggered")
			}
		})
	}
}

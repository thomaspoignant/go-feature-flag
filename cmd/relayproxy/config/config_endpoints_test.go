package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"go.uber.org/zap"
)

func TestByteSize_UnmarshalText(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		text    string
		want    config.ByteSize
		wantErr bool
	}{
		{name: "plain bytes", text: "1048576", want: 1048576},
		{name: "negative bytes", text: "-5", want: -5},
		{name: "MiB", text: "1MiB", want: 1 << 20},
		{name: "KiB with space", text: "64 KiB", want: 64 << 10},
		{name: "decimal unit", text: "2KB", want: 2000},
		{name: "invalid", text: "lots", wantErr: true},
		{name: "empty", text: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var got config.ByteSize
			err := got.UnmarshalText([]byte(tt.text))
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestValidateFlagsEndpoint_WithDefaults(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   config.ValidateFlagsEndpoint
		want config.ValidateFlagsEndpoint
	}{
		{
			name: "unset",
			want: config.ValidateFlagsEndpoint{MaxJSONBodySize: 1 << 20, MaxYAMLBodySize: 64 << 10, MaxTOMLBodySize: 64 << 10},
		},
		{
			name: "partially set",
			in:   config.ValidateFlagsEndpoint{MaxYAMLBodySize: 1 << 20},
			want: config.ValidateFlagsEndpoint{MaxJSONBodySize: 1 << 20, MaxYAMLBodySize: 1 << 20, MaxTOMLBodySize: 64 << 10},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.in.WithDefaults())
		})
	}
}

func TestEndpoints_IsValid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		in      config.Endpoints
		wantErr bool
	}{
		{name: "unset"},
		{name: "valid", in: config.Endpoints{ValidateFlags: config.ValidateFlagsEndpoint{MaxJSONBodySize: 10}}},
		{name: "negative json", in: config.Endpoints{ValidateFlags: config.ValidateFlagsEndpoint{MaxJSONBodySize: -1}}, wantErr: true},
		{name: "negative yaml", in: config.Endpoints{ValidateFlags: config.ValidateFlagsEndpoint{MaxYAMLBodySize: -1}}, wantErr: true},
		{name: "negative toml", in: config.Endpoints{ValidateFlags: config.ValidateFlagsEndpoint{MaxTOMLBodySize: -1}}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.in.IsValid()
			if tt.wantErr {
				assert.ErrorContains(t, err, "must not be negative")
				return
			}
			assert.NoError(t, err)
		})
	}
}

// TestEndpoints_Loading checks the loading of the endpoints block from the file and the environment.
// It cannot be parallel because of t.Setenv.
func TestEndpoints_Loading(t *testing.T) {
	const base = "retriever:\n  kind: file\n  path: ./flags.yaml\n"
	tests := []struct {
		name        string
		file        string
		env         map[string]string
		want        config.ValidateFlagsEndpoint
		wantInvalid bool
	}{
		{name: "unset keeps zero values (defaults are applied by the handler)", file: base},
		{
			name: "integers in the file",
			file: base + "endpoints:\n  validateFlags:\n    maxJSONBodySize: 2048\n    maxYAMLBodySize: 4096\n    maxTOMLBodySize: 8192\n",
			want: config.ValidateFlagsEndpoint{MaxJSONBodySize: 2048, MaxYAMLBodySize: 4096, MaxTOMLBodySize: 8192},
		},
		{
			name: "strings in the file",
			file: base + "endpoints:\n  validateFlags:\n    maxJSONBodySize: 2MiB\n    maxYAMLBodySize: 128KiB\n    maxTOMLBodySize: \"256 KiB\"\n",
			want: config.ValidateFlagsEndpoint{MaxJSONBodySize: 2 << 20, MaxYAMLBodySize: 128 << 10, MaxTOMLBodySize: 256 << 10},
		},
		{
			name: "environment variables",
			file: base,
			env: map[string]string{
				"ENDPOINTS_VALIDATEFLAGS_MAXJSONBODYSIZE": "3MiB",
				"ENDPOINTS_VALIDATEFLAGS_MAXYAMLBODYSIZE": "100000",
				"ENDPOINTS_VALIDATEFLAGS_MAXTOMLBODYSIZE": "32KiB",
			},
			want: config.ValidateFlagsEndpoint{MaxJSONBodySize: 3 << 20, MaxYAMLBodySize: 100000, MaxTOMLBodySize: 32 << 10},
		},
		{
			name: "environment overrides the file",
			file: base + "endpoints:\n  validateFlags:\n    maxJSONBodySize: 2048\n",
			env:  map[string]string{"ENDPOINTS_VALIDATEFLAGS_MAXJSONBODYSIZE": "1KiB"},
			want: config.ValidateFlagsEndpoint{MaxJSONBodySize: 1 << 10},
		},
		{
			name:        "negative value is rejected",
			file:        base + "endpoints:\n  validateFlags:\n    maxYAMLBodySize: -1\n",
			want:        config.ValidateFlagsEndpoint{MaxYAMLBodySize: -1},
			wantInvalid: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "goff-proxy.yaml")
			require.NoError(t, os.WriteFile(path, []byte(tt.file), 0o600))
			for k, v := range tt.env {
				t.Setenv(k, v)
			}
			fs := pflag.NewFlagSet("config", pflag.ContinueOnError)
			fs.String("config", "", "Location of your config file")
			require.NoError(t, fs.Parse([]string{"--config=" + path}))
			loader := config.NewConfigLoader(fs, zap.NewNop(), "1.0.0", false)
			got, err := loader.ToConfig()
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.Endpoints.ValidateFlags)
			if tt.wantInvalid {
				assert.Error(t, got.IsValid())
			} else {
				assert.NoError(t, got.IsValid())
			}
		})
	}
}

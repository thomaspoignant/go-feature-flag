package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/retrieverconf"
)

func TestConfig_UnixSocket_Validation(t *testing.T) {
	tests := []struct {
		name       string
		config     *config.Config
		wantErr    bool
		errMessage string
	}{
		{
			name: "Valid config with Unix socket only",
			config: &config.Config{
				CommonFlagSet: config.CommonFlagSet{
					Retriever: &retrieverconf.RetrieverConf{
						Kind: "file",
						Path: "../testdata/config/valid-file.yaml",
					},
				},
				UnixSocket: "/var/run/goff/goff.sock",
			},
			wantErr: false,
		},
		{
			name: "Valid config with TCP port only",
			config: &config.Config{
				CommonFlagSet: config.CommonFlagSet{
					Retriever: &retrieverconf.RetrieverConf{
						Kind: "file",
						Path: "../testdata/config/valid-file.yaml",
					},
				},
				ListenPort: 8080,
			},
			wantErr: false,
		},
		{
			name: "Valid config with both Unix socket and TCP port",
			config: &config.Config{
				CommonFlagSet: config.CommonFlagSet{
					Retriever: &retrieverconf.RetrieverConf{
						Kind: "file",
						Path: "../testdata/config/valid-file.yaml",
					},
				},
				ListenPort: 8080,
				UnixSocket: "/var/run/goff/goff.sock",
			},
			wantErr: false,
		},
		{
			name: "Invalid config with neither Unix socket nor TCP port",
			config: &config.Config{
				CommonFlagSet: config.CommonFlagSet{
					Retriever: &retrieverconf.RetrieverConf{
						Kind: "file",
						Path: "../testdata/config/valid-file.yaml",
					},
				},
				ListenPort: 0,
				UnixSocket: "",
			},
			wantErr:    true,
			errMessage: "either listen port or unix socket must be configured",
		},
		{
			name: "Valid config with Unix socket in subdirectory",
			config: &config.Config{
				CommonFlagSet: config.CommonFlagSet{
					Retriever: &retrieverconf.RetrieverConf{
						Kind: "file",
						Path: "../testdata/config/valid-file.yaml",
					},
				},
				UnixSocket: "/var/run/goff/subdir/goff.sock",
			},
			wantErr: false,
		},
		{
			name: "Valid config with relative Unix socket path",
			config: &config.Config{
				CommonFlagSet: config.CommonFlagSet{
					Retriever: &retrieverconf.RetrieverConf{
						Kind: "file",
						Path: "../testdata/config/valid-file.yaml",
					},
				},
				UnixSocket: "./goff.sock",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.IsValid()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMessage != "" {
					assert.Contains(t, err.Error(), tt.errMessage)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

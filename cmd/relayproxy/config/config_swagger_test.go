package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
)

func TestConfigIsSwaggerEnabled(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.Config
		want        bool
		description string
	}{
		{
			name: "Swagger.Enabled is true",
			config: &config.Config{
				Swagger: config.Swagger{
					Enabled: true,
				},
				EnableSwagger: false,
			},
			want:        true,
			description: "Should return true when Swagger.Enabled is true, even if EnableSwagger is false",
		},
		{
			name: "Swagger.Enabled is false but EnableSwagger is true",
			config: &config.Config{
				Swagger: config.Swagger{
					Enabled: false,
				},
				EnableSwagger: true,
			},
			want:        true,
			description: "Should return true when EnableSwagger is true (deprecated field)",
		},
		{
			name: "Both Swagger.Enabled and EnableSwagger are true",
			config: &config.Config{
				Swagger: config.Swagger{
					Enabled: true,
				},
				EnableSwagger: true,
			},
			want:        true,
			description: "Should return true when both are true",
		},
		{
			name: "Both Swagger.Enabled and EnableSwagger are false",
			config: &config.Config{
				Swagger: config.Swagger{
					Enabled: false,
				},
				EnableSwagger: false,
			},
			want:        false,
			description: "Should return false when both are false",
		},
		{
			name: "Swagger.Enabled is true and EnableSwagger is false",
			config: &config.Config{
				Swagger: config.Swagger{
					Enabled: true,
					Host:    "example.com",
				},
				EnableSwagger: false,
				Host:          "legacy-host.com",
			},
			want:        true,
			description: "Should return true when Swagger.Enabled is true, regardless of EnableSwagger",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.IsSwaggerEnabled()
			assert.Equal(t, tt.want, got, tt.description)
		})
	}
}

func TestConfigSwaggerHost(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.Config
		want        string
		description string
	}{
		{
			name: "Swagger.Host is set",
			config: &config.Config{
				Swagger: config.Swagger{
					Host: "swagger.example.com",
				},
				Host: "legacy-host.com",
			},
			want:        "swagger.example.com",
			description: "Should return Swagger.Host when it is set, ignoring deprecated Host field",
		},
		{
			name: "Swagger.Host is empty but Host is set",
			config: &config.Config{
				Swagger: config.Swagger{
					Host: "",
				},
				Host: "legacy-host.com",
			},
			want:        "legacy-host.com",
			description: "Should return deprecated Host field when Swagger.Host is empty",
		},
		{
			name: "Both Swagger.Host and Host are empty",
			config: &config.Config{
				Swagger: config.Swagger{
					Host: "",
				},
				Host: "",
			},
			want:        "localhost",
			description: "Should return 'localhost' as default when both are empty",
		},
		{
			name: "Swagger.Host is empty string and Host is empty string",
			config: &config.Config{
				Swagger: config.Swagger{},
				Host:    "",
			},
			want:        "localhost",
			description: "Should return 'localhost' when both fields are empty strings",
		},
		{
			name: "Swagger.Host is set with special characters",
			config: &config.Config{
				Swagger: config.Swagger{
					Host: "swagger-api.example.com:8080",
				},
			},
			want:        "swagger-api.example.com:8080",
			description: "Should return Swagger.Host as-is, including special characters",
		},
		{
			name: "Swagger.Host is empty and Host is set",
			config: &config.Config{
				Swagger: config.Swagger{},
				Host:    "swagger-api.example.com:8080",
			},
			want:        "swagger-api.example.com:8080",
			description: "Should return Swagger.Host as-is, including special characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.SwaggerHost()
			assert.Equal(t, tt.want, got, tt.description)
		})
	}
}

package config

import (
	"fmt"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/pflag"
)

const EnvPrefix = "GOFF_MGMT_"

func Load(flags *pflag.FlagSet) (*Config, error) {
	k := koanf.New(".")

	cfg := Default()
	if err := k.Load(confmap.Provider(structToMap(cfg), "."), nil); err != nil {
		return nil, fmt.Errorf("load defaults: %w", err)
	}

	if flags != nil {
		if cfgPath, _ := flags.GetString("config"); cfgPath != "" {
			if err := k.Load(file.Provider(cfgPath), yaml.Parser()); err != nil {
				return nil, fmt.Errorf("load config file %s: %w", cfgPath, err)
			}
		}
	}

	envCb := func(s string) string {
		s = strings.TrimPrefix(s, EnvPrefix)
		s = strings.ToLower(s)
		return strings.ReplaceAll(s, "_", ".")
	}
	if err := k.Load(env.Provider(EnvPrefix, ".", envCb), nil); err != nil {
		return nil, fmt.Errorf("load env: %w", err)
	}

	if flags != nil {
		if err := k.Load(posflag.Provider(flags, ".", k), nil); err != nil {
			return nil, fmt.Errorf("load flags: %w", err)
		}
	}

	var out Config
	if err := k.Unmarshal("", &out); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	return &out, nil
}

// structToMap returns default values keyed using the same lowercase scheme
// produced by the env-var callback (see Load). Keeping both sources on a
// single key-casing convention avoids koanf storing duplicate entries
// ("auth.cookieSecure" vs "auth.cookiesecure") that mapstructure would then
// resolve non-deterministically.
func structToMap(c Config) map[string]any {
	return map[string]any{
		"server.port":            c.Server.Port,
		"server.shutdowntimeout": c.Server.ShutdownTimeout,
		"server.enableswagger":   c.Server.EnableSwagger,
		"db.maxconns":            c.DB.MaxConns,
		"db.minconns":            c.DB.MinConns,
		"db.connmaxlifetime":     c.DB.ConnMaxLifetime,
		"oidc.scopes":            c.OIDC.Scopes,
		"auth.sessionmaxage":     c.Auth.SessionMaxAge,
		"auth.cookiesecure":      c.Auth.CookieSecure,
		"auth.postloginredirect": c.Auth.PostLoginRedirect,
		"log.level":              c.Log.Level,
		"log.format":             c.Log.Format,
	}
}

func RegisterFlags(fs *pflag.FlagSet) {
	fs.String("config", "", "Path to YAML config file")
	fs.Int("server.port", 0, "HTTP server port")
	fs.String("db.url", "", "PostgreSQL DSN")
	fs.String("log.level", "", "Log level (debug|info|warn|error)")
}

package config

import "time"

type Config struct {
	Server ServerConfig `koanf:"server"`
	DB     DBConfig     `koanf:"db"`
	OIDC   OIDCConfig   `koanf:"oidc"`
	Auth   AuthConfig   `koanf:"auth"`
	Log    LogConfig    `koanf:"log"`
}

type ServerConfig struct {
	Port            int           `koanf:"port"`
	ShutdownTimeout time.Duration `koanf:"shutdownTimeout"`
	EnableSwagger   bool          `koanf:"enableSwagger"`
}

type DBConfig struct {
	URL             string        `koanf:"url"`
	MaxConns        int32         `koanf:"maxConns"`
	MinConns        int32         `koanf:"minConns"`
	ConnMaxLifetime time.Duration `koanf:"connMaxLifetime"`
}

type OIDCConfig struct {
	Issuer       string   `koanf:"issuer"`
	ClientID     string   `koanf:"clientId"`
	ClientSecret string   `koanf:"clientSecret"`
	RedirectURL  string   `koanf:"redirectUrl"`
	Scopes       []string `koanf:"scopes"`
}

type AuthConfig struct {
	JWTSecret     string        `koanf:"jwtSecret"`
	SessionMaxAge time.Duration `koanf:"sessionMaxAge"`
	CookieDomain  string        `koanf:"cookieDomain"`
	CookieSecure  bool          `koanf:"cookieSecure"`
	AdminEmails   []string      `koanf:"adminEmails"`
}

type LogConfig struct {
	Level  string `koanf:"level"`
	Format string `koanf:"format"`
}

func Default() Config {
	return Config{
		Server: ServerConfig{
			Port:            8080,
			ShutdownTimeout: 15 * time.Second,
			EnableSwagger:   true,
		},
		DB: DBConfig{
			MaxConns:        20,
			MinConns:        2,
			ConnMaxLifetime: time.Hour,
		},
		OIDC: OIDCConfig{
			Scopes: []string{"openid", "profile", "email"},
		},
		Auth: AuthConfig{
			SessionMaxAge: 24 * time.Hour,
			CookieSecure:  true,
		},
		Log: LogConfig{
			Level:  "info",
			Format: "json",
		},
	}
}

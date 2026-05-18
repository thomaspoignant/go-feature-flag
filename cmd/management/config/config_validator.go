package config

import "fmt"

func (c *Config) Validate() error {
	if c.Server.Port <= 0 {
		return fmt.Errorf("server.port must be > 0")
	}
	if c.DB.URL == "" {
		return fmt.Errorf("db.url is required")
	}
	if c.OIDC.Issuer == "" {
		return fmt.Errorf("oidc.issuer is required")
	}
	if c.OIDC.ClientID == "" {
		return fmt.Errorf("oidc.clientId is required")
	}
	if c.OIDC.ClientSecret == "" {
		return fmt.Errorf("oidc.clientSecret is required")
	}
	if c.OIDC.RedirectURL == "" {
		return fmt.Errorf("oidc.redirectUrl is required")
	}
	if c.Auth.JWTSecret == "" {
		return fmt.Errorf("auth.jwtSecret is required")
	}
	if len(c.Auth.JWTSecret) < 32 {
		return fmt.Errorf("auth.jwtSecret must be at least 32 chars")
	}
	return nil
}

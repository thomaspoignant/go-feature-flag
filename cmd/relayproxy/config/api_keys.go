package config

import "sync"

// APIKeys is a struct to store the API keys for the different endpoints
type APIKeys struct {
	Admin      []string `mapstructure:"admin"      koanf:"admin"`
	Evaluation []string `mapstructure:"evaluation" koanf:"evaluation"`
}

type ApiKeyType = string

// Enum for the type of the API keys
const (
	EvaluationKeyType ApiKeyType = "EVALUATION"
	AdminKeyType      ApiKeyType = "ADMIN"
	FlagSetKeyType    ApiKeyType = "FLAGSET"
	ErrorKeyType      ApiKeyType = "ERROR"
)

// APIKeysAdminExists is checking if an admin API Key exist in the relay proxy configuration
func (c *Config) APIKeysAdminExists(apiKey string) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.apiKeysSet[apiKey] == AdminKeyType
}

// APIKeyExists is checking if an API Key exist in the relay proxy configuration
func (c *Config) APIKeyExists(apiKey string) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	_, ok := c.apiKeysSet[apiKey]
	return ok
}

// IsAuthenticationEnabled returns true if we need to be authenticated.
func (c *Config) IsAuthenticationEnabled() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.forceAuthenticatedRequests
}

// preloadAPIKeysLocked performs the actual API key preloading logic.
// It assumes the caller holds the write lock (c.mutex.Lock()).
// This function should never be called directly; use preloadAPIKeys() or ForceReloadAPIKeys() instead.
func (c *Config) preloadAPIKeysLocked() {
	apiKeySet := make(map[string]ApiKeyType)
	addAPIKeys := func(keys []string, keyType ApiKeyType) {
		for _, k := range keys {
			apiKeySet[k] = keyType
		}
	}

	addAPIKeys(c.AuthorizedKeys.Evaluation, EvaluationKeyType)
	addAPIKeys(c.APIKeys, EvaluationKeyType)

	for _, flagSet := range c.FlagSets {
		addAPIKeys(flagSet.APIKeys, FlagSetKeyType)
	}

	// it has to be before adding the admin keys, because when we add only the admin keys,
	// we don't want to force the authentication (except for the admin endpoints).
	if len(apiKeySet) > 0 {
		c.forceAuthenticatedRequests = true
	} else {
		c.forceAuthenticatedRequests = false
	}

	addAPIKeys(c.AuthorizedKeys.Admin, AdminKeyType)
	c.apiKeysSet = apiKeySet
}

// preloadAPIKeys is storing in the struct all the API Keys available for the relay-proxy.
func (c *Config) preloadAPIKeys() {
	c.mutex.RLock()
	once := &c.apiKeyPreload
	c.mutex.RUnlock()

	once.Do(func() {
		c.mutex.Lock()
		defer c.mutex.Unlock()
		c.preloadAPIKeysLocked()
	})
}

// ForceReloadAPIKeys is forcing the reload of the API Keys.
// This is used to reload the API Keys when the configuration changes.
// The entire reload operation is atomic to prevent race conditions.
func (c *Config) ForceReloadAPIKeys() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	// Reset sync.Once to allow re-initialization
	c.apiKeyPreload = sync.Once{}
	// Reset fields
	c.forceAuthenticatedRequests = false
	c.apiKeysSet = nil
	// Repopulate immediately under the same lock to ensure atomicity
	c.preloadAPIKeysLocked()
}

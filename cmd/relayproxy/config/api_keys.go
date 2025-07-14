package config

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
	c.preloadAPIKeys()
	return c.apiKeysSet[apiKey] == AdminKeyType
}

// APIKeyExists is checking if an API Key exist in the relay proxy configuration
func (c *Config) APIKeyExists(apiKey string) bool {
	c.preloadAPIKeys()
	_, ok := c.apiKeysSet[apiKey]
	return ok
}

func (c *Config) GetAPIKeyType(apiKey string) ApiKeyType {
	c.preloadAPIKeys()
	if keyType, ok := c.apiKeysSet[apiKey]; ok {
		return keyType
	}
	return ErrorKeyType
}

// IsAuthenticationEnabled returns true if we need to be authenticated.
func (c *Config) IsAuthenticationEnabled() bool {
	c.preloadAPIKeys()
	return c.forceAuthenticatedRequests
}

// preloadAPIKeys is storing in the struct all the API Keys available for the relay-proxy.
func (c *Config) preloadAPIKeys() {
	c.apiKeyPreload.Do(func() {
		if c.apiKeysSet != nil {
			return
		}
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

		addAPIKeys(c.AuthorizedKeys.Admin, AdminKeyType)

		c.apiKeysSet = apiKeySet
		if len(apiKeySet) > 0 {
			c.forceAuthenticatedRequests = true
		}
	})
}

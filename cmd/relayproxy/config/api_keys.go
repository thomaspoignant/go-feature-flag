package config

// APIKeys is a struct to store the API keys for the different endpoints
type APIKeys struct {
	Admin      []string `mapstructure:"admin" koanf:"admin"`
	Evaluation []string `mapstructure:"evaluation" koanf:"evaluation"`
}

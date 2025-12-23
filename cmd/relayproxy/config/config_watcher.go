package config

// AttachConfigChangeCallback attaches a callback to be called when the configuration changes
func (c *Config) AttachConfigChangeCallback(callback func(newConfig *Config)) {
	if c.configLoader == nil {
		return
	}
	c.configLoader.AddConfigChangeCallback(callback)
}

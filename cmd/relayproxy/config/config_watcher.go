package config

// AttachConfigChangeCallback attaches a callback to be called when the configuration changes
func (c *Config) AttachConfigChangeCallback(callback func(newConfig *Config)) {
	if c.configLoader == nil {
		if c.logger != nil {
			c.logger.Error("configLoader is not initialized, impossible to attach a callback to the configuration changes")
		}
		return
	}
	c.configLoader.AddConfigChangeCallback(callback)
}

// StopConfigChangeWatcher stops the watcher for the configuration changes
func (c *Config) StopConfigChangeWatcher() error {
	if c.configLoader == nil {
		return nil
	}
	return c.configLoader.StopWatchChanges()
}

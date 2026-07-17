package config

func (c *Config) ApplyDefaults(defaults *Config) {
	if c == nil || defaults == nil {
		return
	}

	if c.Log == (Log{}) {
		c.Log = defaults.Log
	}

	if c.Server == (Server{}) {
		c.Server = defaults.Server
	}

	if c.Actions == nil {
		c.Actions = defaults.Actions
	}
}

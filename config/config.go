package config

func (l Log) IsZero() bool {
	return l.Dir == nil && l.Level == nil && l.FileFormat == nil
}

// Note: Prefork=false cannot distinguish "not set" from "set to false",
// so a Server with only Prefork set is treated as zero.
func (s Server) IsZero() bool {
	return s.Address == nil &&
		!s.Prefork &&
		s.TLS.Cert == nil && s.TLS.Key == nil && s.TLS.RedirectHTTP == nil &&
		s.ACME == nil
}

func ApplyDefaults(cfg *Config, defaults *Config) *Config {
	if cfg == nil || defaults == nil {
		return cfg
	}

	result := *cfg // shallow copy

	if result.Log.IsZero() {
		result.Log = defaults.Log
	}

	if result.Server.IsZero() {
		result.Server = defaults.Server
	}

	if result.Actions == nil {
		result.Actions = defaults.Actions
	}

	return &result
}

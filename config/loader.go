package config

type Loader interface {
	GetTag() string
	Load(cfg *Config, fileName string) error
	Save(cfg *Config, fileName string) error
	// GetAllowFileExtensions lowercase
	GetAllowFileExtensions() []string
}

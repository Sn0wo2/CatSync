package config

import (
	"cmp"

	"github.com/Sn0wo2/CatSync/log"
	"github.com/Sn0wo2/ordo"
	"github.com/Sn0wo2/ordo/format"
	"gopkg.in/yaml.v3"
)

var CLIConfigPath string

const defaultConfigPath = "./data/config.yml"

var loader = &ordo.Loader[Config]{
	Formats: []format.Format[Config]{
		format.NewFormatter[Config]("yaml", []string{".yml", ".yaml"}, 10, yaml.Unmarshal, yaml.Marshal),
		format.JSON[Config](),
	},
	Default:  GetDefaultConfig(),
	Validate: (*Config).Validate,
	OnLoaded: func(path string) { log.Writef("Config loaded from file: %s", path) },
}

func LoadConfig(configPath string) (*Config, string, error) {
	return loader.Load(cmp.Or(configPath, CLIConfigPath, defaultConfigPath))
}

func SaveConfig(cfg *Config, fileName string) error {
	return loader.Save(cfg, fileName)
}

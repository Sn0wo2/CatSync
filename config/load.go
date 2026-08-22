package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sn0wo2/CatSync/log"
)

var CLIConfigPath string

func LoadConfig(configPath string, loaders ...Loader) (*Config, string, error) {
	if len(loaders) == 0 {
		return nil, CLIConfigPath, errors.New("no loaders provided")
	}

	loaderByExt := make(map[string]Loader)

	for _, loader := range loaders {
		for _, ext := range loader.GetAllowFileExtensions() {
			loaderByExt["."+strings.ToLower(ext)] = loader
		}
	}

	if configPath == "" {
		configPath = "./.data/config.yml"

		if CLIConfigPath != "" {
			if _, err := os.Stat(CLIConfigPath); err == nil {
				configPath = CLIConfigPath

				goto LOAD
			}

			base := strings.TrimSuffix(CLIConfigPath, filepath.Ext(CLIConfigPath))
			for ext := range loaderByExt {
				tryPath := base + ext
				if _, err := os.Stat(tryPath); err == nil {
					configPath = tryPath

					goto LOAD
				}
			}

			configPath = CLIConfigPath

			goto LOAD
		}

		for ext := range loaderByExt {
			cfgPath := filepath.Join("./.data/", "config"+ext)
			if _, err := os.Stat(cfgPath); err == nil {
				configPath = cfgPath

				goto LOAD
			}
		}
	}

LOAD:
	fileCfg := &Config{}

	ext := strings.ToLower(filepath.Ext(configPath))

	loader, ok := loaderByExt[ext]
	if ok {
		if err := loader.Load(fileCfg, configPath); err != nil {
			return nil, configPath, fmt.Errorf("failed to load config file %s: %w", configPath, err)
		}
	} else {
		var loadErrs []error

		loaded := false

		for _, l := range loaders {
			if err := l.Load(fileCfg, configPath); err == nil {
				loaded = true

				break
			} else {
				loadErrs = append(loadErrs, fmt.Errorf("loader %s: %w", l.GetTag(), err))
				_, _ = fmt.Fprintf(os.Stderr, "loader %s failed for config file %s: %v\n", l.GetTag(), configPath, err)
			}
		}

		if !loaded {
			if len(loadErrs) == 0 {
				return nil, "", fmt.Errorf("no loader found for config file %s", configPath)
			}

			return nil, configPath, fmt.Errorf("all loaders failed for config file %s: %w", configPath, errors.Join(loadErrs...))
		}
	}

	fileCfg = ApplyDefaults(fileCfg, GetDefaultConfig())

	if err := fileCfg.Validate(); err != nil {
		return fileCfg, configPath, fmt.Errorf("validation failed for config file %s: %w", configPath, err)
	}

	log.Writef("Config loaded from file: %s", configPath)

	return fileCfg, configPath, nil
}

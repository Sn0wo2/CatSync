package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sn0wo2/CatSync/debug"
	"github.com/Sn0wo2/CatSync/log"
)

func Load(loaders ...Loader) (*LoadResult, error) {
	loaderByExt, err := buildLoaderIndex(loaders)
	if err != nil {
		return nil, err
	}

	location := findConfigPath(loaderByExt)

	result, err := loadConfig(location.Path, loaders, loaderByExt)
	if !location.Found && errors.Is(err, os.ErrNotExist) {
		return result, ErrConfigNotFound
	}

	return result, err
}

func LoadConfig(configPath string, loaders ...Loader) (*LoadResult, error) {
	loaderByExt, err := buildLoaderIndex(loaders)
	if err != nil {
		return nil, err
	}

	return loadConfig(configPath, loaders, loaderByExt)
}

func loadConfig(configPath string, loaders []Loader, loaderByExt map[string]Loader) (*LoadResult, error) {
	fileCfg := &Config{}
	ext := strings.ToLower(filepath.Ext(configPath))

	loader, ok := loaderByExt[ext]
	if ok {
		if err := loader.Load(fileCfg, configPath); err != nil {
			return &LoadResult{Path: configPath}, fmt.Errorf("failed to load config file %s: %w", configPath, err)
		}
	} else {
		var loadErr error

		loaded := false

		for i, l := range loaders {
			if i > 0 {
				_, _ = fmt.Fprintf(os.Stderr, "retrying with next loader: %s %d/%d\n", l.GetTag(), i+1, len(loaders))
			}

			if err := l.Load(fileCfg, configPath); err == nil {
				loaded = true

				break
			} else {
				loadErr = err
				_, _ = fmt.Fprintf(os.Stderr, "loader %s failed to load config file %s: %v\n", l.GetTag(), configPath, err)
			}
		}

		if !loaded {
			if loadErr == nil {
				return &LoadResult{Path: configPath}, fmt.Errorf("no loader found for config file %s", configPath)
			}

			return &LoadResult{Path: configPath}, fmt.Errorf("all loaders failed for config file %s, last error: %w", configPath, loadErr)
		}
	}

	fileCfg.ApplyDefaults(GetDefaultConfig())

	if err := fileCfg.Validate(); err != nil {
		return &LoadResult{Path: configPath}, fmt.Errorf("validation failed for config file %s: %w", configPath, err)
	}

	log.Writef("Config loaded from file: %s", configPath)

	return &LoadResult{Config: fileCfg, Path: configPath}, nil
}

type configLocation struct {
	Path  string
	Found bool
}

func buildLoaderIndex(loaders []Loader) (map[string]Loader, error) {
	if len(loaders) == 0 {
		return nil, errors.New("no loaders provided")
	}

	loaderByExt := make(map[string]Loader)

	for _, loader := range loaders {
		for _, ext := range loader.GetAllowFileExtensions() {
			loaderByExt["."+strings.ToLower(ext)] = loader
		}
	}

	return loaderByExt, nil
}

func findConfigPath(loaderByExt map[string]Loader) configLocation {
	configPath := os.Getenv("CONFIG_PATH")
	if debug.IsDebugging() {
		if p := os.Getenv("DEBUG_CONFIG_PATH"); p != "" {
			configPath = p
		}
	}

	if configPath != "" {
		//nolint:gosec // CONFIG_PATH is an operator-controlled configuration path.
		if _, err := os.Stat(configPath); err == nil {
			return configLocation{Path: configPath, Found: true}
		}

		base := strings.TrimSuffix(configPath, filepath.Ext(configPath))
		for ext := range loaderByExt {
			tryPath := base + ext
			//nolint:gosec // The fallback keeps the operator-selected configuration path.
			if _, err := os.Stat(tryPath); err == nil {
				return configLocation{Path: tryPath, Found: true}
			}
		}

		return configLocation{Path: configPath, Found: true}
	}

	for ext := range loaderByExt {
		configPath = filepath.Join("./data/", "config"+ext)
		if _, err := os.Stat(configPath); err == nil {
			return configLocation{Path: configPath, Found: true}
		}
	}

	return configLocation{Path: "./data/config.yml"}
}

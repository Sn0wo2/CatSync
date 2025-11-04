package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sn0wo2/CatSync/debug"
)

var ErrConfigNotFound = errors.New("config file not found")

func New(loaders ...Loader) (*Config, error) {
	if len(loaders) == 0 {
		return nil, errors.New("no loaders provided")
	}

	loaderByExt := make(map[string]Loader)

	for _, l := range loaders {
		for _, ext := range l.GetAllowFileExtensions() {
			loaderByExt["."+strings.ToLower(ext)] = l
		}
	}

	envPath := os.Getenv("CONFIG_PATH")
	if debug.IsDebugging() {
		if p := os.Getenv("DEBUG_CONFIG_PATH"); p != "" {
			envPath = p
		}
	}

	if envPath != "" {
		if _, err := os.Stat(envPath); err == nil {
			Path = envPath
		} else {
			base := strings.TrimSuffix(envPath, filepath.Ext(envPath))
			for ext := range loaderByExt {
				tryPath := base + ext
				if _, err := os.Stat(tryPath); err == nil {
					Path = tryPath

					break
				}
			}
		}
	}

	if Path == "" {
		searchPaths := []string{"./data/"}

	searchLoop:
		for _, p := range searchPaths {
			for ext := range loaderByExt {
				fullPath := filepath.Join(p, "config"+ext)
				if _, err := os.Stat(fullPath); err == nil {
					Path = fullPath

					break searchLoop
				}
			}
		}
	}

	if Path == "" {
		Path = envPath

		return nil, ErrConfigNotFound
	}

	ext := strings.ToLower(filepath.Ext(Path))

	loader, ok := loaderByExt[ext]
	if !ok {
		return nil, fmt.Errorf("unsupported config file extension: %s", ext)
	}

	fileCfg := &Config{}
	if err := loader.Load(fileCfg, Path); err != nil {
		return nil, fmt.Errorf("failed to load config file %s: %w", Path, err)
	}

	if err := fileCfg.Validate(loader.GetTag()); err != nil {
		return nil, fmt.Errorf("validation failed for config file %s: %w", Path, err)
	}

	fileCfg.Merge(DefaultConfig)

	return fileCfg, nil
}

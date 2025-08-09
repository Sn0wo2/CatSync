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

func Init(loaders ...Loader) error {
	var err error

	Instance, err = NewConfig(loaders...)

	return err
}

func NewConfig(loaders ...Loader) (*Config, error) {
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

	var foundPath string

	if envPath != "" {
		if _, err := os.Stat(envPath); err == nil {
			foundPath = envPath
		} else {
			base := strings.TrimSuffix(envPath, filepath.Ext(envPath))
			for ext := range loaderByExt {
				tryPath := base + ext
				if _, err := os.Stat(tryPath); err == nil {
					foundPath = tryPath

					break
				}
			}
		}
	}

	if foundPath == "" {
		searchPaths := []string{"./data/"}

	searchLoop:
		for _, p := range searchPaths {
			for ext := range loaderByExt {
				fullPath := filepath.Join(p, "config"+ext)
				if _, err := os.Stat(fullPath); err == nil {
					foundPath = fullPath

					break searchLoop
				}
			}
		}
	}

	if foundPath == "" {
		return nil, ErrConfigNotFound
	}

	ext := strings.ToLower(filepath.Ext(foundPath))

	loader, ok := loaderByExt[ext]
	if !ok {
		return nil, fmt.Errorf("unsupported config file extension: %s", ext)
	}

	cfg, err := loader.Load(foundPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config file %s: %w", foundPath, err)
	}

	return cfg, validate(cfg)
}

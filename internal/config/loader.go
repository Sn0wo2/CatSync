package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sn0wo2/FileSync/pkg/debug"
)

func Init(loaders ...Loader) error {
	var err error
	Instance, err = NewConfig(loaders...)
	return err
}

func NewConfig(loaders ...Loader) (*Config, error) {
	loaderByExt := make(map[string]Loader)
	for _, l := range loaders {
		for _, ext := range l.GetAllowFileExtensions() {
			loaderByExt["."+strings.ToLower(ext)] = l
		}
	}

	allowedExts := make([]string, 0, len(loaderByExt))
	for ext := range loaderByExt {
		allowedExts = append(allowedExts, ext)
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
			for _, ext := range allowedExts {
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
			for _, ext := range allowedExts {
				fullPath := filepath.Join(p, "config"+ext)
				if _, err := os.Stat(fullPath); err == nil {
					foundPath = fullPath
					break searchLoop
				}
			}
		}
	}

	if foundPath == "" {
		if envPath != "" {
			return nil, fmt.Errorf("config file specified by env var not found: %s", envPath)
		}
		return nil, fmt.Errorf("no config file found in search paths")
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

	if err := validate(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

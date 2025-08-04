package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sn0wo2/FileSync/pkg/debug"
)

func Init(loaders ...Loader) error {
	loaderByExt := make(map[string]Loader)
	var allowedExts []string
	{
		allowedExtsSet := make(map[string]struct{})
		for _, l := range loaders {
			for _, ext := range l.GetAllowFileExtensions() {
				dottedExt := "." + strings.ToLower(ext)
				loaderByExt[dottedExt] = l
				allowedExtsSet[dottedExt] = struct{}{}
			}
		}
		for ext := range allowedExtsSet {
			allowedExts = append(allowedExts, ext)
		}
	}

	var foundPath string

	envPath := os.Getenv("CONFIG_PATH")
	if debug.IsDebugging() {
		if p := os.Getenv("DEBUG_CONFIG_PATH"); p != "" {
			envPath = p
		}
	}

	if envPath != "" {
		if _, err := os.Stat(envPath); err == nil {
			foundPath = envPath
		} else {
			for _, ext := range allowedExts {
				tryPath := envPath + ext
				if _, err := os.Stat(tryPath); err == nil {
					foundPath = tryPath
					break
				}
			}
		}
	}

	if foundPath == "" {
		searchPaths := []string{"./data/"}
		if debug.IsDebugging() {
			searchPaths = []string{"./"}
		}

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
			return fmt.Errorf("config file specified by env var not found: %s", envPath)
		}
		return fmt.Errorf("no config file found in search paths")
	}

	ext := strings.ToLower(filepath.Ext(foundPath))
	loader, ok := loaderByExt[ext]
	if !ok {
		return fmt.Errorf("unsupported config file extension: %s", ext)
	}

	cfg, err := loader.Load(foundPath)
	if err != nil {
		return err
	}

	if err = validate(cfg); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	Instance = cfg
	return nil
}

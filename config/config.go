package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sn0wo2/CatSync/debug"
	"github.com/Sn0wo2/CatSync/internal/util"
)

func init() {
	util.DefaultConfigProvider = func() (any, bool) {
		return DefaultConfig, true
	}
}

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

	Path = os.Getenv("CONFIG_PATH")
	if debug.IsDebugging() { // debug config high priority
		if p := os.Getenv("DEBUG_CONFIG_PATH"); p != "" {
			Path = p
		}
	}

	if Path != "" {
		if _, err := os.Stat(Path); err != nil {
			base := strings.TrimSuffix(Path, filepath.Ext(Path))
			for ext := range loaderByExt {
				tryPath := base + ext
				if _, err := os.Stat(tryPath); err == nil {
					Path = tryPath

					break
				}
			}
		}
	}

	// p1 fallback
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

	fileCfg := &Config{}

	// p1 & p2 fallback
	if Path == "" {
		Path = "./data/config.yml"
		return nil, ErrConfigNotFound
	}

	ext := strings.ToLower(filepath.Ext(Path))

	loader, ok := loaderByExt[ext]
	retryIndex := 0
retryLoaders:
	if !ok {
		if retryIndex >= len(loaders) {
			return nil, fmt.Errorf("no loader found for config file %s", Path)
		}

		loader = loaders[retryIndex]
		retryIndex++
		_, _ = fmt.Fprintf(os.Stderr, "failed to find config loader %s. Retrying with next loader: %s %d/%d\n", Path, loader.GetTag(), retryIndex, len(loaders))
	}

	if err := loader.Load(fileCfg, Path); err != nil {
		if !ok {
			_, _ = fmt.Fprintf(os.Stderr, "loader %s failed to load config file %s: %v. Retrying with next loader... %d/%d: %v\n", loader.GetTag(), Path, err, retryIndex, len(loaders), err)
			goto retryLoaders
		}
		return nil, fmt.Errorf("failed to load config file %s: %w", Path, err)
	}

	if err := fileCfg.Validate(loader.GetTag()); err != nil {
		return nil, fmt.Errorf("validation failed for config file %s: %w", Path, err)
	}

	fileCfg.Merge(DefaultConfig)

	fileCfg.Check()

	return fileCfg, nil
}

func (c *Config) Validate(tag string) error {
	return util.Validate(c, tag)
}

func (c *Config) Merge(src *Config) {
	util.Merge(src, c)
}

func (c *Config) Check() {
	for i, a := range c.Actions {
		if a.Operation == "" {
			if a.Action == 0 {
				_, _ = fmt.Fprintf(os.Stderr, "Config > Action 'operation' field is empty!\n"+
					" - route: [%s]\n"+
					" - index: [%d]\n", a.Route, i)
			} else {
				_, _ = fmt.Fprintf(os.Stderr, "Config > Action 'action' field is deprecated and will be removed in v1.11.0. Please use the 'operation' field instead.\n"+
					" - route: [%s]\n"+
					" - index: [%d]\n", a.Route, i)
			}

			a.Operation = a.Action.ToOperation()
		}
	}
}

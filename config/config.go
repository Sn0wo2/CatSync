package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sn0wo2/CatSync/debug"
	"github.com/Sn0wo2/CatSync/internal/util"
	"go.uber.org/zap"
)

func init() {
	util.DefaultConfigProvider = func() (any, bool) {
		return GetDefaultConfig(), true
	}
}

func (c *Config) Reload(loaders ...Loader) error {
	newCfg, err := New(loaders...)
	if err != nil {
		return err
	}

	*c = *newCfg

	return nil
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
	if debug.IsDebugging() {
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

	fileCfg.Merge(GetDefaultConfig())

	return fileCfg, nil
}

func (c *Config) Validate(tag string) error {
	return util.Validate(c, tag)
}

func (c *Config) Merge(src *Config) {
	util.Merge(src, c)
}

const (
	defaultACMEHTTP01 = "http-01"
	defaultACMEDNS01  = "dns-01"
	defaultDNSExec    = "exec"
)

type checkErrs struct {
	err []error
}

func (e *checkErrs) add(err error) {
	if err != nil {
		e.err = append(e.err, err)
	}
}

func checkACME(c *Config, add func(error)) {
	if c.Server.ACME == nil || !c.Server.ACME.Enable {
		return
	}

	if len(c.Server.ACME.Hosts) == 0 {
		add(errors.New("server.acme.hosts is required when server.acme.enable=true"))
	}

	if c.Server.ACME.HTTP01 != nil && c.Server.ACME.DNS01 != nil {
		add(errors.New("server.acme.http01 and server.acme.dns01 are mutually exclusive"))

		return
	}

	// Only validate dns01 when it is configured.
	if c.Server.ACME.DNS01 == nil {
		return
	}

	provider := strings.ToLower(strings.TrimSpace(c.Server.ACME.DNS01.Provider))
	if provider == "" {
		provider = defaultDNSExec
	}

	switch provider {
	case defaultDNSExec, "cloudflare", "dnspod", "alidns", "route53":
		// ok
	default:
		add(fmt.Errorf("invalid server.acme.dns01.provider: %q", c.Server.ACME.DNS01.Provider))
	}

	if provider != defaultDNSExec {
		return
	}

	if len(c.Server.ACME.DNS01.PresentCmd) == 0 {
		add(errors.New("server.acme.dns01.presentCmd is required for exec provider"))
	}

	if len(c.Server.ACME.DNS01.CleanUpCmd) == 0 {
		add(errors.New("server.acme.dns01.cleanupCmd is required for exec provider"))
	}
}

func checkStatus(where string, status uint16) error {
	if status == 0 {
		return nil
	}

	if status < 100 || status > 599 {
		return fmt.Errorf("invalid status code at %s: %d", where, status)
	}

	return nil
}

func checkAuth(where string, auth *ActionModifierAuth, actionCount int) error {
	if auth == nil {
		return nil
	}

	var authErrs []error

	addAuthErr := func(err error) {
		if err != nil {
			authErrs = append(authErrs, err)
		}
	}

	if auth.Fallback == nil || auth.Fallback.Type == "" {
		addAuthErr(fmt.Errorf("auth fallback is required at %s", where))

		return errors.Join(authErrs...)
	}

	switch auth.Fallback.Type {
	case AuthFallbackNext:
		// ok
	case AuthFallbackJump:
		if actionCount == 0 {
			addAuthErr(fmt.Errorf("auth fallback jumpTo out of range at %s: no actions", where))

			return errors.Join(authErrs...)
		}

		if auth.Fallback.JumpTo < 0 || auth.Fallback.JumpTo >= actionCount {
			addAuthErr(fmt.Errorf("auth fallback jumpTo out of range at %s: %d", where, auth.Fallback.JumpTo))
		}
	default:
		addAuthErr(fmt.Errorf("invalid auth fallback type at %s: %q", where, auth.Fallback.Type))
	}

	for k, patterns := range auth.Header {
		for _, pat := range patterns {
			if _, err := util.GetCompiledRegexp(pat); err != nil {
				addAuthErr(fmt.Errorf("invalid auth header regexp at %s (header=%q, pattern=%q): %w", where, k, pat, err))
			}
		}
	}

	for k, pat := range auth.Query {
		if _, err := util.GetCompiledRegexp(pat); err != nil {
			addAuthErr(fmt.Errorf("invalid auth query regexp at %s (key=%q, pattern=%q): %w", where, k, pat, err))
		}
	}

	return errors.Join(authErrs...)
}

func checkGlobalModifiers(c *Config, add func(error)) {
	actionCount := len(c.Actions)
	for i, gm := range c.Modifiers {
		if gm.ActionModifierStatus != nil {
			add(checkStatus(fmt.Sprintf("modifiers[%d].actionModifierStatus", i), gm.Status))
		}

		add(checkAuth(fmt.Sprintf("modifiers[%d].actionModifierAuth", i), gm.ActionModifierAuth, actionCount))
	}
}

func checkActions(c *Config, logger *zap.Logger, add func(error)) {
	actionCount := len(c.Actions)
	for i, act := range c.Actions {
		if act.Route == "" {
			logger.Warn("Config >> action route is empty; action is jump-only",
				zap.Int("index", i),
				zap.String("type", string(act.Type)),
			)
		} else {
			if _, err := util.GetCompiledRegexp(act.Route); err != nil {
				add(fmt.Errorf("invalid action route regexp at actions[%d].route (%q): %w", i, act.Route, err))
			}
		}

		if act.ActionModifierStatus != nil {
			add(checkStatus(fmt.Sprintf("actions[%d].actionModifierStatus", i), act.Status))
		}

		add(checkAuth(fmt.Sprintf("actions[%d].actionModifierAuth", i), act.ActionModifierAuth, actionCount))

		switch act.Type {
		case ActionString:
			if act.ActionString == nil {
				add(fmt.Errorf("actions[%d] type=string but string is nil", i))

				break
			}

			if act.ActionString.ActionModifierStatus != nil {
				add(checkStatus(fmt.Sprintf("actions[%d].string.actionModifierStatus", i), act.ActionString.Status))
			}

			add(checkAuth(fmt.Sprintf("actions[%d].string.actionModifierAuth", i), act.ActionString.ActionModifierAuth, actionCount))
		case ActionFile:
			if act.ActionFile == nil {
				add(fmt.Errorf("actions[%d] type=file but file is nil", i))

				break
			}

			if act.ActionFile.ActionModifierStatus != nil {
				add(checkStatus(fmt.Sprintf("actions[%d].file.actionModifierStatus", i), act.ActionFile.Status))
			}

			add(checkAuth(fmt.Sprintf("actions[%d].file.actionModifierAuth", i), act.ActionFile.ActionModifierAuth, actionCount))
		}
	}
}

func (c *Config) Check(logger *zap.Logger) error {
	if c == nil {
		return errors.New("nil config")
	}

	if logger == nil {
		return errors.New("nil logger")
	}

	actionCount := len(c.Actions)

	ec := &checkErrs{}
	addErr := ec.add

	// 0) ACME config check.
	checkACME(c, addErr)

	// 1) Notfound behavior: always the last action.
	if actionCount == 0 {
		logger.Warn("Config >> no actions configured; router will fall through to fiber (ctx.Next())")
	} else {
		last := c.Actions[actionCount-1]
		logger.Info("Config >> notfound handler is the last action",
			zap.Int("index", actionCount-1),
			zap.String("type", string(last.Type)),
		)

		if last.Type == ActionFile {
			logger.Warn("Config >> notfound handler is file action; may leak file contents",
				zap.Int("index", actionCount-1),
				zap.String("type", string(last.Type)),
			)
		}
	}

	// 2) Validate and precompile patterns, enforce auth fallback requirements.
	checkGlobalModifiers(c, addErr)
	checkActions(c, logger, addErr)

	return errors.Join(ec.err...)
}

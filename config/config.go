package config

import (
	"errors"
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"

	"github.com/Sn0wo2/CatSync/config/reader"
	"github.com/Sn0wo2/CatSync/debug"
	"github.com/Sn0wo2/CatSync/internal/util"
	"github.com/Sn0wo2/CatSync/log"
	"go.uber.org/zap"
)

var currentConfig atomic.Pointer[Config]

func SetCurrentConfig(cfg *Config) {
	currentConfig.Store(cfg)
}

func GetCurrentConfig() *Config {
	return currentConfig.Load()
}

func (c *Config) ResetStrings() {
	var stack []reflect.Value

	stack = append(stack, reflect.ValueOf(c).Elem())

	for len(stack) > 0 {
		v := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if v.IsZero() {
			continue
		}

		if v.Type() == reflect.TypeOf((*reader.String)(nil)) {
			if s, ok := v.Addr().Interface().(*reader.String); ok {
				s.Reset()
			}

			continue
		}

		switch v.Kind() {
		case reflect.Ptr:
			if !v.IsNil() {
				stack = append(stack, v.Elem())
			}
		case reflect.Slice:
			for i := range v.Len() {
				elem := v.Index(i)
				if elem.Kind() == reflect.Ptr && !elem.IsNil() {
					stack = append(stack, elem.Elem())
				} else if elem.Kind() == reflect.Struct {
					stack = append(stack, elem)
				}
			}
		case reflect.Struct:
			for j := range v.NumField() {
				stack = append(stack, v.Field(j))
			}
		case reflect.Invalid,
			reflect.Bool,
			reflect.Int,
			reflect.Int8,
			reflect.Int16,
			reflect.Int32,
			reflect.Int64,
			reflect.Uint,
			reflect.Uint8,
			reflect.Uint16,
			reflect.Uint32,
			reflect.Uint64,
			reflect.Uintptr,
			reflect.Float32,
			reflect.Float64,
			reflect.Complex64,
			reflect.Complex128,
			reflect.Array,
			reflect.Chan,
			reflect.Func,
			reflect.Interface,
			reflect.Map,
			reflect.String,
			reflect.UnsafePointer:
		}
	}
}

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

	newCfg.ResetStrings()

	currentConfig.Store(newCfg)

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

	findConfigPath := func() string {
		if Path != "" {
			if _, err := os.Stat(Path); err == nil {
				return Path
			}

			base := strings.TrimSuffix(Path, filepath.Ext(Path))
			for ext := range loaderByExt {
				tryPath := base + ext
				if _, err := os.Stat(tryPath); err == nil {
					return tryPath
				}
			}

			return Path // will fail later
		}

		searchPaths := []string{"./data/"}
		for _, p := range searchPaths {
			for ext := range loaderByExt {
				fullPath := filepath.Join(p, "config"+ext)
				if _, err := os.Stat(fullPath); err == nil {
					return fullPath
				}
			}
		}

		return ""
	}

	Path = findConfigPath()

	fileCfg := &Config{}

	if Path == "" {
		Path = "./data/config.yml"

		return nil, ErrConfigNotFound
	}

	ext := strings.ToLower(filepath.Ext(Path))

	loader, ok := loaderByExt[ext]
	if ok {
		if err := loader.Load(fileCfg, Path); err != nil {
			return nil, fmt.Errorf("failed to load config file %s: %w", Path, err)
		}
	} else {
		var loadErr error

		loaded := false

		for i, l := range loaders {
			if i > 0 {
				_, _ = fmt.Fprintf(os.Stderr, "retrying with next loader: %s %d/%d\n", l.GetTag(), i+1, len(loaders))
			}

			if err := l.Load(fileCfg, Path); err == nil {
				loader = l
				loaded = true

				break
			} else {
				loadErr = err
				_, _ = fmt.Fprintf(os.Stderr, "loader %s failed to load config file %s: %v\n", l.GetTag(), Path, err)
			}
		}

		if !loaded {
			if loadErr == nil {
				return nil, fmt.Errorf("no loader found for config file %s", Path)
			}

			return nil, fmt.Errorf("all loaders failed for config file %s, last error: %w", Path, loadErr)
		}
	}

	if err := fileCfg.Validate(loader.GetTag()); err != nil {
		return nil, fmt.Errorf("validation failed for config file %s: %w", Path, err)
	}

	fileCfg.Merge(GetDefaultConfig())

	log.Writef("Config loaded from file: %s", Path)

	return fileCfg, nil
}

func (c *Config) Validate(tag string) error {
	return util.Validate(c, tag)
}

func (c *Config) Merge(src *Config) {
	util.Merge(src, c)
}

const defaultDNSExec = "exec"

type checkErrs struct {
	err []error
}

func (e *checkErrs) add(err error) {
	if err != nil {
		e.err = append(e.err, err)
	}
}

func (c *Config) checkACME(add func(error)) {
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

	if c.Server.ACME.DNS01 == nil {
		return
	}

	provider, ok := c.Server.ACME.DNS01.Provider.LiteralTrim()
	if !ok {
		add(errors.New("server.acme.dns01.provider must be a literal string (type=string)"))

		return
	}

	provider = strings.ToLower(provider)
	if provider == "" {
		provider = defaultDNSExec
	}

	switch provider {
	case defaultDNSExec, "cloudflare", "dnspod", "alidns", "route53":
		// ok
	default:
		add(fmt.Errorf("invalid server.acme.dns01.provider: %q", provider))
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

func (c *Config) checkStatus(where string, status uint16) error {
	if status == 0 {
		return nil
	}

	if status < 100 || status > 599 {
		return fmt.Errorf("invalid status code at %s: %d", where, status)
	}

	return nil
}

func (c *Config) checkAuth(where string, auth *ActionModifierAuth, actionCount int) error {
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
		for _, pr := range patterns {
			pat, ok := pr.LiteralTrim()
			if !ok {
				addAuthErr(fmt.Errorf("auth.header pattern must be literal string at %s (header=%q)", where, k))

				continue
			}

			if pat == "" {
				addAuthErr(fmt.Errorf("auth.header pattern is empty at %s (header=%q)", where, k))

				continue
			}

			if _, err := util.GetCompiledRegexp(pat); err != nil {
				addAuthErr(fmt.Errorf("invalid auth header regexp at %s (header=%q, pattern=%q): %w", where, k, pat, err))
			}
		}
	}

	for k, pr := range auth.Query {
		pat, ok := pr.LiteralTrim()
		if !ok {
			addAuthErr(fmt.Errorf("auth.query pattern must be literal string at %s (key=%q)", where, k))

			continue
		}

		if pat == "" {
			addAuthErr(fmt.Errorf("auth.query pattern is empty at %s (key=%q)", where, k))

			continue
		}

		if _, err := util.GetCompiledRegexp(pat); err != nil {
			addAuthErr(fmt.Errorf("invalid auth query regexp at %s (key=%q, pattern=%q): %w", where, k, pat, err))
		}
	}

	if len(auth.IPAllow) > 0 {
		for _, raw := range auth.IPAllow {
			s := strings.TrimSpace(raw)
			if s == "" {
				continue
			}

			if strings.Contains(s, "/") {
				if _, err := netip.ParsePrefix(s); err != nil {
					addAuthErr(fmt.Errorf("invalid auth ipAllowlist CIDR at %s (value=%q): %w", where, raw, err))
				}

				continue
			}

			if _, err := netip.ParseAddr(s); err != nil {
				addAuthErr(fmt.Errorf("invalid auth ipAllowlist ip at %s (value=%q): %w", where, raw, err))
			}
		}
	}

	if auth.IPFile != nil {
		if err := auth.IPFile.ValidateNoIO(); err != nil {
			addAuthErr(fmt.Errorf("invalid auth ipAllowlistFile at %s: %w", where, err))

			return errors.Join(authErrs...)
		}
	}

	return errors.Join(authErrs...)
}

func (c *Config) checkGlobalModifiers(add func(error)) {
	actionCount := len(c.Actions)
	for i, gm := range c.Modifiers {
		if gm.ActionModifierStatus != nil {
			add(c.checkStatus(fmt.Sprintf("modifiers[%d].actionModifierStatus", i), gm.Status))
		}

		add(c.checkAuth(fmt.Sprintf("modifiers[%d].actionModifierAuth", i), gm.ActionModifierAuth, actionCount))
	}
}

func (c *Config) checkActions(logger *zap.Logger, add func(error)) {
	actionCount := len(c.Actions)
	for i, act := range c.Actions {
	route, ok := act.Route.LiteralTrim()
		if !ok {
			add(fmt.Errorf("actions[%d].route must be a literal string (type=string)", i))

			route = ""
		}

		if route == "" {
			logger.Info("Config >> action route is empty; action is jump-only",
				zap.Int("index", i),
				zap.String("type", string(act.Type)),
			)
		} else {
			if _, err := util.GetCompiledRegexp(route); err != nil {
				add(fmt.Errorf("invalid action route regexp at actions[%d].route (%q): %w", i, route, err))
			}
		}

		if act.ActionModifierStatus != nil {
			add(c.checkStatus(fmt.Sprintf("actions[%d].actionModifierStatus", i), act.Status))
		}

		add(c.checkAuth(fmt.Sprintf("actions[%d].actionModifierAuth", i), act.ActionModifierAuth, actionCount))

		payload := act.GetPayload()
		if payload == nil {
			add(fmt.Errorf("actions[%d] type=%s but payload is nil", i, act.Type))
		} else {
			if act.Type == ActionServer {
				if act.ActionServer.Directory == nil {
					add(fmt.Errorf("actions[%d] type=server but server.directory is nil", i))
				}
			}

			c.checkActionGlobalModifier(fmt.Sprintf("actions[%d].%s", i, act.Type), payload.GetGlobalModifier(), actionCount, add)
		}
	}
}

// embedded in an action's GlobalModifier.
func (c *Config) checkActionGlobalModifier(prefix string, gm *GlobalModifier, actionCount int, add func(error)) {
	if gm.ActionModifierStatus != nil {
		add(c.checkStatus(prefix+".actionModifierStatus", gm.Status))
	}

	add(c.checkAuth(prefix+".actionModifierAuth", gm.ActionModifierAuth, actionCount))
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

	c.checkACME(addErr)

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

	c.checkGlobalModifiers(addErr)
	c.checkActions(logger, addErr)

	return errors.Join(ec.err...)
}

package config

import (
	"errors"
	"fmt"
	"net/netip"
	"strings"

	"github.com/Sn0wo2/CatSync/config/reader"
	"github.com/Sn0wo2/CatSync/internal/util"
	"go.uber.org/zap"
)

func validateRequiredString(where string, value *reader.String) error {
	if value == nil {
		return fmt.Errorf("%s is required", where)
	}

	if err := value.ValidateNoIO(); err != nil {
		return fmt.Errorf("invalid %s: %w", where, err)
	}

	return nil
}

func validateOptionalString(where string, value *reader.String) error {
	if value == nil {
		return nil
	}

	if content, literal := value.LiteralTrim(); literal && content == "" {
		return nil
	}

	if err := value.ValidateNoIO(); err != nil {
		return fmt.Errorf("invalid %s: %w", where, err)
	}

	return nil
}

func (c *Config) Validate() error {
	if c == nil {
		return errors.New("nil config")
	}

	var ec []error

	addErr := func(err error) {
		if err != nil {
			ec = append(ec, err)
		}
	}

	addErr(validateRequiredString("log.fileFormat", c.Log.FileFormat))
	addErr(validateRequiredString("server.address", c.Server.Address))

	if len(c.Actions) == 0 {
		addErr(errors.New("actions is required"))
	}

	for i := range c.Actions {
		if c.Actions[i].TypeName() == "" {
			addErr(fmt.Errorf("actions[%d].type is required", i))
		}
	}

	actionCount := len(c.Actions)

	labelIndex := make(map[string]int, actionCount)

	for i, act := range c.Actions {
		if act.Label != "" {
			labelIndex[act.Label] = i
		}
	}

	for i, gm := range c.Modifiers {
		c.validateModifier(fmt.Sprintf("modifiers[%d]", i), &gm, actionCount, labelIndex, addErr)
	}

	for i, act := range c.Actions {
		if act.Label != "" {
			if _, dup := labelIndex[act.Label]; dup {
				addErr(fmt.Errorf("duplicate action label at actions[%d]: %q", i, act.Label))
			} else {
				labelIndex[act.Label] = i
			}
		}

		route, ok := act.Route.LiteralTrim()
		if !ok {
			addErr(fmt.Errorf("actions[%d].route must be a literal string (type=string)", i))

			route = ""
		}

		if route != "" {
			if _, err := util.GetCompiledRegexp(route); err != nil {
				addErr(fmt.Errorf("invalid action route regexp at actions[%d].route (%q): %w", i, route, err))
			}
		}

		c.validateModifier(fmt.Sprintf("actions[%d]", i), &act.GlobalModifier, actionCount, labelIndex, addErr)

		payload := act.GetPayload()
		if payload == nil {
			addErr(fmt.Errorf("actions[%d] type=%s but payload is nil", i, act.TypeName()))

			continue
		}

		switch act.TypeName() {
		case ActionFile:
			addErr(validateRequiredString(fmt.Sprintf("actions[%d].file.path", i), act.ActionFile.Path))
		case ActionString:
			addErr(validateOptionalString(fmt.Sprintf("actions[%d].string.content", i), act.ActionString.Content))
		case ActionServer:
			addErr(validateRequiredString(fmt.Sprintf("actions[%d].server.directory", i), act.ActionServer.Directory))

			for j, indexFile := range act.ActionServer.IndexFiles {
				addErr(validateRequiredString(fmt.Sprintf("actions[%d].server.indexFiles[%d]", i, j), indexFile))
			}
		case ActionReload:
		}

		c.validateModifier(fmt.Sprintf("actions[%d].%s", i, act.TypeName()), payload.GetGlobalModifier(), actionCount, labelIndex, addErr)
	}

	return errors.Join(ec...)
}

func (c *Config) validateModifier(prefix string, modifier *GlobalModifier, actionCount int, labels map[string]int, add func(error)) {
	modifier.EachModifier(func(m any) {
		switch mod := m.(type) {
		case *ActionModifierStatus:
			where := prefix + ".actionModifierStatus"

			if err := validateOptionalString(where+".upstream", mod.Upstream); err != nil {
				add(err)
				return
			}

			if modifier.Status == 0 {
				if err := validateRequiredString(where+".upstream", mod.Upstream); err != nil {
					add(fmt.Errorf("status or valid upstream is required at %s: %w", where, err))
					return
				}
			}

			if modifier.Status != 0 && (modifier.Status < 100 || modifier.Status > 599) {
				add(fmt.Errorf("invalid status code at %s: %d", where, modifier.Status))
				return
			}
		case *ActionModifierAuth:
			where := prefix + ".actionModifierAuth"
			if mod.Fallback == nil || mod.Fallback.Type == "" {
				add(fmt.Errorf("auth fallback is required at %s", where))
				return
			}

			switch mod.Fallback.Type {
			case AuthFallbackJump:
				if actionCount == 0 {
					add(fmt.Errorf("auth fallback jumpTo out of range at %s: no actions", where))

					return
				}

				if mod.Fallback.JumpLabel != "" {
					if _, exists := labels[mod.Fallback.JumpLabel]; !exists {
						add(fmt.Errorf("auth fallback jumpLabel %q not found at %s", mod.Fallback.JumpLabel, where))
					}
				} else if mod.Fallback.JumpTo < 0 || mod.Fallback.JumpTo >= actionCount {
					add(fmt.Errorf("auth fallback jumpTo out of range at %s: %d", where, mod.Fallback.JumpTo))
				}
			case AuthFallbackNext:
			default:
				add(fmt.Errorf("invalid auth fallback type at %s: %q", where, mod.Fallback.Type))
			}

			for k, patterns := range mod.Header {
				for _, pr := range patterns {
					pat, ok := pr.LiteralTrim()
					if !ok || pat == "" {
						add(fmt.Errorf("auth.header pattern is empty at %s (header=%q)", where, k))

						continue
					}

					if _, err := util.GetCompiledRegexp(pat); err != nil {
						add(fmt.Errorf("invalid auth header regexp at %s (header=%q, pattern=%q): %w", where, k, pat, err))
					}
				}
			}

			for k, pr := range mod.Query {
				pat, ok := pr.LiteralTrim()
				if !ok || pat == "" {
					add(fmt.Errorf("auth.query pattern is empty at %s (key=%q)", where, k))

					continue
				}

				if _, err := util.GetCompiledRegexp(pat); err != nil {
					add(fmt.Errorf("invalid auth query regexp at %s (key=%q, pattern=%q): %w", where, k, pat, err))
				}
			}

			if len(mod.IPWhiteList) > 0 {
				for _, raw := range mod.IPWhiteList {
					s := strings.TrimSpace(raw)
					if s == "" {
						continue
					}

					if strings.Contains(s, "/") {
						if _, err := netip.ParsePrefix(s); err != nil {
							add(fmt.Errorf("invalid auth ipAllowlist CIDR at %s (value=%q): %w", where, raw, err))
						}

						continue
					}

					if _, err := netip.ParseAddr(s); err != nil {
						add(fmt.Errorf("invalid auth ipAllowlist ip at %s (value=%q): %w", where, raw, err))
					}
				}
			}

			if mod.IPFile != nil {
				if err := mod.IPFile.ValidateNoIO(); err != nil {
					add(fmt.Errorf("invalid auth ipAllowlistFile at %s: %w", where, err))

					return
				}
			}
		case *ActionModifierResponseHeader:
			add(validateOptionalString(prefix+".actionModifierResponseHeader.upstream", mod.Upstream))
		case *ActionModifierVersion:
			add(validateRequiredString(prefix+".actionVersionModifier.placeholder", mod.Placeholder))
		case *ActionModifierAge:
			where := prefix + ".actionModifierAge"

			if len(modifier.Recipients) == 0 {
				add(fmt.Errorf("at least one recipient is required at %s", where))
			}

			for i, r := range modifier.Recipients {
				if err := validateRequiredString(fmt.Sprintf("%s.recipients[%d]", where, i), r); err != nil {
					add(err)
				}
			}
		}
	})
}

func (c *Config) LogWarnings(logger *zap.Logger) {
	if c == nil || logger == nil {
		return
	}

	if len(c.Actions) == 0 {
		logger.Warn("Config >> no actions configured; router will fall through to fiber (ctx.Next())")

		return
	}

	for i := range c.Actions {
		action := &c.Actions[i]

		route, _ := action.Route.LiteralTrim()
		if route == "" {
			logger.Info("Config >> action route is empty; action is jump-only",
				zap.Int("index", i),
				zap.String("type", string(action.TypeName())),
			)
		}
	}

	lastIndex := len(c.Actions) - 1
	last := c.Actions[lastIndex]
	logger.Info("Config >> notfound handler is the last action",
		zap.Int("index", lastIndex),
		zap.String("type", string(last.TypeName())),
	)

	if last.TypeName() == ActionFile {
		logger.Warn("Config >> notfound handler is file action; may leak file contents",
			zap.Int("index", lastIndex),
			zap.String("type", string(last.TypeName())),
		)
	}
}

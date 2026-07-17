package config

import (
	"errors"
	"fmt"
	"net/netip"
	"strings"

	"github.com/Sn0wo2/CatSync/internal/util"
)

const defaultDNSExec = "exec"

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

func (c *Config) checkStatus(where string, modifier *ActionModifierStatus) error {
	if err := validateOptionalString(where+".upstream", modifier.Upstream); err != nil {
		return err
	}

	if modifier.Status == 0 {
		if err := validateRequiredString(where+".upstream", modifier.Upstream); err != nil {
			return fmt.Errorf("status or valid upstream is required at %s: %w", where, err)
		}
	}

	if modifier.Status != 0 && (modifier.Status < 100 || modifier.Status > 599) {
		return fmt.Errorf("invalid status code at %s: %d", where, modifier.Status)
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
			if !ok || pat == "" {
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
		if !ok || pat == "" {
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

func (c *Config) validateModifier(prefix string, modifier *GlobalModifier, actionCount int, add func(error)) {
	modifier.Visit(GlobalModifierVisitor{
		Status: func(status *ActionModifierStatus) {
			add(c.checkStatus(prefix+".actionModifierStatus", status))
		},
		Auth: func(auth *ActionModifierAuth) {
			add(c.checkAuth(prefix+".actionModifierAuth", auth, actionCount))
		},
		ResponseHeader: func(header *ActionModifierResponseHeader) {
			add(validateOptionalString(prefix+".actionModifierResponseHeader.upstream", header.Upstream))
		},
		Version: func(version *ActionModifierVersion) {
			add(validateRequiredString(prefix+".actionVersionModifier.placeholder", version.Placeholder))
		},
	})
}

func (c *Config) checkGlobalModifiers(add func(error)) {
	actionCount := len(c.Actions)
	for i, gm := range c.Modifiers {
		c.validateModifier(fmt.Sprintf("modifiers[%d]", i), &gm, actionCount, add)
	}
}

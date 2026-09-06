package action

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"filippo.io/age"
	"github.com/Sn0wo2/CatSync/cli"
	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/internal/util"
)

func BuildGlobalModifiers(cfg *config.Config) ([]Modifier, error) {
	if cfg == nil {
		return nil, nil
	}

	modifiers := make([]Modifier, 0, len(cfg.Modifiers))
	for i := range cfg.Modifiers {
		mods, err := BuildModifiers(&cfg.Modifiers[i])
		if err != nil {
			return nil, fmt.Errorf("modifiers[%d]: %w", i, err)
		}

		modifiers = append(modifiers, mods...)
	}

	return modifiers, nil
}

func BuildModifiers(gm *config.GlobalModifier) ([]Modifier, error) {
	if gm == nil {
		return nil, nil
	}

	modifiers := make([]Modifier, 0, 4)

	var buildErr error

	gm.EachModifier(func(m any) {
		if buildErr != nil {
			return
		}

		switch mod := m.(type) {
		case *config.ActionModifierStatus:
			upstream, err := mod.Upstream.ReadString(context.Background())
			if err != nil {
				buildErr = fmt.Errorf("actionModifierStatus.upstream: %w", err)

				return
			}

			modifiers = append(modifiers, &StatusModifier{status: mod.Status, upstream: upstream})
		case *config.ActionModifierAuth:
			m, err := buildAuthModifier(mod)
			if err != nil {
				buildErr = err

				return
			}

			modifiers = append(modifiers, m)
		case *config.ActionModifierResponseHeader:
			upstream, err := mod.Upstream.ReadString(context.Background())
			if err != nil {
				buildErr = fmt.Errorf("actionModifierResponseHeader.upstream: %w", err)

				return
			}

			modifiers = append(modifiers, &ResponseHeaderModifier{header: mod.Header, upstream: upstream})
		case *config.ActionModifierVersion:
			if mod.Placeholder == nil {
				buildErr = errors.New("actionVersionModifier.placeholder is required")

				return
			}

			placeholder, err := mod.Placeholder.ReadString(context.Background())
			if err != nil {
				buildErr = fmt.Errorf("actionVersionModifier.placeholder: %w", err)

				return
			}

			modifiers = append(modifiers, &PlaceholderModifier{placeholder: placeholder, value: cli.GetFormatVersion()})
		case *config.ActionModifierAge:
			recipients := make([]age.Recipient, 0, len(mod.Recipients))

			for i, r := range mod.Recipients {
				if r == nil {
					continue
				}

				key, err := r.ReadString(context.Background())
				if err != nil {
					buildErr = fmt.Errorf("actionModifierAge.recipients[%d]: %w", i, err)

					return
				}

				parsed, err := age.ParseRecipients(strings.NewReader(strings.TrimSpace(key)))
				if err != nil {
					buildErr = fmt.Errorf("actionModifierAge.recipients[%d]: %w", i, err)

					return
				}

				recipients = append(recipients, parsed...)
			}

			armorOut := true
			if mod.Armor != nil {
				armorOut = *mod.Armor
			}

			modifiers = append(modifiers, &AgeModifier{recipients: recipients, armor: armorOut})
		}
	})

	if buildErr != nil {
		return nil, buildErr
	}

	return modifiers, nil
}

func buildAuthModifier(mod *config.ActionModifierAuth) (Modifier, error) {
	m := &AuthModifier{}

	if len(mod.Header) > 0 {
		headerRules := make(map[string][]*regexp.Regexp, len(mod.Header))

		for k, patterns := range mod.Header {
			rules := make([]*regexp.Regexp, 0, len(patterns))

			for j, p := range patterns {
				pat, ok := p.LiteralTrim()
				if !ok {
					return nil, fmt.Errorf("header[%q][%d] pattern must be literal string", k, j)
				}

				re, err := util.GetCompiledRegexp(pat)
				if err != nil {
					return nil, fmt.Errorf("header[%q][%d]: %w", k, j, err)
				}

				rules = append(rules, re)
			}

			headerRules[k] = rules
		}

		m.headerRules = headerRules
	}

	if len(mod.Query) > 0 {
		queryRules := make(map[string]*regexp.Regexp, len(mod.Query))

		for k, p := range mod.Query {
			pat, ok := p.LiteralTrim()
			if !ok {
				return nil, fmt.Errorf("query[%q] pattern must be literal string", k)
			}

			re, err := util.GetCompiledRegexp(pat)
			if err != nil {
				return nil, fmt.Errorf("query[%q]: %w", k, err)
			}

			queryRules[k] = re
		}

		m.queryRules = queryRules
	}

	if len(mod.IPWhiteList) > 0 || mod.IPFile != nil {
		entries := append([]string(nil), mod.IPWhiteList...)

		if mod.IPFile != nil {
			lines, err := mod.IPFile.ReadLines(context.Background())
			if err != nil {
				return nil, fmt.Errorf("ipAllowlistFile: %w", err)
			}

			entries = append(entries, lines...)
		}

		wl, err := ParseIPWhiteList(entries)
		if err != nil {
			return nil, fmt.Errorf("ipAllowlist: %w", err)
		}

		m.ipWhiteList = wl
	}

	if mod.Fallback != nil {
		policy := AuthFallbackPolicyNext
		jumpTo := 0

		if mod.Fallback.Type == config.AuthFallbackJump {
			policy = AuthFallbackPolicyJump
			jumpTo = mod.Fallback.JumpTo
		}

		m.fallback = policy
		m.fallbackJump = jumpTo
	}

	return m, nil
}

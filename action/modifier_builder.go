package action

import (
	"context"

	"github.com/Sn0wo2/CatSync/cli"
	"github.com/Sn0wo2/CatSync/config"
)

func BuildGlobalModifiers(cfg *config.Config) []Modifier {
	if cfg == nil {
		return nil
	}

	modifiers := make([]Modifier, 0, len(cfg.Modifiers))
	for i := range cfg.Modifiers {
		modifiers = append(modifiers, buildModifiers(&cfg.Modifiers[i])...)
	}

	return modifiers
}

func BuildActionModifiers(act *config.Action) []Modifier {
	if act == nil {
		return nil
	}

	return buildModifiers(&act.GlobalModifier)
}

func BuildPayloadModifiers(data config.ActionData) []Modifier {
	if data == nil {
		return nil
	}

	return buildModifiers(data.GetGlobalModifier())
}

func buildModifiers(gm *config.GlobalModifier) []Modifier {
	if gm == nil {
		return nil
	}

	modifiers := make([]Modifier, 0, 4)

	gm.EachModifier(func(m any) {
		switch mod := m.(type) {
		case *config.ActionModifierStatus:
			upstream, _ := mod.Upstream.ReadString(context.Background())
			modifiers = append(modifiers, NewStatusModifier().WithStatus(mod.Status).WithUpstream(upstream))
		case *config.ActionModifierAuth:
			modifiers = append(modifiers, NewAuthModifier(*mod))
		case *config.ActionModifierResponseHeader:
			upstream, _ := mod.Upstream.ReadString(context.Background())
			modifiers = append(modifiers, NewResponseHeaderModifier().WithHeader(mod.Header).WithUpstream(upstream))
		case *config.ActionModifierVersion:
			placeholder, _ := mod.Placeholder.ReadString(context.Background())
			modifiers = append(modifiers, NewPlaceholderModifier().WithPlaceholder(placeholder).WithValue(cli.GetFormatVersion()))
		}
	})

	return modifiers
}

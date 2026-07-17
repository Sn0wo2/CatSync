package action

import (
	"context"

	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/version"
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

	gm.Visit(config.GlobalModifierVisitor{
		Status: func(status *config.ActionModifierStatus) {
			upstream, _ := status.Upstream.ReadString(context.Background())
			modifiers = append(modifiers, NewStatusModifier().WithStatus(status.Status).WithUpstream(upstream))
		},
		Auth: func(auth *config.ActionModifierAuth) {
			modifiers = append(modifiers, NewAuthModifier(*auth))
		},
		ResponseHeader: func(header *config.ActionModifierResponseHeader) {
			upstream, _ := header.Upstream.ReadString(context.Background())
			modifiers = append(modifiers, NewResponseHeaderModifier().WithHeader(header.Header).WithUpstream(upstream))
		},
		Version: func(modifier *config.ActionModifierVersion) {
			placeholder, _ := modifier.Placeholder.ReadString(context.Background())
			modifiers = append(modifiers, NewPlaceholderModifier().WithPlaceholder(placeholder).WithValue(version.GetFormatVersion()))
		},
	})

	return modifiers
}

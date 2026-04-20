package action

import (
	"context"

	"github.com/Sn0wo2/CatSync/config"
)

type ModifierBuilder struct{}

func NewModifierBuilder() *ModifierBuilder {
	return &ModifierBuilder{}
}

func (b *ModifierBuilder) BuildGlobal(cfg *config.Config) []Modifier {
	if cfg == nil {
		return nil
	}

	modifiers := make([]Modifier, 0, len(cfg.Modifiers))
	for i := range cfg.Modifiers {
		modifiers = append(modifiers, b.BuildFromGlobalModifier(&cfg.Modifiers[i])...)
	}

	return modifiers
}

func (b *ModifierBuilder) BuildAction(act *config.Action) []Modifier {
	if act == nil {
		return nil
	}

	return b.BuildFromGlobalModifier(&act.GlobalModifier)
}

func (b *ModifierBuilder) BuildPayload(data config.ActionData) []Modifier {
	if data == nil {
		return nil
	}

	return b.BuildFromGlobalModifier(data.GetGlobalModifier())
}

func (b *ModifierBuilder) BuildFromGlobalModifier(gm *config.GlobalModifier) []Modifier {
	if gm == nil {
		return nil
	}

	modifiers := make([]Modifier, 0, 3)
	if gm.ActionModifierStatus != nil {
		upstream, _ := gm.ActionModifierStatus.Upstream.ReadString(context.Background())
		modifiers = append(modifiers, NewStatusModifier().WithStatus(gm.Status).WithUpstream(upstream))
	}

	if gm.ActionModifierAuth != nil {
		modifiers = append(modifiers, NewAuthModifier(*gm.ActionModifierAuth))
	}

	if gm.ActionModifierResponseHeader != nil {
		upstream, _ := gm.ActionModifierResponseHeader.Upstream.ReadString(context.Background())
		modifiers = append(modifiers, NewResponseHeaderModifier().WithHeader(gm.ActionModifierResponseHeader.Header).WithUpStream(upstream))
	}

	if gm.ActionModifierVersion != nil {
		placeholder, _ := gm.Placeholder.ReadString(context.Background())
		modifiers = append(modifiers, NewPlaceholderModifier().WithPlaceholder(placeholder))
	}

	return modifiers
}

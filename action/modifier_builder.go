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
	switch v := data.(type) {
	case *config.ActionStringData:
		if v == nil {
			return nil
		}

		return b.BuildFromGlobalModifier(&v.GlobalModifier)
	case *config.ActionFileData:
		if v == nil {
			return nil
		}

		return b.BuildFromGlobalModifier(&v.GlobalModifier)
	case *config.ActionServerData:
		if v == nil {
			return nil
		}

		return b.BuildFromGlobalModifier(&v.GlobalModifier)
	case *config.ActionReloadData:
		if v == nil {
			return nil
		}

		return b.BuildFromGlobalModifier(&v.GlobalModifier)
	default:
		return nil
	}
}

func (b *ModifierBuilder) BuildFromGlobalModifier(gm *config.GlobalModifier) []Modifier {
	if gm == nil {
		return nil
	}

	modifiers := make([]Modifier, 0, 3)
	if gm.ActionModifierStatus != nil {
		modifiers = append(modifiers, NewStatusModifier().WithStatus(gm.ActionModifierStatus.Status))
	}

	if gm.ActionModifierAuth != nil {
		modifiers = append(modifiers, NewAuthModifier(*gm.ActionModifierAuth))
	}

	if gm.ActionModifierResponseHeader != nil {
		upstream, err := gm.ActionModifierResponseHeader.Upstream.ReadString(context.Background())
		if err != nil {
			return nil
		}
		modifiers = append(modifiers, NewResponseHeaderModifier().WithHeader(gm.ActionModifierResponseHeader.Header).WithUpStream(upstream))
	}

	return modifiers
}

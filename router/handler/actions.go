package handler

import (
	"fmt"

	"github.com/Sn0wo2/CatSync/action"
	"github.com/Sn0wo2/CatSync/action/execute"
	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/params"
	"github.com/gofiber/fiber/v2"
)

func buildGlobalModifiers(cfg *config.Config) []action.Modifier {
	var modifiers []action.Modifier

	if cfg == nil {
		return modifiers
	}

	for i := range cfg.Modifiers {
		modifiers = append(modifiers, buildModifiersFromGlobalModifier(&cfg.Modifiers[i])...)
	}

	return modifiers
}

func buildModifiersFromGlobalModifier(gm *config.GlobalModifier) []action.Modifier {
	var modifiers []action.Modifier
	if gm == nil {
		return modifiers
	}

	if gm.ActionModifierStatus != nil {
		modifiers = append(modifiers, action.NewStatusModifier().WithStatus(gm.Status))
	}

	if gm.ActionModifierAuth != nil {
		modifiers = append(modifiers, action.NewAuthModifier(*gm.ActionModifierAuth))
	}

	if gm.ActionModifierResponseHeader != nil {
		for k, v := range gm.ActionModifierResponseHeader.Header {
			modifiers = append(modifiers, action.NewResponseHeaderModifier(k, v...))
		}
	}

	return modifiers
}

func buildActionModifiers(act *config.Action) []action.Modifier {
	if act == nil {
		return nil
	}

	return buildModifiersFromGlobalModifier(&act.GlobalModifier)
}

func buildPayloadModifiers(data config.ActionData) []action.Modifier {
	switch v := data.(type) {
	case *config.ActionStringData:
		if v == nil {
			return nil
		}

		return buildModifiersFromGlobalModifier(&v.GlobalModifier)
	case *config.ActionFileData:
		if v == nil {
			return nil
		}

		return buildModifiersFromGlobalModifier(&v.GlobalModifier)
	case *config.ActionServerData:
		if v == nil {
			return nil
		}

		return buildModifiersFromGlobalModifier(&v.GlobalModifier)
	}

	return nil
}

func Actions(c *params.Ctx) fiber.Handler {
	pl, err := execute.Compile(c.GetConfig(), execute.Builders{Global: buildGlobalModifiers, Action: buildActionModifiers, Payload: buildPayloadModifiers})
	if err != nil {
		panic(err)
	}

	return func(ctx *fiber.Ctx) error {
		exec := pl.Runner(c, ctx)

		jumpVisited := map[int]bool{}
		forceIndex := -1

		end := len(c.GetConfig().Actions)
		for i := 0; i < end; i++ {
			exec.WithSkipRouteCheck(i == forceIndex)
			res, err := exec.ExecuteAt(i)
			forceIndex = -1

			if err != nil {
				return err
			}

			if res.NotMatched {
				continue
			}

			if res.JumpTo != nil {
				if *res.JumpTo < 0 || *res.JumpTo >= len(c.GetConfig().Actions) {
					return fmt.Errorf("invalid auth fallback jumpTo index: %d", *res.JumpTo)
				}

				if jumpVisited[*res.JumpTo] {
					return fmt.Errorf("auth fallback jump loop detected: %d", *res.JumpTo)
				}

				jumpVisited[*res.JumpTo] = true
				forceIndex = *res.JumpTo
				i = *res.JumpTo - 1

				continue
			}

			if res.Matched {
				return nil
			}
		}

		// Notfound convention: always execute the last action.
		if len(c.GetConfig().Actions) == 0 {
			return ctx.Next()
		}

		lastIndex := len(c.GetConfig().Actions) - 1

		exec.WithSkipRouteCheck(true)

		res, err := exec.ExecuteAt(lastIndex)
		if err != nil {
			return err
		}

		if res.Matched {
			return nil
		}

		return ctx.Next()
	}
}

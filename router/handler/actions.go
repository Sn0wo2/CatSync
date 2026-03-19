package handler

import (
	"fmt"

	"github.com/Sn0wo2/CatSync/action"
	"github.com/Sn0wo2/CatSync/action/execute"
	"github.com/Sn0wo2/CatSync/params"
	"github.com/gofiber/fiber/v3"
)

func Actions(c *params.Ctx) fiber.Handler {
	b := action.NewModifierBuilder()

	exec := execute.New().WithConfig(c.GetConfig()).WithBuilders(execute.Builders{Global: b.BuildGlobal, Action: b.BuildAction, Payload: b.BuildPayload})

	return func(ctx fiber.Ctx) error {
		exec.WithContext(c, ctx)

		jumpVisited := map[int]bool{}
		forceIndex := -1

		for i := 0; i < len(c.GetConfig().Actions); i++ {
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

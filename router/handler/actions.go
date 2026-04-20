package handler

import (
	"github.com/Sn0wo2/CatSync/action"
	"github.com/Sn0wo2/CatSync/action/execute"
	"github.com/Sn0wo2/CatSync/params"
	"github.com/gofiber/fiber/v3"
)

func Actions(c *params.Ctx) fiber.Handler {
	b := action.NewModifierBuilder()

	exec := execute.New().
		WithConfig(c.GetConfig()).
		WithBuilders(execute.Builders{Global: b.BuildGlobal, Action: b.BuildAction, Payload: b.BuildPayload}).
		Build()

	return func(ctx fiber.Ctx) error {
		matched, err := exec.Dispatch(&execute.RequestContext{Ctx: c, FiberCtx: ctx})
		if err != nil {
			return err
		}

		if matched {
			return nil
		}

		return ctx.Next()
	}
}

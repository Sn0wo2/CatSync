package handler

import (
	"github.com/Sn0wo2/CatSync/action"
	"github.com/Sn0wo2/CatSync/action/execute"
	"github.com/Sn0wo2/CatSync/params"
	"github.com/Sn0wo2/CatSync/runtime"
	"github.com/gofiber/fiber/v3"
)

type Runtime interface {
	action.Reloader
	Current() *runtime.Snapshot
}

func Actions(c *params.Ctx, manager Runtime) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		snapshot := manager.Current()
		if snapshot == nil || snapshot.Executor == nil {
			return fiber.NewError(fiber.StatusInternalServerError, "runtime snapshot unavailable")
		}

		matched, err := snapshot.Executor.Dispatch(&execute.RequestContext{
			Ctx:      c,
			FiberCtx: ctx,
			Reloader: manager,
		})
		if err != nil {
			return err
		}

		if matched {
			return nil
		}

		return ctx.Next()
	}
}

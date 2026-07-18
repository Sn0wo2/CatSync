package router

import (
	"github.com/Sn0wo2/CatSync/action"
	"github.com/Sn0wo2/CatSync/action/execute"
	"github.com/Sn0wo2/CatSync/internal/appctx"
	"github.com/Sn0wo2/CatSync/runtime"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

type Runtime interface {
	action.Reloader
	Current() *runtime.Snapshot
}

func Init(c *appctx.Ctx, manager Runtime) {
	server := c.FW
	if server == nil {
		panic("framework not found")
	}

	server.Use(recover.New())
	server.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))

	server.Use(Actions(c, manager))
}

func Actions(c *appctx.Ctx, manager Runtime) fiber.Handler {
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

package router

import (
	"github.com/Sn0wo2/CatSync/action/execute"
	"github.com/Sn0wo2/CatSync/runtime"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	recoverer "github.com/gofiber/fiber/v3/middleware/recover"
)

func Init(server *fiber.App, manager *runtime.Manager) {
	if server == nil {
		panic("framework not found")
	}

	server.Use(recoverer.New())
	server.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))

	server.Use(func(ctx fiber.Ctx) error {
		executor := manager.Executor.Load()
		if executor == nil {
			return fiber.NewError(fiber.StatusInternalServerError, "runtime snapshot unavailable")
		}

		matched, err := executor.Dispatch(&execute.RequestContext{
			Logger:   manager.Logger,
			FiberCtx: ctx,
			Reloader: manager,
		})
		if err != nil || matched {
			return err
		}

		return ctx.Next()
	})
}

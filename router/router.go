package router

import (
	"github.com/Sn0wo2/CatSync/params"
	"github.com/Sn0wo2/CatSync/router/handler"
	"github.com/Sn0wo2/CatSync/router/middleware"
	"github.com/Sn0wo2/CatSync/router/notfound"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func Init(c *params.Ctx) {
	app := c.GetFramework()
	if app == nil {
		panic("framework not found")
	}

	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}), cors.New(), middleware.Server(c))

	debug := app.Group("/v0")
	debug.Get("/error", handler.Error(c))

	api := app.Group("/v1")
	api.Get("/health", handler.Health(c))

	for _, a := range c.GetConfig().Actions {
		app.Get(a.Route, handler.Actions(c, a))
	}

	notfound.Init(c.GetLogger(), app)
}

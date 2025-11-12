package router

import (
	"github.com/Sn0wo2/CatSync/framework"
	"github.com/Sn0wo2/CatSync/router/handler"
	"github.com/Sn0wo2/CatSync/router/middleware"
	"github.com/Sn0wo2/CatSync/router/notfound"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func Init(app *framework.Context) {
	router := app.App

	router.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}), cors.New(), middleware.Server(app.Logger, app.Config))

	debug := router.Group("/v0")
	debug.Get("/error", handler.Error(app))

	api := router.Group("/v1")
	api.Get("/health", handler.Health(app))

	for _, a := range app.Config.Actions {
		router.Get(a.Route, handler.Actions(app, a))
	}

	notfound.Init(app.Logger, router)
}

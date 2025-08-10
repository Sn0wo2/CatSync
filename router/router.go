package router

import (
	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/router/handler"
	"github.com/Sn0wo2/CatSync/router/notfound"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func Init(router fiber.Router) {
	router.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}), cors.New())

	debug := router.Group("/v0")
	debug.Get("/error", handler.Error())

	api := router.Group("/v1")
	api.Get("/health", handler.Health())

	for _, a := range config.Instance.Actions {
		router.Get(a.Route, handler.Actions(a))
	}

	notfound.Init(router)
}

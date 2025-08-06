package router

import (
	"github.com/Sn0wo2/FileSync/config"
	handler2 "github.com/Sn0wo2/FileSync/router/handler"
	"github.com/Sn0wo2/FileSync/router/notfound"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func Init(router fiber.Router) {
	router.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}), cors.New())

	api := router.Group("/v1")
	api.Get("/health", handler2.Health())
	api.Get("/error", handler2.Error())

	for _, a := range config.Instance.Actions {
		router.Get(a.Route, handler2.Actions(a))
	}

	notfound.Init("route not found", router)
}

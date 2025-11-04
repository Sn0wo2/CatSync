package router

import (
	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/router/handler"
	"github.com/Sn0wo2/CatSync/router/middleware"
	"github.com/Sn0wo2/CatSync/router/notfound"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"go.uber.org/zap"
)

func Init(logger *zap.Logger, cfg *config.Config, router fiber.Router) {
	router.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}), cors.New(), middleware.Server(logger, cfg))

	debug := router.Group("/v0")
	debug.Get("/error", handler.Error(logger, cfg))

	api := router.Group("/v1")
	api.Get("/health", handler.Health(logger, cfg))

	for _, a := range cfg.Actions {
		router.Get(a.Route, handler.Actions(logger, cfg, a))
	}

	notfound.Init(logger, router)
}

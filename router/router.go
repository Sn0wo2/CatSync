package router

import (
	"github.com/Sn0wo2/CatSync/internal/appctx"
	"github.com/Sn0wo2/CatSync/router/handler"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

func Init(c *appctx.Ctx, manager handler.Runtime) {
	server := c.FW
	if server == nil {
		panic("framework not found")
	}

	server.Use(recover.New())
	server.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}), cors.New())

	server.Use(handler.Actions(c, manager))
}

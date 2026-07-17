package router

import (
	"github.com/Sn0wo2/CatSync/params"
	"github.com/Sn0wo2/CatSync/router/handler"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

func Init(c *params.Ctx, manager handler.Runtime) {
	server := c.GetFramework()
	if server == nil {
		panic("framework not found")
	}

	server.Use(recover.New())
	server.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}), cors.New())

	server.Get("*", handler.Actions(c, manager))
}

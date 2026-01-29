package router

import (
	"github.com/Sn0wo2/CatSync/params"
	"github.com/Sn0wo2/CatSync/router/handler"
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
	}), cors.New())

	app.Get("*", handler.Actions(c))
}

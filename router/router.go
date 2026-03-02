package router

import (
	"github.com/Sn0wo2/CatSync/params"
	"github.com/Sn0wo2/CatSync/router/handler"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

func Init(c *params.Ctx) {
	app := c.GetFramework()
	if app == nil {
		panic("framework not found")
	}

	app.Use(recover.New())
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}), cors.New())

	app.Get("*", handler.Actions(c))
}

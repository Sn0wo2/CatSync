package middleware

import (
	"github.com/Sn0wo2/CatSync/params"
	"github.com/gofiber/fiber/v2"
)

func Server(p *params.Ctx) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("Server", p.GetConfig().Server.Header)

		return c.Next()
	}
}

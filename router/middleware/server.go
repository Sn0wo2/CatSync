package middleware

import "github.com/gofiber/fiber/v2"

func Server() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("Server", "CatSync")
		return c.Next()
	}
}

package handler

import (
	"github.com/gofiber/fiber/v2"
)

func Health() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendString("OK")
	}
}

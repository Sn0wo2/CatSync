package middleware

import (
	"github.com/Sn0wo2/CatSync/config"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func Server(_ *zap.Logger, _ *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("Server", "CatSync")

		return c.Next()
	}
}

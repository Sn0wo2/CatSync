package middleware

import (
	"github.com/Sn0wo2/CatSync/config"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func Server(_ *zap.Logger, cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("Server", cfg.Server.Header)

		return c.Next()
	}
}

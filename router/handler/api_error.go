package handler

import (
	"errors"

	"github.com/Sn0wo2/CatSync/internal/util"
	"github.com/Sn0wo2/CatSync/params"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func Error(c *params.Ctx) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		c.GetLogger().Info("Router >> Error test", zap.String("ctx", util.FiberContextString(ctx)))

		return errors.New("test error")
	}
}

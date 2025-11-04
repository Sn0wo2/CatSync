package handler

import (
	"errors"

	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/internal/util"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func Error(logger *zap.Logger, _ *config.Config) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		logger.Info("Router >> Error test", zap.String("ctx", util.FiberContextString(ctx)))

		return errors.New("test error")
	}
}

package handler

import (
	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/internal/util"
	"github.com/Sn0wo2/CatSync/response"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func Health(logger *zap.Logger, _ *config.Config) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		logger.Info("Router >> Health", zap.String("ctx", util.FiberContextString(ctx)))

		return response.New("ok").Write(ctx)
	}
}

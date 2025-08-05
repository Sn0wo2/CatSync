package handler

import (
	"github.com/Sn0wo2/FileSync/internal/log"
	"github.com/Sn0wo2/FileSync/internal/response"
	"github.com/Sn0wo2/FileSync/internal/util"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func Health() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		log.Instance.Info("H >> Health", zap.String("ctx", util.FiberContextString(ctx)))
		return response.New("ok").Write(ctx)
	}
}

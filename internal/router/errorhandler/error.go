package errorhandler

import (
	"github.com/Sn0wo2/FileSync/internal/log"
	"github.com/Sn0wo2/FileSync/internal/response"
	"github.com/Sn0wo2/FileSync/internal/util"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func Error(ctx *fiber.Ctx, err error) error {
	log.Instance.Error("Error handler caught error",
		zap.Error(err),
		zap.String("ctx", util.FiberContextString(ctx)),
	)

	return ctx.Status(fiber.StatusInternalServerError).JSON(response.New("oops, something went wrong"))
}

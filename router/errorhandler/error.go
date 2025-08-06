package errorhandler

import (
	"github.com/Sn0wo2/CatSync/internal/util"
	"github.com/Sn0wo2/CatSync/log"
	"github.com/Sn0wo2/CatSync/response"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func Error(ctx *fiber.Ctx, err error) error {
	traceID := uuid.NewString()

	log.Instance.Error("EH >> Error handler caught error",
		zap.String("traceID", traceID),
		zap.Error(err),
		zap.String("ctx", util.FiberContextString(ctx)),
	)

	return response.New("oops, something went wrong", fiber.Map{"traceID": traceID}).Write(ctx, fiber.StatusInternalServerError)
}

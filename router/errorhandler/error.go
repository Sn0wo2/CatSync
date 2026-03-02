package errorhandler

import (
	"fmt"
	"runtime/debug"

	"github.com/Sn0wo2/CatSync/internal/util"
	"github.com/Sn0wo2/CatSync/response"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func Error(logger *zap.Logger) func(ctx fiber.Ctx, err error) error {
	return func(ctx fiber.Ctx, err error) error {
		traceID := uuid.NewString()

		stack := string(debug.Stack())
		logger.Error(fmt.Sprintf("EH >> %+v\n%s", err, stack),
			zap.String("traceID", traceID),
			zap.String("ctx", util.FiberContextString(ctx)),
		)

		return response.New("oops, something went wrong", fiber.Map{"traceID": traceID}).Write(ctx, fiber.StatusInternalServerError)
	}
}

package errorhandler

import (
	"fmt"
	"log/slog"

	"github.com/Sn0wo2/CatSync/internal/util"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type Response struct {
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
}

func Error(logger *slog.Logger) func(ctx fiber.Ctx, err error) error {
	return func(ctx fiber.Ctx, err error) error {
		traceID := uuid.NewString()

		logger.Error(fmt.Sprintf("EH >> %+v", err),
			"traceID", traceID,
			"stack", util.LazyStack(),
			"ctx", util.LazyFiberContext(ctx),
		)

		ctx.Status(fiber.StatusInternalServerError)

		return ctx.JSON(&Response{
			Msg:  "oops, something went wrong",
			Data: fiber.Map{"traceID": traceID},
		})
	}
}

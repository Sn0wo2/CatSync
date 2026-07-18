package errorhandler

import (
	"fmt"

	"github.com/Sn0wo2/CatSync/internal/util"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Response struct {
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
}

func New(msg string, data ...any) *Response {
	r := &Response{Msg: msg}
	if len(data) == 1 {
		r.Data = data[0]
	} else if len(data) > 1 {
		r.Data = data
	}

	return r
}

func (r *Response) Write(ctx fiber.Ctx, status ...int) error {
	if len(status) > 0 {
		ctx.Status(status[0])
	}

	return ctx.JSON(r)
}

func Error(logger *zap.Logger) func(ctx fiber.Ctx, err error) error {
	return func(ctx fiber.Ctx, err error) error {
		traceID := uuid.NewString()

		logger.Error(fmt.Sprintf("EH >> %+v", err),
			zap.String("traceID", traceID),
			zap.Stack("stack"),
			util.LazyFiberContext(ctx),
		)

		return New("oops, something went wrong", fiber.Map{"traceID": traceID}).Write(ctx, fiber.StatusInternalServerError)
	}
}

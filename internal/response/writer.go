package response

import (
	"github.com/gofiber/fiber/v2"
)

func (r *Response) Write(ctx *fiber.Ctx, status ...int) error {
	if len(status) > 0 {
		ctx.Status(status[0])
	}
	return ctx.JSON(r)
}

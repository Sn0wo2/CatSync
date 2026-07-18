package errorhandler

import "github.com/gofiber/fiber/v3"

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

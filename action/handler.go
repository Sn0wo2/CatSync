package action

import (
	"github.com/Sn0wo2/CatSync/params"
	"github.com/gofiber/fiber/v2"
)

type Handler interface {
	Execute(app *params.Ctx, ctx *fiber.Ctx, data any) error
}

package handler

import (
	"errors"

	"github.com/Sn0wo2/CatSync/framework"
	"github.com/Sn0wo2/CatSync/internal/util"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func Error(app *framework.Context) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		app.Logger.Info("Router >> Error test", zap.String("ctx", util.FiberContextString(ctx)))

		return errors.New("test error")
	}
}

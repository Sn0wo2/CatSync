package notfound

import (
	"strings"

	"github.com/Sn0wo2/FileSync/internal/log"
	"github.com/Sn0wo2/FileSync/internal/response"
	"github.com/Sn0wo2/FileSync/internal/util"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func Init(msg string, router fiber.Router) {
	if msg = strings.ToLower(msg); msg == "" {
		msg = "page not found"
	}

	router.Use("*", func(ctx *fiber.Ctx) error {
		log.Instance.Warn(util.TitleCase(msg),
			zap.String("ctx", util.FiberContextString(ctx)))

		return ctx.Status(fiber.StatusNotFound).JSON(response.New(msg))
	})
}

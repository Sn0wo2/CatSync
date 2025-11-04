package notfound

import (
	"strings"

	"github.com/Sn0wo2/CatSync/internal/util"
	"github.com/Sn0wo2/CatSync/response"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func Init(logger *zap.Logger, router fiber.Router, msg ...string) {
	m := strings.Join(msg, " ")
	if m = strings.ToLower(m); m == "" {
		m = "page not found"
	}

	router.Use("*", func(ctx *fiber.Ctx) error {
		logger.Warn("Router >> "+util.TitleCase(m),
			zap.String("ctx", util.FiberContextString(ctx)))

		return response.New(m).Write(ctx, fiber.StatusNotFound)
	})
}

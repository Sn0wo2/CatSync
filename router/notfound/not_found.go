package notfound

import (
	"strings"

	"github.com/Sn0wo2/CatSync/internal/util"
	"github.com/Sn0wo2/CatSync/log"
	"github.com/Sn0wo2/CatSync/response"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func Init(msg string, router fiber.Router) {
	if msg = strings.ToLower(msg); msg == "" {
		msg = "page not found"
	}

	router.Use("*", func(ctx *fiber.Ctx) error {
		log.Instance.Warn("NF >> "+util.TitleCase(msg),
			zap.String("ctx", util.FiberContextString(ctx)))

		return response.New(msg).Write(ctx, fiber.StatusNotFound)
	})
}

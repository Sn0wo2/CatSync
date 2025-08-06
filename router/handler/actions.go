package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Sn0wo2/CatSync/action"
	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/internal/util"
	"github.com/Sn0wo2/CatSync/log"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func Actions(act config.Action) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		if act.UA != "" {
			matched, err := regexp.Match(act.UA, ctx.Request().Header.UserAgent())
			if err != nil {
				return fmt.Errorf("failed to match user agent: %w", err)
			}

			if !matched {
				log.Instance.Info("A >> User agent not matched", zap.String("ua", act.UA), zap.String("ctx", util.FiberContextString(ctx)))

				return ctx.Next()
			}
		}

		switch act.Action {
		case action.File:
			safePath, err := filepath.Abs(act.ActionData)
			if err != nil {
				return fmt.Errorf("failed to get absolute path: %w", err)
			}

			wd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get working directory: %w", err)
			}

			if !strings.HasPrefix(safePath, filepath.Join(wd, "data")) {
				return fmt.Errorf("file path is not in data directory: %s", safePath)
			}

			fileBytes, err := os.ReadFile(safePath) //nolint:gosec
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			ctx.Set(fiber.HeaderContentType, http.DetectContentType(fileBytes))

			log.Instance.Info("A >> Serving file", zap.String("file", safePath), zap.String("ctx", util.FiberContextString(ctx)))

			return ctx.Send(fileBytes)
		case action.String:
			log.Instance.Info("A >> Serving string", zap.String("string", act.ActionData), zap.String("ctx", util.FiberContextString(ctx)))

			return ctx.SendString(act.ActionData)
		case action.URL302:
			log.Instance.Info("A >> Redirecting to URL", zap.String("url", act.ActionData), zap.String("ctx", util.FiberContextString(ctx)))

			return ctx.Status(fiber.StatusFound).Redirect(act.ActionData)
		}

		log.Instance.Info("A >> Unknown action", zap.Int("action", int(act.Action)), zap.String("ctx", util.FiberContextString(ctx)))

		return ctx.Next()
	}
}

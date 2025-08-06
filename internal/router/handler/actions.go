package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sn0wo2/FileSync/internal/action"
	"github.com/Sn0wo2/FileSync/internal/config"
	"github.com/Sn0wo2/FileSync/internal/log"
	"github.com/Sn0wo2/FileSync/internal/util"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func Actions(act config.Action) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
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

		return ctx.Next()
	}
}

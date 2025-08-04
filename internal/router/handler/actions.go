package handler

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sn0wo2/FileSync/internal/action"
	"github.com/Sn0wo2/FileSync/internal/config"
	"github.com/gofiber/fiber/v2"
)

func Actions(a config.Action) fiber.Handler {
	return func(c *fiber.Ctx) error {
		switch a.Action {
		case action.File:
			safePath, err := filepath.Abs(a.ActionData)
			if err != nil {
				return fiber.ErrInternalServerError
			}
			wd, err := os.Getwd()
			if err != nil {
				return fiber.ErrInternalServerError
			}
			if !strings.HasPrefix(safePath, wd) {
				return fiber.ErrForbidden
			}
			fileBytes, err := os.ReadFile(safePath)
			if err != nil {
				return fiber.ErrNotFound
			}
			c.Set(fiber.HeaderContentType, http.DetectContentType(fileBytes))
			return c.Send(fileBytes)
		case action.String:
			return c.SendString(a.ActionData)
		case action.URL302:
			return c.Status(fiber.StatusFound).Redirect(a.ActionData)
		}
		return c.Next()
	}
}

package action

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type Handler interface {
	Execute(ctx *fiber.Ctx, data Data, logger zap.Logger) error
}

var HandlerRegistry = make(map[Type]Handler)

func init() {
	HandlerRegistry[String] = &StringHandler{}
	HandlerRegistry[File] = &FileHandler{}
	HandlerRegistry[Redirect] = &RedirectHandler{Permanent: true}
	HandlerRegistry[TempRedirect] = &RedirectHandler{Permanent: false}
	HandlerRegistry[JSON] = &JSONHandler{}
}

type StringHandler struct{}

func (h *StringHandler) Execute(ctx *fiber.Ctx, data Data, logger zap.Logger) error {
	actionData, ok := data.(string)
	if !ok {
		return fmt.Errorf("invalid action data type for string action: %T", data)
	}
	return ctx.SendString(actionData)
}

type FileHandler struct{}

func (h *FileHandler) Execute(ctx *fiber.Ctx, data Data, logger zap.Logger) error {
	actionData, ok := data.(string)
	if !ok {
		return fmt.Errorf("invalid action data type for file action: %T", data)
	}

	safePath, err := filepath.Abs(actionData)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	if !filepath.HasPrefix(safePath, filepath.Join(wd, "data")) {
		return fmt.Errorf("file path is not in data directory: %s", safePath)
	}

	fileBytes, err := os.ReadFile(safePath) //nolint:gosec
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	ctx.Set(fiber.HeaderContentType, http.DetectContentType(fileBytes))
	return ctx.Send(fileBytes)
}

type RedirectHandler struct {
	Permanent bool
}

func (h *RedirectHandler) Execute(ctx *fiber.Ctx, data Data, logger zap.Logger) error {
	actionData, ok := data.(string)
	if !ok {
		return fmt.Errorf("invalid action data type for redirect action: %T", data)
	}

	status := fiber.StatusFound
	if h.Permanent {
		status = fiber.StatusMovedPermanently
	}

	return ctx.Status(status).Redirect(actionData)
}

type JSONHandler struct{}

func (h *JSONHandler) Execute(ctx *fiber.Ctx, data Data, logger zap.Logger) error {
	return ctx.JSON(data)
}

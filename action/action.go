package action

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sn0wo2/CatSync/internal/util"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type Handler interface {
	Execute(logger *zap.Logger, ctx *fiber.Ctx, data Data) error
}

var HandlerRegistry = make(map[Type]Handler)

func init() {
	HandlerRegistry[String] = NewString()
	HandlerRegistry[File] = NewFile()
	HandlerRegistry[TempRedirect] = NewTempRedirect()
	HandlerRegistry[Redirect] = NewRedirect()
	HandlerRegistry[JSON] = NewJSON()
}

type StringHandler struct{}

func NewString() *StringHandler {
	return &StringHandler{}
}

func (h *StringHandler) Execute(logger *zap.Logger, ctx *fiber.Ctx, data Data) error {
	actionData, ok := data.(string)
	if !ok {
		return fmt.Errorf("invalid action data type for string action: %T", data)
	}

	logger.Info("Action >> Serve String", zap.String("data", actionData), zap.String("ctx", util.FiberContextString(ctx)))

	return ctx.SendString(actionData)
}

type FileHandler struct{}

func NewFile() *FileHandler {
	return &FileHandler{}
}

func (h *FileHandler) Execute(logger *zap.Logger, ctx *fiber.Ctx, data Data) error {
	actionData, ok := data.(string)
	if !ok {
		return fmt.Errorf("invalid action data type for file action: %T", data)
	}

	safePath, err := filepath.Abs(filepath.Clean(actionData))
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	absDataDir, _ := filepath.Abs(filepath.Join(wd, "data"))

	dataDirWithSep := absDataDir + string(filepath.Separator)
	if safePath != absDataDir && !strings.HasPrefix(safePath, dataDirWithSep) {
		return fmt.Errorf("file path is not in data directory: %s", safePath)
	}

	fileInfo, err := os.Stat(safePath)
	if err != nil {
		return fmt.Errorf("failed to access file: %w", err)
	}

	if fileInfo.IsDir() {
		return fmt.Errorf("cannot read directory: %s", safePath)
	}

	fileBytes, err := os.ReadFile(filepath.Clean(safePath))
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	ctx.Set(fiber.HeaderContentType, http.DetectContentType(fileBytes))

	return ctx.Send(fileBytes)
}

type RedirectHandler struct {
	Permanent bool
}

func NewTempRedirect() *RedirectHandler {
	return &RedirectHandler{}
}

func NewRedirect() *RedirectHandler {
	return &RedirectHandler{
		Permanent: true,
	}
}

func (h *RedirectHandler) Execute(logger *zap.Logger, ctx *fiber.Ctx, data Data) error {
	actionData, ok := data.(string)
	if !ok {
		return fmt.Errorf("invalid action data type for redirect action: %T", data)
	}

	status := fiber.StatusFound
	if h.Permanent {
		status = fiber.StatusMovedPermanently
	}

	logger.Info("Action >> Redirect", zap.String("to", actionData), zap.Int("status", status), zap.String("ctx", util.FiberContextString(ctx)))

	return ctx.Status(status).Redirect(actionData)
}

type JSONHandler struct{}

func NewJSON() *JSONHandler {
	return &JSONHandler{}
}

func (h *JSONHandler) Execute(logger *zap.Logger, ctx *fiber.Ctx, data Data) error {
	logger.Info("Action >> Serve JSON", zap.Any("data", data), zap.String("ctx", util.FiberContextString(ctx)))

	return ctx.JSON(data)
}

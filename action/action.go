package action

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/framework"
	"github.com/Sn0wo2/CatSync/internal/util"
	"github.com/Sn0wo2/CatSync/version"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type Handler interface {
	Execute(app *framework.Context, ctx *fiber.Ctx, data Data) error
}

var HandlerRegistry = make(map[config.ActionOperation]Handler)

func init() {
	HandlerRegistry[config.OperationVersion] = NewVersion()
	HandlerRegistry[config.OperationReload] = NewReload()
	HandlerRegistry[config.OperationString] = NewString()
	HandlerRegistry[config.OperationFile] = NewFile()
	HandlerRegistry[config.OperationTempRedirect] = NewTempRedirect()
	HandlerRegistry[config.OperationRedirect] = NewRedirect()
	HandlerRegistry[config.OperationJSON] = NewJSON()
}

type VersionHandler struct{}

func NewVersion() *VersionHandler {
	return &VersionHandler{}
}

func (h *VersionHandler) Execute(app *framework.Context, ctx *fiber.Ctx, data Data) error {
	app.Logger.Info("Action >> Serve Version", zap.Any("version", version.GetFormatVersion()), zap.String("ctx", util.FiberContextString(ctx)))

	return ctx.JSON(util.ReplaceVersionInAny(data, version.GetFormatVersion()))
}

type ReloadHandler struct{}

func NewReload() *ReloadHandler {
	return &ReloadHandler{}
}

func (h *ReloadHandler) Execute(app *framework.Context, ctx *fiber.Ctx, data Data) error {
	app.Logger.Info("Action >> Reload requested", zap.String("ctx", util.FiberContextString(ctx)))

	go func() {
		if err := app.Config.Reload(); err != nil {
			app.Logger.Error("Action >> Config reload failed", zap.Error(err))
			return
		}
	}()

	return ctx.JSON(data)
}

type StringHandler struct{}

func NewString() *StringHandler {
	return &StringHandler{}
}

func (h *StringHandler) Execute(app *framework.Context, ctx *fiber.Ctx, data Data) error {
	str, ok := data.(string)
	if !ok {
		return fmt.Errorf("invalid action data type for string action: expected string, got %T", data)
	}
	app.Logger.Info("Action >> Serve String", zap.String("data", str), zap.String("ctx", util.FiberContextString(ctx)))

	return ctx.SendString(str)
}

type FileHandler struct{}

func NewFile() *FileHandler {
	return &FileHandler{}
}

func (h *FileHandler) Execute(app *framework.Context, ctx *fiber.Ctx, data Data) error {
	path, ok := data.(string)
	if !ok {
		return fmt.Errorf("invalid action data type for file action: expected string, got %T", data)
	}

	app.Logger.Info("Action >> Serve File", zap.String("path", path), zap.String("ctx", util.FiberContextString(ctx)))

	safePath, err := filepath.Abs(filepath.Clean(path))
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

func (h *RedirectHandler) Execute(app *framework.Context, ctx *fiber.Ctx, data Data) error {
	url, ok := data.(string)
	if !ok {
		return fmt.Errorf("invalid action data type for redirect action: expected string, got %T", data)
	}

	status := fiber.StatusFound
	if h.Permanent {
		status = fiber.StatusMovedPermanently
	}

	app.Logger.Info("Action >> Redirect", zap.String("to", url), zap.Int("status", status), zap.String("ctx", util.FiberContextString(ctx)))

	return ctx.Status(status).Redirect(url)
}

type JSONHandler struct{}

func NewJSON() *JSONHandler {
	return &JSONHandler{}
}

func (h *JSONHandler) Execute(app *framework.Context, ctx *fiber.Ctx, data Data) error {
	app.Logger.Info("Action >> Serve JSON", zap.Any("data", data), zap.String("ctx", util.FiberContextString(ctx)))

	return ctx.JSON(data)
}

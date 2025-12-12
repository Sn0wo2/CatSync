package action

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/internal/util"
	"github.com/Sn0wo2/CatSync/params"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func init() {
	HandlerRegistry[config.OperationString] = NewString()
	HandlerRegistry[config.OperationFile] = NewFile()
}

type StringHandler struct{}

func NewString() Handler {
	return &StringHandler{}
}

func (h *StringHandler) Execute(c *params.Ctx, ctx *fiber.Ctx, data any) error {
	stringData, ok := data.(*config.StringData)
	if !ok {
		return fmt.Errorf("invalid action data type for string action: expected *StringData, got %T", data)
	}

	c.GetLogger().Info("Action >> Serve String", zap.String("data", stringData.Content), zap.String("ctx", util.FiberContextString(ctx)))

	return ctx.SendString(stringData.Content)
}

type FileHandler struct{}

func NewFile() Handler {
	return &FileHandler{}
}

func (h *FileHandler) Execute(c *params.Ctx, ctx *fiber.Ctx, data any) error {
	fileData, ok := data.(*config.FileData)
	if !ok {
		return fmt.Errorf("invalid action data type for file action: expected *config.FileData, got %T", data)
	}

	c.GetLogger().Info("Action >> Serve File", zap.String("path", fileData.Path), zap.String("ctx", util.FiberContextString(ctx)))

	safePath, err := filepath.Abs(filepath.Clean(fileData.Path))
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

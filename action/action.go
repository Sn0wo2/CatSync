package action

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/internal/util"
	"github.com/Sn0wo2/CatSync/version"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func init() {
	HandlerRegistry[config.ActionString] = NewString()
	HandlerRegistry[config.ActionFile] = NewFile()
}

type StringHandler struct {
	BaseHandler
}

func NewString() Handler {
	return &StringHandler{}
}

func (h *StringHandler) ProcessAction(p *ProcessData) error {
	stringData, ok := p.PayLoad.(*config.ActionStringData)
	if !ok {
		return fmt.Errorf("invalid action data type for string action: expected *ActionStringData, got %T", p.PayLoad)
	}

	if stringData.ActionVersionModifier.Placeholder != "" {
		p = NewVersionModifier(version.GetFormatVersion()).ProcessModifier(h).HookProcessData()(p)
	}

	p.Ctx.GetLogger().Info("Action >> Serve String", zap.String("data", stringData.Content), zap.String("ctx", util.FiberContextString(p.C)))

	return p.C.SendString(stringData.Content)
}

type FileHandler struct {
	BaseHandler
}

func NewFile() Handler {
	return &FileHandler{}
}

func (h *FileHandler) ProcessAction(p *ProcessData) error {
	fileData, ok := p.PayLoad.(*config.ActionFileData)
	if !ok {
		return fmt.Errorf("invalid action data type for file action: expected *config.ActionFileData, got %T", p.PayLoad)
	}

	if fileData.ActionVersionModifier.Placeholder != "" {
		p = NewVersionModifier(version.GetFormatVersion()).ProcessModifier(h).HookProcessData()(p)
	}

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

	p.C.Set(fiber.HeaderContentType, http.DetectContentType(fileBytes))

	p.Ctx.GetLogger().Info("Action >> Serve File", zap.String("path", fileData.Path), zap.String("ctx", util.FiberContextString(p.C)))

	return p.C.Send(fileBytes)
}

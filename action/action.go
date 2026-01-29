package action

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/internal/util"
	"github.com/Sn0wo2/go-common/helper"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func init() {
	HandlerRegistry[config.ActionString] = NewString()
	HandlerRegistry[config.ActionFile] = NewFile()
}

type StringHandler struct{}

func NewString() Handler {
	return &StringHandler{}
}

func (h *StringHandler) ProcessAction(p *ProcessData) error {
	stringData, ok := (*p.PayLoad).(*config.ActionStringData)
	if !ok {
		return fmt.Errorf("invalid action data type for string action: expected *ActionStringData, got %T", p.PayLoad)
	}

	p.Ctx.GetLogger().Info("Action >> Serve String", zap.String("data", stringData.Content), zap.String("ctx", util.FiberContextString(p.C)))

	return p.C.SendString(stringData.Content)
}

type FileHandler struct{}

func NewFile() Handler {
	return &FileHandler{}
}

func (h *FileHandler) ProcessAction(p *ProcessData) error {
	fileData, ok := (*p.PayLoad).(*config.ActionFileData)
	if !ok {
		return fmt.Errorf("invalid action data type for file action: expected *config.ActionFileData, got %T", p.PayLoad)
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
		if errors.Is(err, os.ErrNotExist) {
			if err := os.MkdirAll(filepath.Dir(safePath), 0o750); err != nil {
				return fmt.Errorf("failed to create data directory: %w", err)
			}
			if err := os.WriteFile(filepath.Clean(safePath), helper.StringToBytes("CatSync!\n"), 0o600); err != nil {
				return fmt.Errorf("failed to create default file: %w", err)
			}
			fileInfo, err = os.Stat(safePath)
			if err != nil {
				return fmt.Errorf("failed to access file after create: %w", err)
			}
		} else {
			return fmt.Errorf("failed to access file: %w", err)
		}
	}

	if fileInfo.IsDir() {
		return fmt.Errorf("cannot read directory: %s", safePath)
	}

	if fileInfo.Size() == 0 {
		if err := os.WriteFile(filepath.Clean(safePath), []byte("CatSync!\n"), 0o600); err != nil {
			return fmt.Errorf("failed to write default file content: %w", err)
		}
	}

	fileBytes, err := os.ReadFile(filepath.Clean(safePath))
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	if !fileData.DontSetContentType {
		p.C.Set(fiber.HeaderContentType, http.DetectContentType(fileBytes))
	}

	p.Ctx.GetLogger().Info("Action >> Serve File", zap.String("path", fileData.Path), zap.String("ctx", util.FiberContextString(p.C)))

	return p.C.Send(fileBytes)
}

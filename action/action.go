package action

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/config/loader"
	"github.com/Sn0wo2/CatSync/config/reader"
	"github.com/Sn0wo2/CatSync/internal/filecache"
	"github.com/Sn0wo2/CatSync/internal/util"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

func init() {
	HandlerRegistry[config.ActionString] = NewString()
	HandlerRegistry[config.ActionFile] = NewFile()
	HandlerRegistry[config.ActionServer] = NewServer()
	HandlerRegistry[config.ActionReload] = NewReload()
}

type StringHandler struct{}

func NewString() Handler {
	return &StringHandler{}
}

func (h *StringHandler) ProcessAction(p *ProcessData) error {
	stringData, ok := p.Payload.(*config.ActionStringData)
	if !ok {
		return fmt.Errorf("invalid action data type for string action: expected *ActionStringData, got %T", p.Payload)
	}

	body := stringData.Content.Must()

	p.Ctx.GetLogger().Info("Action >> Serve String", util.LazyFiberContext(p.C))

	return p.C.SendString(body)
}

var fileCache = filecache.New(1 * time.Second)

type FileHandler struct{}

func NewFile() Handler {
	return &FileHandler{}
}

func (h *FileHandler) ProcessAction(p *ProcessData) error {
	fileData, ok := p.Payload.(*config.ActionFileData)
	if !ok {
		return fmt.Errorf("invalid action data type for file action: expected *config.ActionFileData, got %T", p.Payload)
	}

	pathStr := fileData.Path.Must()

	safePath, err := filepath.Abs(filepath.Clean(pathStr))
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

	if err := h.ensureFile(safePath); err != nil {
		return err
	}

	entry, err := fileCache.Get(safePath, !fileData.DontSetContentType)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	if !fileData.DontSetContentType && entry.ContentType != "" {
		p.C.Set(fiber.HeaderContentType, entry.ContentType)
	}

	p.Ctx.GetLogger().Info("Action >> Serve File", zap.String("path", pathStr), util.LazyFiberContext(p.C))

	return p.C.Send(entry.Content)
}

func (h *FileHandler) ensureFile(safePath string) error {
	info, err := os.Stat(safePath)
	if err == nil {
		if info.IsDir() {
			return fmt.Errorf("cannot read directory: %s", safePath)
		}

		if info.Size() == 0 {
			if err := os.WriteFile(filepath.Clean(safePath), []byte("CatSync!\n"), 0o600); err != nil {
				return fmt.Errorf("failed to write default file content: %w", err)
			}
		}

		return nil
	}

	if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to access file: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(safePath), 0o750); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	if err := os.WriteFile(filepath.Clean(safePath), []byte("CatSync!\n"), 0o600); err != nil {
		return fmt.Errorf("failed to create default file: %w", err)
	}

	return nil
}

type ServerHandler struct{}

func NewServer() Handler {
	return &ServerHandler{}
}

func (h *ServerHandler) sendErrorPage(p *ProcessData, serverData *config.ActionServerData, status int, dir, file string) error {
	p.C.Set(fiber.HeaderContentType, "text/html; charset=utf-8")
	p.C.Status(status)
	p.Ctx.GetLogger().Info("Action >> Serve Server", zap.String("dir", dir), zap.String("file", file), zap.String("status", strconv.Itoa(status)), util.LazyFiberContext(p.C))

	html := config.DefaultNotFoundHTML
	if serverData.NotFoundHTML != nil {
		html = serverData.NotFoundHTML.Must()
	}

	return p.C.SendString(html)
}

func (h *ServerHandler) findIndexFile(fullPath string, indexFiles []*reader.String) string {
	if len(indexFiles) == 0 {
		indexFiles = []*reader.String{
			reader.Str("index.html"),
			reader.Str("index.htm"),
		}
	}

	for _, sf := range indexFiles {
		if sf == nil {
			continue
		}

		indexPath := filepath.Join(fullPath, sf.Must())
		if _, err := os.Stat(indexPath); err == nil {
			return indexPath
		}
	}

	return ""
}

func (h *ServerHandler) ProcessAction(p *ProcessData) error {
	serverData, ok := p.Payload.(*config.ActionServerData)
	if !ok {
		return fmt.Errorf("invalid action data type for server action: expected *config.ActionServerData, got %T", p.Payload)
	}

	dirStr := serverData.Directory.Must()
	if dirStr == "" {
		return errors.New("directory is required for server action")
	}

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	baseDir := filepath.Join(wd, "data", dirStr)

	absBaseDir, err := filepath.Abs(filepath.Clean(baseDir))
	if err != nil {
		return fmt.Errorf("failed to get absolute base directory: %w", err)
	}

	absBaseDirSep := absBaseDir + string(filepath.Separator)
	absServerDir, _ := filepath.Abs(filepath.Join(wd, "data", "server"))

	serverDirWithSep := absServerDir + string(filepath.Separator)
	if absBaseDir == absServerDir {
		return fmt.Errorf("directory cannot be server itself, must be a subdirectory under data/server: %s", dirStr)
	}

	if !strings.HasPrefix(absBaseDir, serverDirWithSep) {
		return fmt.Errorf("directory must be under data/server: %s", dirStr)
	}

	reqPath := p.C.Path()
	if strings.Contains(reqPath, "..") {
		return fiber.NewError(fiber.StatusForbidden, "path traversal not allowed")
	}

	reqPath = strings.TrimPrefix(reqPath, "/")

	fullPath := filepath.Clean(filepath.Join(absBaseDir, reqPath))
	if !strings.HasPrefix(fullPath, absBaseDirSep) && fullPath != absBaseDir {
		return fiber.NewError(fiber.StatusForbidden, "access denied")
	}

	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return h.sendErrorPage(p, serverData, fiber.StatusNotFound, dirStr, reqPath)
		}

		return fmt.Errorf("failed to access file: %w", err)
	}

	if fileInfo.IsDir() {
		foundPath := h.findIndexFile(fullPath, serverData.IndexFiles)
		if foundPath == "" {
			return h.sendErrorPage(p, serverData, fiber.StatusForbidden, dirStr, reqPath)
		}

		fullPath = foundPath
	}

	p.Ctx.GetLogger().Info("Action >> Serve Server", zap.String("dir", dirStr), zap.String("file", reqPath), util.LazyFiberContext(p.C))

	return p.C.SendFile(fullPath)
}

type ReloadHandler struct{}

func NewReload() Handler {
	return &ReloadHandler{}
}

func (h *ReloadHandler) ProcessAction(p *ProcessData) error {
	_, ok := p.Payload.(*config.ActionReloadData)
	if !ok {
		return fmt.Errorf("invalid action data type for reload action: expected *config.ActionReloadData, got %T", p.Payload)
	}

	cfg := config.GetCurrentConfig()
	if cfg == nil {
		return errors.New("no config loaded")
	}

	err := cfg.Reload(loader.NewYAMLLoader(), loader.NewJSONLoader())
	if err != nil {
		p.Ctx.GetLogger().Error("Action >> Reload Config Failed", zap.Error(err), util.LazyFiberContext(p.C))
		p.C.Status(fiber.StatusInternalServerError)

		return p.C.SendString(fmt.Sprintf("Config reload failed: %v", err))
	}

	p.Ctx.GetLogger().Info("Action >> Reload Config Success", util.LazyFiberContext(p.C))
	p.C.Status(fiber.StatusOK)

	return p.C.SendString("Config reloaded successfully")
}

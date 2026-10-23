package action

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/config/reader"
	"github.com/Sn0wo2/CatSync/internal/filecache"
	"github.com/Sn0wo2/CatSync/internal/util"
	"github.com/gofiber/fiber/v3"
)

type ProcessData struct {
	Logger   *slog.Logger
	FCtx     fiber.Ctx
	Payload  config.ActionData
	Reloader Reloader
}

type Reloader interface {
	Reload() error
}

type Handler interface {
	ProcessAction(data *ProcessData) ExecutionResult
}

type ExecutionStatus uint8

const (
	ExecutionCompleted    ExecutionStatus = 0
	ExecutionFallbackNext ExecutionStatus = 2
	ExecutionFallbackJump ExecutionStatus = 3
)

type ExecutionResult struct {
	Status ExecutionStatus
	Err    error
	JumpTo int
}

func NewRegistry() map[config.ActionType]Handler {
	return map[config.ActionType]Handler{
		config.ActionString: &StringHandler{},
		config.ActionFile:   &FileHandler{cache: filecache.New(time.Second)},
		config.ActionServer: &ServerHandler{},
		config.ActionReload: &ReloadHandler{},
	}
}

type Modifier interface {
	Before(data *ProcessData) (*ProcessData, ExecutionResult)
	After(data *ProcessData) (*ProcessData, ExecutionResult)
}

type ModifiableHandler struct {
	Handler
	Modifiers []Modifier
}

func (h *ModifiableHandler) ProcessAction(data *ProcessData) ExecutionResult {
	for i := len(h.Modifiers) - 1; i >= 0; i-- {
		var result ExecutionResult

		data, result = h.Modifiers[i].Before(data)
		if result.Err != nil || result.Status != ExecutionCompleted {
			return result
		}
	}

	if result := h.Handler.ProcessAction(data); result.Err != nil || result.Status != ExecutionCompleted {
		return result
	}

	for _, modifier := range h.Modifiers {
		var result ExecutionResult

		data, result = modifier.After(data)
		if result.Err != nil || result.Status != ExecutionCompleted {
			return result
		}
	}

	return ExecutionResult{}
}

type StringHandler struct{}

func (h *StringHandler) ProcessAction(p *ProcessData) ExecutionResult {
	stringData, ok := p.Payload.(*config.ActionStringData)
	if !ok {
		return ExecutionResult{Err: fmt.Errorf("invalid action data type for string action: expected *ActionStringData, got %T", p.Payload)}
	}

	body := stringData.Content.Must()

	p.Logger.Info("Action >> Serve String", "ctx", util.LazyFiberContext{Ctx: p.FCtx})

	return ExecutionResult{Err: p.FCtx.SendString(body)}
}

type FileHandler struct {
	cache *filecache.Cache
}

func (h *FileHandler) Stop() {
	h.cache.Stop()
}

func (h *FileHandler) ProcessAction(p *ProcessData) ExecutionResult {
	fileData, ok := p.Payload.(*config.ActionFileData)
	if !ok {
		return ExecutionResult{Err: fmt.Errorf("invalid action data type for file action: expected *config.ActionFileData, got %T", p.Payload)}
	}

	pathStr := fileData.Path.Must()

	safePath, err := filepath.Abs(pathStr)
	if err != nil {
		return ExecutionResult{Err: fmt.Errorf("failed to get absolute path: %w", err)}
	}

	wd, err := os.Getwd()
	if err != nil {
		return ExecutionResult{Err: fmt.Errorf("failed to get working directory: %w", err)}
	}

	absDataDir, _ := filepath.Abs(filepath.Join(wd, "data"))

	dataDirWithSep := absDataDir + string(filepath.Separator)
	if safePath != absDataDir && !strings.HasPrefix(safePath, dataDirWithSep) {
		return ExecutionResult{Err: fmt.Errorf("file path is not in data directory: %s", safePath)}
	}

	info, err := os.Stat(safePath)
	switch {
	case errors.Is(err, os.ErrNotExist):
		if err := os.MkdirAll(filepath.Dir(safePath), 0o750); err != nil {
			return ExecutionResult{Err: fmt.Errorf("failed to create data directory: %w", err)}
		}

		if err := os.WriteFile(safePath, []byte("CatSync!\n"), 0o600); err != nil {
			return ExecutionResult{Err: fmt.Errorf("failed to create default file: %w", err)}
		}
	case err != nil:
		return ExecutionResult{Err: fmt.Errorf("failed to access file: %w", err)}
	case info.IsDir():
		return ExecutionResult{Err: fmt.Errorf("cannot read directory: %s", safePath)}
	case info.Size() == 0:
		if err := os.WriteFile(safePath, []byte("CatSync!\n"), 0o600); err != nil {
			return ExecutionResult{Err: fmt.Errorf("failed to write default file content: %w", err)}
		}
	}

	entry, err := h.cache.Get(safePath, !fileData.DontSetContentType)
	if err != nil {
		return ExecutionResult{Err: fmt.Errorf("failed to read file: %w", err)}
	}

	if !fileData.DontSetContentType && entry.ContentType != "" {
		p.FCtx.Set(fiber.HeaderContentType, entry.ContentType)
	}

	p.Logger.Info("Action >> Serve File", "path", pathStr, "ctx", util.LazyFiberContext{Ctx: p.FCtx})

	return ExecutionResult{Err: p.FCtx.Send(entry.Content)}
}

type ServerHandler struct{}

func (h *ServerHandler) ProcessAction(p *ProcessData) ExecutionResult {
	serverData, ok := p.Payload.(*config.ActionServerData)
	if !ok {
		return ExecutionResult{Err: fmt.Errorf("invalid action data type for server action: expected *config.ActionServerData, got %T", p.Payload)}
	}

	dirStr := serverData.Directory.Must()
	if dirStr == "" {
		return ExecutionResult{Err: errors.New("directory is required for server action")}
	}

	wd, err := os.Getwd()
	if err != nil {
		return ExecutionResult{Err: fmt.Errorf("failed to get working directory: %w", err)}
	}

	baseDir := filepath.Join(wd, "data", dirStr)

	absBaseDir, err := filepath.Abs(baseDir)
	if err != nil {
		return ExecutionResult{Err: fmt.Errorf("failed to get absolute base directory: %w", err)}
	}

	absBaseDirSep := absBaseDir + string(filepath.Separator)
	absServerDir, _ := filepath.Abs(filepath.Join(wd, "data", "server"))

	serverDirWithSep := absServerDir + string(filepath.Separator)
	if absBaseDir == absServerDir {
		return ExecutionResult{Err: fmt.Errorf("directory cannot be server itself, must be a subdirectory under data/server: %s", dirStr)}
	}

	if !strings.HasPrefix(absBaseDir, serverDirWithSep) {
		return ExecutionResult{Err: fmt.Errorf("directory must be under data/server: %s", dirStr)}
	}

	reqPath := strings.TrimPrefix(p.FCtx.Path(), "/")

	fullPath := filepath.Join(absBaseDir, reqPath)
	if !strings.HasPrefix(fullPath, absBaseDirSep) && fullPath != absBaseDir {
		return ExecutionResult{Err: fiber.NewError(fiber.StatusForbidden, "access denied")}
	}

	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ExecutionResult{Err: h.sendErrorPage(p, serverData, fiber.StatusNotFound, dirStr, reqPath)}
		}

		return ExecutionResult{Err: fmt.Errorf("failed to access file: %w", err)}
	}

	if fileInfo.IsDir() {
		indexFiles := serverData.IndexFiles
		if len(indexFiles) == 0 {
			indexFiles = []*reader.String{{Content: "index.html"}, {Content: "index.htm"}}
		}

		found := false

		for _, sf := range indexFiles {
			if sf == nil {
				continue
			}

			indexPath := filepath.Join(fullPath, sf.Must())
			if _, err := os.Stat(indexPath); err == nil {
				fullPath = indexPath
				found = true

				break
			}
		}

		if !found {
			return ExecutionResult{Err: h.sendErrorPage(p, serverData, fiber.StatusForbidden, dirStr, reqPath)}
		}
	}

	p.Logger.Info("Action >> Serve Server", "dir", dirStr, "file", reqPath, "ctx", util.LazyFiberContext{Ctx: p.FCtx})

	return ExecutionResult{Err: p.FCtx.SendFile(fullPath)}
}

func (h *ServerHandler) sendErrorPage(p *ProcessData, serverData *config.ActionServerData, status int, dir, file string) error {
	p.FCtx.Set(fiber.HeaderContentType, "text/html; charset=utf-8")
	p.FCtx.Status(status)
	p.Logger.Info("Action >> Serve Server", "dir", dir, "file", file, "status", strconv.Itoa(status), "ctx", util.LazyFiberContext{Ctx: p.FCtx})

	html := config.DefaultNotFoundHTML
	if serverData.NotFoundHTML != nil {
		html = serverData.NotFoundHTML.Must()
	}

	return p.FCtx.SendString(html)
}

type ReloadHandler struct{}

func (h *ReloadHandler) ProcessAction(p *ProcessData) ExecutionResult {
	_, ok := p.Payload.(*config.ActionReloadData)
	if !ok {
		return ExecutionResult{Err: fmt.Errorf("invalid action data type for reload action: expected *config.ActionReloadData, got %T", p.Payload)}
	}

	if p.Reloader == nil {
		return ExecutionResult{Err: errors.New("runtime reloader is unavailable")}
	}

	if err := p.Reloader.Reload(); err != nil {
		p.Logger.Error("Action >> Reload Config Failed", "error", err, "stack", util.LazyStack{}, "ctx", util.LazyFiberContext{Ctx: p.FCtx})
		p.FCtx.Status(fiber.StatusInternalServerError)

		return ExecutionResult{Err: p.FCtx.SendString(fmt.Sprintf("Config reload failed: %v", err))}
	}

	p.Logger.Info("Action >> Reload Config Success", "ctx", util.LazyFiberContext{Ctx: p.FCtx})
	p.FCtx.Status(fiber.StatusOK)

	return ExecutionResult{Err: p.FCtx.SendString("Config reloaded successfully")}
}

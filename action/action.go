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
	"github.com/Sn0wo2/CatSync/config/reader"
	"github.com/Sn0wo2/CatSync/internal/cstx"
	"github.com/Sn0wo2/CatSync/internal/filecache"
	"github.com/Sn0wo2/CatSync/internal/util"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

type ProcessData struct {
	CStx     *cstx.Ctx
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

type Registry map[config.ActionType]Handler

func NewRegistry() Registry {
	return Registry{
		config.ActionString: NewString(),
		config.ActionFile:   NewFile(),
		config.ActionServer: NewServer(),
		config.ActionReload: NewReload(),
	}
}

type Modifier interface {
	Before(data *ProcessData) (*ProcessData, ExecutionResult)
	After(data *ProcessData) (*ProcessData, ExecutionResult)
}

type ModifiableHandler struct {
	Handler
	modifiers []Modifier
}

func WrapHandler(h Handler) *ModifiableHandler {
	return &ModifiableHandler{Handler: h}
}

func (h *ModifiableHandler) WithModifier(m Modifier) *ModifiableHandler {
	h.modifiers = append(h.modifiers, m)

	return h
}

func (h *ModifiableHandler) ProcessAction(data *ProcessData) ExecutionResult {
	for i := len(h.modifiers) - 1; i >= 0; i-- {
		var result ExecutionResult

		data, result = h.modifiers[i].Before(data)
		if result.Err != nil || result.Status != ExecutionCompleted {
			return result
		}
	}

	if result := h.Handler.ProcessAction(data); result.Err != nil || result.Status != ExecutionCompleted {
		return result
	}

	for _, modifier := range h.modifiers {
		var result ExecutionResult

		data, result = modifier.After(data)
		if result.Err != nil || result.Status != ExecutionCompleted {
			return result
		}
	}

	return ExecutionResult{}
}

type StringHandler struct{}

func NewString() Handler {
	return &StringHandler{}
}

func (h *StringHandler) ProcessAction(p *ProcessData) ExecutionResult {
	stringData, ok := p.Payload.(*config.ActionStringData)
	if !ok {
		return ExecutionResult{Err: fmt.Errorf("invalid action data type for string action: expected *ActionStringData, got %T", p.Payload)}
	}

	body := stringData.Content.Must()

	p.CStx.Logger.Info("Action >> Serve String", util.LazyFiberContext(p.FCtx))

	return ExecutionResult{Err: p.FCtx.SendString(body)}
}

type FileHandler struct {
	cache *filecache.Cache
}

func NewFile() Handler {
	return &FileHandler{
		cache: filecache.New(time.Second),
	}
}

func (h *FileHandler) ProcessAction(p *ProcessData) ExecutionResult {
	fileData, ok := p.Payload.(*config.ActionFileData)
	if !ok {
		return ExecutionResult{Err: fmt.Errorf("invalid action data type for file action: expected *config.ActionFileData, got %T", p.Payload)}
	}

	pathStr := fileData.Path.Must()

	safePath, err := filepath.Abs(filepath.Clean(pathStr))
	if err != nil {
		return ExecutionResult{Err: fmt.Errorf("failed to get absolute path: %w", err)}
	}

	wd, err := os.Getwd()
	if err != nil {
		return ExecutionResult{Err: fmt.Errorf("failed to get working directory: %w", err)}
	}

	absDataDir, _ := filepath.Abs(filepath.Join(wd, ".data"))

	dataDirWithSep := absDataDir + string(filepath.Separator)
	if safePath != absDataDir && !strings.HasPrefix(safePath, dataDirWithSep) {
		return ExecutionResult{Err: fmt.Errorf("file path is not in data directory: %s", safePath)}
	}

	if err := h.ensureFile(safePath); err != nil {
		return ExecutionResult{Err: err}
	}

	entry, err := h.cache.Get(safePath, !fileData.DontSetContentType)
	if err != nil {
		return ExecutionResult{Err: fmt.Errorf("failed to read file: %w", err)}
	}

	if !fileData.DontSetContentType && entry.ContentType != "" {
		p.FCtx.Set(fiber.HeaderContentType, entry.ContentType)
	}

	p.CStx.Logger.Info("Action >> Serve File", zap.String("path", pathStr), util.LazyFiberContext(p.FCtx))

	return ExecutionResult{Err: p.FCtx.Send(entry.Content)}
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

	baseDir := filepath.Join(wd, ".data", dirStr)

	absBaseDir, err := filepath.Abs(filepath.Clean(baseDir))
	if err != nil {
		return ExecutionResult{Err: fmt.Errorf("failed to get absolute base directory: %w", err)}
	}

	absBaseDirSep := absBaseDir + string(filepath.Separator)
	absServerDir, _ := filepath.Abs(filepath.Join(wd, ".data", "server"))

	serverDirWithSep := absServerDir + string(filepath.Separator)
	if absBaseDir == absServerDir {
		return ExecutionResult{Err: fmt.Errorf("directory cannot be server itself, must be a subdirectory under .data/server: %s", dirStr)}
	}

	if !strings.HasPrefix(absBaseDir, serverDirWithSep) {
		return ExecutionResult{Err: fmt.Errorf("directory must be under data/server: %s", dirStr)}
	}

	reqPath := strings.TrimPrefix(p.FCtx.Path(), "/")

	fullPath := filepath.Clean(filepath.Join(absBaseDir, reqPath))
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
		foundPath := h.findIndexFile(fullPath, serverData.IndexFiles)
		if foundPath == "" {
			return ExecutionResult{Err: h.sendErrorPage(p, serverData, fiber.StatusForbidden, dirStr, reqPath)}
		}

		fullPath = foundPath
	}

	p.CStx.Logger.Info("Action >> Serve Server", zap.String("dir", dirStr), zap.String("file", reqPath), util.LazyFiberContext(p.FCtx))

	return ExecutionResult{Err: p.FCtx.SendFile(fullPath)}
}

func (h *ServerHandler) sendErrorPage(p *ProcessData, serverData *config.ActionServerData, status int, dir, file string) error {
	p.FCtx.Set(fiber.HeaderContentType, "text/html; charset=utf-8")
	p.FCtx.Status(status)
	p.CStx.Logger.Info("Action >> Serve Server", zap.String("dir", dir), zap.String("file", file), zap.String("status", strconv.Itoa(status)), util.LazyFiberContext(p.FCtx))

	html := config.DefaultNotFoundHTML
	if serverData.NotFoundHTML != nil {
		html = serverData.NotFoundHTML.Must()
	}

	return p.FCtx.SendString(html)
}

func (h *ServerHandler) findIndexFile(fullPath string, indexFiles []*reader.String) string {
	if len(indexFiles) == 0 {
		indexFiles = []*reader.String{
			reader.NewStr("index.html"),
			reader.NewStr("index.htm"),
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

type ReloadHandler struct{}

func NewReload() Handler {
	return &ReloadHandler{}
}

func (h *ReloadHandler) ProcessAction(p *ProcessData) ExecutionResult {
	_, ok := p.Payload.(*config.ActionReloadData)
	if !ok {
		return ExecutionResult{Err: fmt.Errorf("invalid action data type for reload action: expected *config.ActionReloadData, got %T", p.Payload)}
	}

	if p.Reloader == nil {
		return ExecutionResult{Err: errors.New("runtime reloader is unavailable")}
	}

	if err := p.Reloader.Reload(); err != nil {
		p.CStx.Logger.Error("Action >> Reload Config Failed", zap.Error(err), util.LazyFiberContext(p.FCtx))
		p.FCtx.Status(fiber.StatusInternalServerError)

		return ExecutionResult{Err: p.FCtx.SendString(fmt.Sprintf("Config reload failed: %v", err))}
	}

	p.CStx.Logger.Info("Action >> Reload Config Success", util.LazyFiberContext(p.FCtx))
	p.FCtx.Status(fiber.StatusOK)

	return ExecutionResult{Err: p.FCtx.SendString("Config reloaded successfully")}
}

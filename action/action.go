package action

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/config/loader"
	"github.com/Sn0wo2/CatSync/config/reader"
	"github.com/Sn0wo2/CatSync/internal/util"
	"github.com/Sn0wo2/go-common/helper"
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
	stringData, ok := p.PayLoad.(*config.ActionStringData)
	if !ok {
		return fmt.Errorf("invalid action data type for string action: expected *ActionStringData, got %T", p.PayLoad)
	}

	body := reader.Must(stringData.Content)

	p.Ctx.GetLogger().Info("Action >> Serve String", zap.String("ctx", util.FiberContextString(p.C)))

	return p.C.SendString(body)
}

type FileHandler struct{}

func NewFile() Handler {
	return &FileHandler{}
}

func (h *FileHandler) ProcessAction(p *ProcessData) error {
	fileData, ok := p.PayLoad.(*config.ActionFileData)
	if !ok {
		return fmt.Errorf("invalid action data type for file action: expected *config.ActionFileData, got %T", p.PayLoad)
	}

	pathStr := reader.Must(fileData.Path)

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

	p.Ctx.GetLogger().Info("Action >> Serve File", zap.String("path", pathStr), zap.String("ctx", util.FiberContextString(p.C)))

	return p.C.Send(fileBytes)
}

type ServerHandler struct{}

func NewServer() Handler {
	return &ServerHandler{}
}

const defaultNotFoundHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>404 - Not Found | CatSync</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
            margin: 0;
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
            color: #eee;
        }
        .container {
            text-align: center;
            padding: 2rem;
        }
        h1 {
            font-size: 4rem;
            margin: 0;
            color: #ff6b6b;
        }
        p {
            font-size: 1.2rem;
            margin: 1rem 0 2rem;
            color: #aaa;
        }
        .btn {
            display: inline-flex;
            align-items: center;
            gap: 8px;
            padding: 12px 24px;
            background: #4a90d9;
            color: white;
            text-decoration: none;
            border-radius: 6px;
            font-weight: 500;
            transition: background 0.2s;
        }
        .btn:hover {
            background: #357abd;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>404</h1>
        <p>CatSync: Sync the "cat" config backend server</p>
        <a href="https://github.com/Sn0wo2/CatSync" target="_blank" class="btn">
            <svg height="20" viewBox="0 0 16 16" width="20" fill="currentColor">
                <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"/>
            </svg>
            View on GitHub
        </a>
    </div>
</body>
</html>`

func (h *ServerHandler) ProcessAction(p *ProcessData) error {
	serverData, ok := p.PayLoad.(*config.ActionServerData)
	if !ok {
		return fmt.Errorf("invalid action data type for server action: expected *config.ActionServerData, got %T", p.PayLoad)
	}

	dirStr := reader.Must(serverData.Directory)
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
			p.C.Set(fiber.HeaderContentType, "text/html; charset=utf-8")
			p.C.Status(fiber.StatusNotFound)
			p.Ctx.GetLogger().Info("Action >> Serve Server", zap.String("dir", dirStr), zap.String("file", reqPath), zap.String("status", "404"), zap.String("ctx", util.FiberContextString(p.C)))

			notFoundHTML := defaultNotFoundHTML
			if serverData.NotFoundHTML != nil {
				notFoundHTML = reader.Must(serverData.NotFoundHTML)
			}

			return p.C.SendString(notFoundHTML)
		}

		return fmt.Errorf("failed to access file: %w", err)
	}

	if fileInfo.IsDir() {
		indexFiles := serverData.IndexFiles
		if len(indexFiles) == 0 {
			indexFiles = []*reader.String{
				reader.Str("index.html"),
				reader.Str("index.htm"),
			}
		}

		var foundPath string

		for _, sf := range indexFiles {
			if sf == nil {
				continue
			}

			indexPath := filepath.Join(fullPath, reader.Must(sf))
			if _, err := os.Stat(indexPath); err == nil {
				foundPath = indexPath

				break
			}
		}

		if foundPath != "" {
			fullPath = foundPath
		} else {
			p.C.Set(fiber.HeaderContentType, "text/html; charset=utf-8")
			p.C.Status(fiber.StatusForbidden)
			p.Ctx.GetLogger().Info("Action >> Serve Server", zap.String("dir", dirStr), zap.String("file", reqPath), zap.String("status", "403"), zap.String("ctx", util.FiberContextString(p.C)))

			notFoundHTML := defaultNotFoundHTML
			if serverData.NotFoundHTML != nil {
				notFoundHTML = reader.Must(serverData.NotFoundHTML)
			}

			return p.C.SendString(notFoundHTML)
		}
	}

	p.Ctx.GetLogger().Info("Action >> Serve Server", zap.String("dir", dirStr), zap.String("file", reqPath), zap.String("ctx", util.FiberContextString(p.C)))

	return p.C.SendFile(fullPath)
}

type ReloadHandler struct{}

func NewReload() Handler {
	return &ReloadHandler{}
}

func (h *ReloadHandler) ProcessAction(p *ProcessData) error {
	_, ok := p.PayLoad.(*config.ActionReloadData)
	if !ok {
		return fmt.Errorf("invalid action data type for reload action: expected *config.ActionReloadData, got %T", p.PayLoad)
	}

	cfg := config.GetCurrentConfig()
	if cfg == nil {
		return errors.New("no config loaded")
	}

	err := cfg.Reload(loader.NewYAMLLoader(), loader.NewJSONLoader())
	if err != nil {
		p.Ctx.GetLogger().Error("Action >> Reload Config Failed", zap.Error(err), zap.String("ctx", util.FiberContextString(p.C)))
		p.C.Status(fiber.StatusInternalServerError)

		return p.C.SendString(fmt.Sprintf("Config reload failed: %v", err))
	}

	p.Ctx.GetLogger().Info("Action >> Reload Config Success", zap.String("ctx", util.FiberContextString(p.C)))
	p.C.Status(fiber.StatusOK)

	return p.C.SendString("Config reloaded successfully")
}

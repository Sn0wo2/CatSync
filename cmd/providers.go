package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sn0wo2/CatSync/action"
	"github.com/Sn0wo2/CatSync/action/execute"
	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/config/loader"
	"github.com/Sn0wo2/CatSync/framework"
	"github.com/Sn0wo2/CatSync/internal/appctx"
	"github.com/Sn0wo2/CatSync/log"
	"github.com/Sn0wo2/CatSync/router"
	"github.com/Sn0wo2/CatSync/runtime"
	"go.uber.org/zap"
)

func NewConfig() (*config.LoadResult, error) {
	result, err := config.Load(loader.NewYAMLLoader(), loader.NewJSONLoader())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			result.Config = config.GetDefaultConfig()
			if err := result.Config.Validate(); err != nil {
				return nil, err
			}

			savePath := result.Path
			switch strings.ToLower(filepath.Ext(savePath)) {
			case ".json":
				err = loader.NewJSONLoader().Save(result.Config, savePath)
			default:
				err = loader.NewYAMLLoader().Save(result.Config, savePath)
			}

			if err != nil {
				return nil, err
			}

			log.Writef("Using default config, saved to: %s", savePath)

			return result, nil
		}

		return nil, err
	}

	return result, nil
}

func NewLogger(result *config.LoadResult) *zap.Logger {
	logger := log.NewLog(result.Config.Log.Dir.Must(), result.Config.Log.Level.Must(), result.Config.Log.FileFormat.Must())
	log.Flush(logger)

	return logger
}

func NewRuntime(result *config.LoadResult, logger *zap.Logger) (*runtime.Manager, error) {
	result.Config.LogWarnings(logger)

	return runtime.NewManager(
		result,
		logger,
		[]config.Loader{loader.NewYAMLLoader(), loader.NewJSONLoader()},
		execute.Builders{
			Global:  action.BuildGlobalModifiers,
			Action:  action.BuildActionModifiers,
			Payload: action.BuildPayloadModifiers,
		},
	)
}

func NewParams(manager *runtime.Manager, logger *zap.Logger) *appctx.Ctx {
	p := appctx.New()
	p.ConfigSource = manager
	p.Logger = logger

	return p
}

func NewFramework(p *appctx.Ctx, manager *runtime.Manager) *framework.Framework {
	fw := framework.NewFiber(p)
	p.FW = fw
	router.Init(p, manager)

	return fw
}

package main

import (
	"errors"
	"os"

	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/framework"
	"github.com/Sn0wo2/CatSync/internal/cstx"
	"github.com/Sn0wo2/CatSync/log"
	"github.com/Sn0wo2/CatSync/router"
	"github.com/Sn0wo2/CatSync/runtime"
	"github.com/Sn0wo2/caelum"
)

func NewConfig() (*config.LoadResult, error) {
	cfg, path, err := config.LoadConfig("")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			cfg = config.GetDefaultConfig()
			if err := cfg.Validate(); err != nil {
				return nil, err
			}

			savePath := path

			err = config.SaveConfig(cfg, savePath)
			if err != nil {
				return nil, err
			}

			log.Writef("Using default config, saved to: %s", savePath)

			return &config.LoadResult{Config: cfg, Path: savePath}, nil
		}

		return nil, err
	}

	return &config.LoadResult{Config: cfg, Path: path}, nil
}

func NewLogger(result *config.LoadResult) *caelum.Logger {
	logger := log.NewLog(result.Config.Log.Dir.Must(), result.Config.Log.Level.Must(), result.Config.Log.FileFormat.Must())
	log.Flush(logger.Logger)

	return logger
}

func NewRuntime(result *config.LoadResult, logger *caelum.Logger) (*runtime.Manager, error) {
	result.Config.LogWarnings(logger.Logger)

	return runtime.NewManager(result, logger.Logger)
}

func NewParams(manager *runtime.Manager, logger *caelum.Logger) *cstx.Ctx {
	p := cstx.New()
	p.ConfigSource = manager
	p.Logger = logger.Logger

	return p
}

func NewFramework(p *cstx.Ctx, manager *runtime.Manager) *framework.FB {
	fw := framework.NewFiber(p)
	p.FW = fw
	router.Init(p, manager)

	return fw
}

package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/config/loader"
	"github.com/Sn0wo2/CatSync/framework"
	"github.com/Sn0wo2/CatSync/log"
	"github.com/Sn0wo2/CatSync/params"
	"github.com/Sn0wo2/CatSync/router"
	"go.uber.org/zap"
)

func NewConfig() (*config.Config, error) {
	cfg, err := config.New(loader.NewYAMLLoader(), loader.NewJSONLoader())
	if err != nil {
		// If CONFIG_PATH/DEBUG_CONFIG_PATH points to a missing file,
		// config.New returns a wrapped os.ErrNotExist (not ErrConfigNotFound).
		// Treat both as "config not found" and create default config at config.Path.
		if errors.Is(err, config.ErrConfigNotFound) || errors.Is(err, os.ErrNotExist) {
			cfg = config.GetDefaultConfig()

			savePath := config.Path
			switch strings.ToLower(filepath.Ext(savePath)) {
			case ".json":
				err = loader.NewJSONLoader().Save(cfg, savePath)
			default:
				err = loader.NewYAMLLoader().Save(cfg, savePath)
			}

			if err != nil {
				return nil, err
			}

			log.Write("Using default config, saved to: %s", savePath)
			config.SetCurrentConfig(cfg)
			return cfg, nil
		}

		return nil, err
	}

	config.SetCurrentConfig(cfg)
	return cfg, nil
}

func NewLogger(cfg *config.Config) *zap.Logger {
	logger := log.NewLog(sval(cfg.Log.Dir), sval(cfg.Log.Level), sval(cfg.Log.FileFormat))
	log.Flush(logger)
	return logger
}

func NewParams(cfg *config.Config, logger *zap.Logger) *params.Ctx {
	if err := cfg.Check(logger); err != nil {
		logger.Fatal("Config check failed", zap.Error(err))
	}

	p := params.New()
	p.SetConfig(cfg)
	p.SetLogger(logger)

	return p
}

func NewFramework(p *params.Ctx) *framework.Framework {
	fw := framework.NewFiber(p)
	p.SetFramework(fw)
	router.Init(p)

	return fw
}

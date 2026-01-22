package main

import (
	"errors"

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
		if errors.Is(err, config.ErrConfigNotFound) {
			cfg = config.GetDefaultConfig()
			if err := loader.NewYAMLLoader().Save(cfg, config.Path); err != nil {
				return nil, err
			}
			return cfg, nil
		}
		return nil, err
	}
	return cfg, nil
}

func NewLogger(cfg *config.Config) *zap.Logger {
	return log.NewLog(cfg.Log.Dir, cfg.Log.Level, cfg.Log.FileFormat)
}

func NewParams(cfg *config.Config, logger *zap.Logger) *params.Ctx {
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

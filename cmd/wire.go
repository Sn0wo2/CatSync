//go:build wireinject
// +build wireinject

package main

import (
	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/framework"
	"github.com/Sn0wo2/CatSync/params"
	"github.com/google/wire"
	"go.uber.org/zap"
)

func InitializeApp() (*appContext, error) {
	wire.Build(
		NewConfig,
		NewLogger,
		NewParams,
		NewFramework,
		wire.Struct(new(appContext), "*"),
	)
	return nil, nil
}

type appContext struct {
	Cfg    *config.Config
	Logger *zap.Logger
	Params *params.Ctx
	App    *framework.Framework
}

//go:build wireinject
// +build wireinject

package main

import (
	"github.com/Sn0wo2/CatSync/framework"
	"github.com/Sn0wo2/CatSync/runtime"
	"github.com/google/wire"
	"go.uber.org/zap"
)

func InitializeCatSync() (*catSync, error) {
	wire.Build(
		NewConfig,
		NewLogger,
		NewRuntime,
		NewParams,
		NewFramework,
		wire.Struct(new(catSync), "*"),
	)
	return nil, nil
}

type catSync struct {
	Logger  *zap.Logger
	Runtime *runtime.Manager
	Server  *framework.Framework
}

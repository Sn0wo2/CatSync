package cstx

import (
	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/framework"
	"go.uber.org/zap"
)

type Ctx struct {
	ConfigSource ConfigSource
	Logger       *zap.Logger
	FW           *framework.FB
}

type ConfigSource interface {
	CurrentConfig() *config.Config
}

func New() *Ctx {
	return &Ctx{}
}

func (c *Ctx) GetConfig() *config.Config {
	if c.ConfigSource == nil {
		return nil
	}

	return c.ConfigSource.CurrentConfig()
}

func (c *Ctx) GetLogger() *zap.Logger {
	return c.Logger
}

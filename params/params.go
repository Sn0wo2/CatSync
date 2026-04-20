package params

import (
	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/framework"
	"go.uber.org/zap"
)

type Ctx struct {
	cfg    *config.Config
	logger *zap.Logger
	fw     *framework.Framework
}

func New() *Ctx {
	return &Ctx{}
}

func (c *Ctx) GetConfig() *config.Config {
	return c.cfg
}

func (c *Ctx) SetConfig(cfg *config.Config) *Ctx {
	c.cfg = cfg

	return c
}

func (c *Ctx) GetLogger() *zap.Logger {
	return c.logger
}

func (c *Ctx) SetLogger(logger *zap.Logger) *Ctx {
	c.logger = logger

	return c
}

func (c *Ctx) GetFramework() *framework.Framework {
	return c.fw
}

func (c *Ctx) SetFramework(fw *framework.Framework) *Ctx {
	c.fw = fw

	return c
}

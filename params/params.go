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

	extra ExtraStore
}

func New() *Ctx {
	return &Ctx{}
}

func (c *Ctx) GetConfig() *config.Config {
	if c == nil {
		return nil
	}

	return c.cfg
}

func (c *Ctx) SetConfig(cfg *config.Config) *Ctx {
	if c == nil {
		c = &Ctx{}
	}

	c.cfg = cfg

	return c
}

func (c *Ctx) GetLogger() *zap.Logger {
	if c == nil {
		return nil
	}

	return c.logger
}

func (c *Ctx) SetLogger(logger *zap.Logger) *Ctx {
	if c == nil {
		c = &Ctx{}
	}

	c.logger = logger

	return c
}

func (c *Ctx) GetFramework() *framework.Framework {
	if c == nil {
		return nil
	}

	return c.fw
}

func (c *Ctx) SetFramework(fw *framework.Framework) *Ctx {
	if c == nil {
		c = &Ctx{}
	}

	c.fw = fw

	return c
}

func (c *Ctx) Extra() *ExtraStore {
	if c == nil {
		return nil
	}

	return &c.extra
}

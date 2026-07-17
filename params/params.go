package params

import (
	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/framework"
	"go.uber.org/zap"
)

type Ctx struct {
	configSource ConfigSource
	logger       *zap.Logger
	fw           *framework.Framework
}

type ConfigSource interface {
	CurrentConfig() *config.Config
}

func New() *Ctx {
	return &Ctx{}
}

func (c *Ctx) GetConfig() *config.Config {
	if c.configSource == nil {
		return nil
	}

	return c.configSource.CurrentConfig()
}

func (c *Ctx) SetConfigSource(source ConfigSource) *Ctx {
	c.configSource = source

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

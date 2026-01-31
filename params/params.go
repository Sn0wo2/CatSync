package params

import (
	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/framework"
	"go.uber.org/zap"
)

type Ctx struct {
	vals map[any]any
}

func New() *Ctx {
	return &Ctx{vals: map[any]any{}}
}

func Set[K comparable, V any](c *Ctx, key K, value V) *Ctx {
	if c == nil {
		c = &Ctx{}
	}

	if c.vals == nil {
		c.vals = map[any]any{}
	}

	c.vals[key] = value

	return c
}

func Get[K comparable, V any](c *Ctx, key K) (V, bool) {
	var zero V
	if c == nil {
		return zero, false
	}

	v, ok := c.vals[key]
	if !ok {
		return zero, false
	}

	vv, ok := v.(V)
	if !ok {
		return zero, false
	}

	return vv, true
}

func (c *Ctx) GetConfig() *config.Config {
	cfg, _ := Get[Config, *config.Config](c, Config{})

	return cfg
}

func (c *Ctx) SetConfig(cfg *config.Config) *Ctx {
	return Set[Config, *config.Config](c, Config{}, cfg)
}

func (c *Ctx) GetLogger() *zap.Logger {
	logger, _ := Get[Logger, *zap.Logger](c, Logger{})

	return logger
}

func (c *Ctx) SetLogger(logger *zap.Logger) *Ctx {
	return Set[Logger, *zap.Logger](c, Logger{}, logger)
}

func (c *Ctx) GetFramework() *framework.Framework {
	fw, _ := Get[Framework, *framework.Framework](c, Framework{})

	return fw
}

func (c *Ctx) SetFramework(fw *framework.Framework) *Ctx {
	return Set[Framework, *framework.Framework](c, Framework{}, fw)
}

type (
	Config    struct{}
	Logger    struct{}
	Framework struct{}
)

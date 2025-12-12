package params

import (
	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/framework"
	"go.uber.org/zap"
)

type Value[T any] struct{ v T }

func (v Value[T]) Get() T { return v.v }

type Ctx struct {
	vals map[any]any
}

func New() *Ctx {
	return &Ctx{vals: make(map[any]any)}
}

func Set[F any, T any](c *Ctx, field F, val T) *Ctx {
	if c.vals == nil {
		c.vals = make(map[any]any)
	}

	c.vals[field] = Value[T]{v: val}

	return c
}

func Get[F any, T any](c *Ctx, field F) (T, bool) {
	if c.vals == nil {
		var zero T

		return zero, false
	}

	raw, ok := c.vals[field]
	if !ok {
		var zero T

		return zero, false
	}

	v, ok := raw.(Value[T])
	if !ok {
		var zero T

		return zero, false
	}

	return v.Get(), true
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
